package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"github.com/Shelffy/shelffy/internal/app/api"
	"github.com/Shelffy/shelffy/internal/config"
	_ "github.com/aws/aws-sdk-go-v2/aws"
	_ "github.com/aws/aws-sdk-go-v2/config"
	"github.com/go-chi/docgen"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

var (
	flagRoutes     = flag.Bool("routes", false, "generate documentation for routes")
	flagHelp       = flag.Bool("help", false, "prints this message")
	flagPort       = flag.String("port", "", "port to listen")
	flagConfigPath = flag.String("config", "", "path to config file")
)

func main() {
	flag.Parse()
	if *flagHelp {
		flag.PrintDefaults()
		return
	}
	logger := setupLogger()
	cfg := config.DefaultConfig
	if *flagConfigPath != "" {
		cfg = config.MustParse(*flagConfigPath)
	}
	if *flagPort != "" {
		cfg.Server.Port = *flagPort
	}
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()
	app, err := api.New(ctx, cfg, logger)
	if err != nil {
		logger.Error("failed to create app", "error", err)
		os.Exit(1)
	}
	if *flagRoutes {
		doc := docgen.JSONRoutesDoc(app.Router())
		fmt.Println(doc)
		return
	}
	go func() {
		if err := app.Run(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Error("failed to run app", "error", err)
			cancel()
		}
	}()
	<-ctx.Done()
	app.Stop(1 * time.Second)
}

func setupLogger() *slog.Logger {
	handler := slog.NewTextHandler(os.Stdout, nil)
	return slog.New(handler)
}
