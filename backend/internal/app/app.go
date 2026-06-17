// Package app is the reusable composition root. It wires the DI container and
// constructs a runnable server.Application for tests, CLIs, and the main entry
// point.
package app

import (
	"context"
	"errors"
	"fmt"

	"github.com/labstack/echo/v5"
	"github.com/rs/zerolog"
	"github.com/samber/do/v2"

	"github.com/zercle/go-angular-spa-template/internal/config"
	tasksdi "github.com/zercle/go-angular-spa-template/internal/features/tasks/di"
	"github.com/zercle/go-angular-spa-template/internal/infrastructure/db"
	"github.com/zercle/go-angular-spa-template/internal/infrastructure/messaging/valkey"
	"github.com/zercle/go-angular-spa-template/internal/shared/server"
	"github.com/zercle/go-angular-spa-template/internal/shared/telemetry"
	"github.com/zercle/go-angular-spa-template/internal/web"
)

// Version metadata is populated by cmd/server/main.go via these package-level
// variables before Run is called.
var (
	Version   = "dev"
	CommitSHA = "unknown"
	BuildTime = "unknown"
)

// Build wires the DI container in dependency order and returns the
// orchestrated application along with the populated injector.
//
// The sequence is config → telemetry → database → valkey → shared servers →
// tasks feature. On error the partially-wired injector is returned; the
// caller is responsible for calling injector.Shutdown() to release any
// providers that were successfully constructed.
func Build(ctx context.Context, cfg *config.Config) (*server.Application, do.Injector, error) {
	if cfg == nil {
		return nil, nil, fmt.Errorf("config is nil")
	}

	injector := do.New()

	do.ProvideValue(injector, cfg)

	if err := telemetry.Register(ctx, injector); err != nil {
		return nil, injector, fmt.Errorf("register telemetry: %w", err)
	}

	logger, err := do.Invoke[*zerolog.Logger](injector)
	if err != nil {
		return nil, injector, fmt.Errorf("resolve logger: %w", err)
	}
	logger.Info().
		Str("version", Version).
		Str("commit", CommitSHA).
		Str("build_time", BuildTime).
		Str("env", cfg.App.Environment).
		Msg("starting server")

	if err := db.Register(ctx, injector); err != nil {
		return nil, injector, fmt.Errorf("register database: %w", err)
	}

	if err := valkey.Register(ctx, injector); err != nil {
		return nil, injector, fmt.Errorf("register valkey: %w", err)
	}

	if err := server.Register(injector); err != nil {
		return nil, injector, fmt.Errorf("register shared servers: %w", err)
	}

	if err := tasksdi.Register(injector); err != nil {
		return nil, injector, fmt.Errorf("register tasks feature: %w", err)
	}

	// Serve the embedded Angular SPA from the same Echo server (registered last
	// so the API and health/metrics routes take routing priority).
	echoServer, err := do.Invoke[*echo.Echo](injector)
	if err != nil {
		return nil, injector, fmt.Errorf("resolve echo server: %w", err)
	}
	spaFS, err := web.SPA()
	if err != nil {
		return nil, injector, fmt.Errorf("load embedded spa: %w", err)
	}
	if err := web.Register(echoServer, spaFS); err != nil {
		return nil, injector, fmt.Errorf("register spa: %w", err)
	}

	application := server.NewApplication(injector, cfg, logger)
	return application, injector, nil
}

// Run builds the application and runs it until the context is cancelled or a
// server error occurs. It is the simplest entry point for tests and the main
// binary.
func Run(ctx context.Context, cfg *config.Config) error {
	application, injector, err := Build(ctx, cfg)
	if err != nil {
		if injector != nil {
			_ = injector.Shutdown()
		}
		return err
	}

	logger := application.Logger()
	defer func() {
		report := injector.Shutdown()
		if report != nil && !report.Succeed {
			logger.Error().Err(report).Msg("injector shutdown error")
		}
	}()

	if err := application.Run(ctx); err != nil && !errors.Is(err, context.Canceled) {
		return fmt.Errorf("run application: %w", err)
	}
	return nil
}
