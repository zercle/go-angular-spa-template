// Command server is the composition root: it loads config, runs migrations,
// wires the clean-architecture layers, and starts the Fiber HTTP server that
// serves both the JSON API and the embedded Angular SPA.
package main

import (
	"context"
	"log/slog"
	"os"

	"github.com/cerzzlive/gofiber-angular-spa/backend/internal/config"
	"github.com/cerzzlive/gofiber-angular-spa/backend/internal/delivery/rest"
	"github.com/cerzzlive/gofiber-angular-spa/backend/internal/repository/postgres"
	"github.com/cerzzlive/gofiber-angular-spa/backend/internal/usecase"
	"github.com/cerzzlive/gofiber-angular-spa/backend/web"
)

func main() {
	log := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	if err := run(log); err != nil {
		log.Error("fatal", "err", err)
		os.Exit(1)
	}
}

func run(log *slog.Logger) error {
	cfg := config.Load()

	if err := postgres.Migrate(cfg.DatabaseURL); err != nil {
		return err
	}
	log.Info("migrations applied")

	ctx := context.Background()
	pool, err := postgres.NewPool(ctx, cfg.DatabaseURL)
	if err != nil {
		return err
	}
	defer pool.Close()

	repo := postgres.NewTaskRepository(pool)
	svc := usecase.NewTaskService(repo)

	spa, err := web.SPA()
	if err != nil {
		return err
	}

	app := rest.NewRouter(svc, spa)
	log.Info("server listening", "port", cfg.Port, "env", cfg.AppEnv)
	return app.Listen(":" + cfg.Port)
}
