package api

import (
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	s3config "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/plinkplenk/booki/internal/api/gql"
	"github.com/plinkplenk/booki/internal/api/http/routers"
	"github.com/plinkplenk/booki/internal/auth"
	"github.com/plinkplenk/booki/internal/book"
	"github.com/plinkplenk/booki/internal/config"
	"github.com/plinkplenk/booki/internal/user"
	"log/slog"
	"net/http"
	"sync"
	"time"
)

const (
	defaultSessionLifeTime = 24 * time.Hour * 30
)

type App struct {
	ctx    context.Context
	logger *slog.Logger
	router chi.Router
	pool   *pgxpool.Pool
	server *http.Server
}

type repositories struct {
	userRepo user.Repository
	authRepo auth.Repository
	bookRepo book.Repository
}

func newRepositories(conn *pgxpool.Pool) repositories {
	userRepo := user.NewPsqlRepository(conn)
	authRepo := auth.NewPsqlRepository(conn)
	return repositories{
		userRepo: userRepo,
		authRepo: authRepo,
		bookRepo: nil, // WARN: add later
	}
}

type services struct {
	userService user.Service
	authService auth.Service
	//bookService book.Service
}

func newServices(repos repositories, cfg config.Config, logger *slog.Logger) services {
	if cfg.Auth.SessionLifeTime == 0 {
		logger.Warn("session life time is not provided, using default session life time")
		cfg.Auth.SessionLifeTime = defaultSessionLifeTime
	}
	if cfg.Auth.Secret == "" {
		logger.Warn("secret is not provided, using default secret")
		cfg.Auth.Secret = "secret"
	}
	return services{
		userService: user.NewService(
			repos.userRepo,
			cfg.Services.UserServiceTimeout,
			logger.WithGroup("user_service"),
		),
		authService: auth.NewService(
			repos.userRepo,
			repos.authRepo,
			cfg.Services.AuthServiceTimeout,
			cfg.Auth.SessionLifeTime,
			logger.WithGroup("auth_service"),
			cfg.Auth.Secret,
		),
	}
}

func loadS3Config(ctx context.Context, appConfig config.Config) (aws.Config, error) {
	return s3config.LoadDefaultConfig(
		ctx,
		s3config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
			appConfig.S3.AccessKeyID,
			appConfig.S3.AccessSecretKey,
			appConfig.S3.AccountID,
		)),
		s3config.WithRegion("auto"),
	)
}

func New(ctx context.Context, config config.Config, logger *slog.Logger) (App, error) {
	pool, err := pgxpool.New(ctx, config.DB.ConnectionString)
	if err != nil {
		return App{}, err
	}
	s3cfg, err := loadS3Config(ctx, config)
	if err != nil {
		return App{}, err
	}
	_ = s3.NewFromConfig(s3cfg, func(options *s3.Options) { // TODO
		options.BaseEndpoint = aws.String(config.S3.APIEndpoint)
	})
	appServices := newServices(
		newRepositories(pool),
		config,
		logger,
	)
	graphqlHandler := gql.New(gql.Args{
		UserService: appServices.userService,
		AuthService: appServices.authService,
		Logger:      logger,
	})
	router := routers.NewRouter(routers.RouterArgs{
		GQLHandler:  graphqlHandler,
		UserService: appServices.userService,
		AuthService: appServices.authService,
		Logger:      logger,
	})
	return App{
		ctx:    ctx,
		logger: logger,
		pool:   pool,
		router: router,
		server: &http.Server{
			Addr:         config.Server.Port,
			Handler:      router,
			ReadTimeout:  config.Server.ReadTimeout,
			WriteTimeout: config.Server.WriteTimeout,
			IdleTimeout:  config.Server.IdleTimeout,
			ErrorLog: slog.NewLogLogger(
				logger.Handler().WithGroup("http"),
				slog.LevelError,
			),
		},
	}, nil
}

func (a *App) Run() error {
	a.LogRoutes()
	a.logger.Info(fmt.Sprintf("listening port %s", a.server.Addr))
	return a.server.ListenAndServe()
}

func (a *App) Router() chi.Router {
	return a.router
}

func (a *App) LogRoutes() {
	a.logger.Debug("Registered routes:")
	err := chi.Walk(a.router, func(method string, route string, handler http.Handler, middlewares ...func(http.Handler) http.Handler) error {
		a.logger.Debug("Route", "method", method, "path", route)
		return nil
	})
	if err != nil {
		a.logger.Debug("Failed to walk routes", "error", err)
	}
}

func (a *App) Stop(timeout time.Duration) {
	a.logger.Info("stopping the app")
	c, cancel := context.WithTimeout(a.ctx, timeout)
	defer cancel()
	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		if err := a.server.Shutdown(c); err != nil {
			a.logger.Error("failed to shutdown server", "error", err)
		} else {
			a.logger.Info("server stopped")
		}
		wg.Done()
	}()
	go func() {
		ready := make(chan struct{})
		go func() {
			a.pool.Close()
			ready <- struct{}{}
		}()
		select {
		case <-c.Done():
			a.logger.Error("failed to close db connection", "error", c.Err())
		case <-ready:
			a.logger.Info("db connection closed")
		}
		wg.Done()
	}()
	wg.Wait()
}
