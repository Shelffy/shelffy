package api

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"sync"
	"time"

	"github.com/Shelffy/shelffy/internal/api/gql"
	"github.com/Shelffy/shelffy/internal/api/http/routers"
	"github.com/Shelffy/shelffy/internal/config"
	repositories2 "github.com/Shelffy/shelffy/internal/repositories"
	services2 "github.com/Shelffy/shelffy/internal/services"
	"github.com/avito-tech/go-transaction-manager/drivers/pgxv5/v2"
	"github.com/avito-tech/go-transaction-manager/trm/v2/manager"
	"github.com/aws/aws-sdk-go-v2/aws"
	s3config "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/kr/pretty"
	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
)

const (
	defaultSessionLifeTime = 24 * time.Hour * 30
)

type App struct {
	ctx            context.Context
	logger         *slog.Logger
	router         chi.Router
	pool           *pgxpool.Pool
	server         *http.Server
	nc             *nats.Conn
	eventProcessor services2.EventsProcessor
}

type appRepositories struct {
	userRepo repositories2.Users
	authRepo repositories2.Session
	bookRepo repositories2.Books
}

func newRepositories(conn *pgxpool.Pool) appRepositories {
	return appRepositories{
		userRepo: repositories2.NewUsersPSQLRepository(conn),
		authRepo: repositories2.NewAuthPSQLRepository(conn),
		bookRepo: repositories2.NewBooksPSQLRepository(conn),
	}
}

type appServices struct {
	userService     services2.Users
	authService     services2.Auth
	bookService     services2.Books
	storage         services2.FileStorage
	eventsProcessor services2.EventsProcessor
}

func newServices(
	repos appRepositories,
	cfg config.Config,
	s3client *s3.Client,
	js jetstream.JetStream,
	txManager *manager.Manager,
	logger *slog.Logger,
) appServices {
	if cfg.Auth.SessionLifeTime == 0 {
		logger.Warn("session life time is not provided, using default session life time")
		cfg.Auth.SessionLifeTime = defaultSessionLifeTime
	}
	if cfg.Auth.Secret == "" {
		logger.Warn("secret is not provided, using default secret")
		cfg.Auth.Secret = "secret"
	}
	storageService := services2.NewS3Storage(cfg.S3.BooksBucket, s3client)
	booksEventsPublisher := services2.NewNATSBooksEventPublisher(js)
	return appServices{
		userService: services2.NewUsers(
			repos.userRepo,
			cfg.Services.UserServiceTimeout,
			logger.WithGroup("user_service"),
		),
		authService: services2.NewAuth(
			repos.userRepo,
			repos.authRepo,
			cfg.Services.AuthServiceTimeout,
			cfg.Auth.SessionLifeTime,
			logger.WithGroup("auth_service"),
			cfg.Auth.Secret,
		),
		storage: storageService,
		bookService: services2.NewBookService(
			repos.bookRepo,
			storageService,
			cfg.Services.BookServiceTimeout,
			booksEventsPublisher,
			txManager,
			logger.WithGroup("book_services"),
		),
		eventsProcessor: services2.NewNATSEventProcessor(
			js,
			storageService,
			logger.WithGroup("events_processor"),
		),
	}
}

func loadS3Config(ctx context.Context, appConfig config.Config) (aws.Config, error) {
	return s3config.LoadDefaultConfig(
		ctx,
		s3config.WithRequestChecksumCalculation(aws.RequestChecksumCalculationWhenRequired),
		s3config.WithResponseChecksumValidation(aws.ResponseChecksumValidationWhenRequired),
		s3config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
			appConfig.S3.AccessKeyID,
			appConfig.S3.AccessSecretKey,
			"",
		)),
		s3config.WithRegion("auto"),
	)
}

func New(ctx context.Context, config config.Config, logger *slog.Logger) (App, error) {
	if config.Debug {
		fmt.Println(pretty.Sprint(config))
	}
	pool, err := pgxpool.New(ctx, config.DB.ConnectionString)
	if err != nil {
		return App{}, err
	}
	s3cfg, err := loadS3Config(ctx, config)
	if err != nil {
		return App{}, err
	}
	s3conn := s3.NewFromConfig(s3cfg, func(options *s3.Options) {
		options.BaseEndpoint = aws.String(config.S3.APIEndpoint)
	})
	nc, err := nats.Connect(config.NATS.URL)
	if err != nil {
		return App{}, err
	}
	js, err := jetstream.New(nc)
	if err != nil {
		return App{}, err
	}
	mngr, err := manager.New(pgxv5.NewDefaultFactory(pool))
	if err != nil {
		return App{}, err
	}
	appServices := newServices(
		newRepositories(pool),
		config,
		s3conn,
		js,
		mngr,
		logger,
	)
	graphqlHandler := gql.New(
		gql.Args{
			UserService: appServices.userService,
			AuthService: appServices.authService,
			BookService: appServices.bookService,
			Logger:      logger,
		},
		config.Debug,
	)
	router := routers.NewRouter(
		routers.RouterArgs{
			GQLHandler:     graphqlHandler,
			UserService:    appServices.userService,
			AuthService:    appServices.authService,
			BooksService:   appServices.bookService,
			StorageService: appServices.storage,
			Logger:         logger,
		},
	)
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
		eventProcessor: appServices.eventsProcessor,
		nc:             nc,
	}, nil
}

func (a *App) Run() error {
	a.LogRoutes()
	a.logger.Info(fmt.Sprintf("listening port %s", a.server.Addr))
	go func() {
		if err := a.eventProcessor.Run(a.ctx); err != nil {
			a.logger.Error("event processor running error", "error", err)
		}
	}()
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
		a.logger.Debug("Failed to walk trough routes", "error", err)
	}
}

func (a *App) Stop(timeout time.Duration) {
	a.logger.Info("stopping the app")
	c, cancel := context.WithTimeout(a.ctx, timeout)
	defer cancel()
	var wg sync.WaitGroup
	wg.Go(func() {
		if err := a.server.Shutdown(c); err != nil {
			a.logger.Error("failed to shutdown server", "error", err)
		} else {
			a.logger.Info("server stopped")
		}
		wg.Done()
	})
	wg.Go(func() {
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
	})
	wg.Go(func() {
		if err := a.nc.Drain(); err != nil {
			a.logger.Error("failed to drain NATS", "error", err)
		} else {
			a.logger.Info("NATS stopped")
		}
		wg.Done()
	})
	wg.Wait()
}
