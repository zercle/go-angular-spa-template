// STUB FEATURE — delete internal/features/tasks to start your project.

package di

import (
	"fmt"

	"github.com/samber/do/v2"
	valkeygo "github.com/valkey-io/valkey-go"
	"go.opentelemetry.io/otel/metric"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"

	pb "github.com/zercle/go-angular-spa-template/api/pb/tasks/v1"
	"github.com/zercle/go-angular-spa-template/internal/features/tasks/domain"
	grpchandler "github.com/zercle/go-angular-spa-template/internal/features/tasks/handler/grpc"
	httphandler "github.com/zercle/go-angular-spa-template/internal/features/tasks/handler/http"
	"github.com/zercle/go-angular-spa-template/internal/features/tasks/repository"
	"github.com/zercle/go-angular-spa-template/internal/features/tasks/service"
	sqlcdb "github.com/zercle/go-angular-spa-template/internal/infrastructure/db/sqlc"
	sharederrors "github.com/zercle/go-angular-spa-template/internal/shared/errors"

	"github.com/labstack/echo/v5"
	"google.golang.org/grpc"
)

// Register wires the tasks feature into the composition root.
func Register(c do.Injector) error {
	sharederrors.RegisterSentinel(domain.ErrTaskNotFound, sharederrors.ErrNotFound)
	sharederrors.RegisterSentinel(domain.ErrInvalidTitle, sharederrors.ErrInvalidInput)
	sharederrors.RegisterSentinel(domain.ErrInvalidID, sharederrors.ErrInvalidInput)

	do.Provide(c, func(i do.Injector) (domain.Repository, error) {
		queries := do.MustInvoke[*sqlcdb.Queries](i)
		base := repository.NewRepository(queries)

		// Decorate with a Valkey read-through cache + OTel spans + Prometheus
		// counters so the full stack is exercised on the request path.
		client := do.MustInvoke[valkeygo.Client](i)
		tracer := do.MustInvoke[*sdktrace.TracerProvider](i).Tracer("tasks/repository")
		meter := do.MustInvoke[*sdkmetric.MeterProvider](i).Meter("tasks/repository")

		hits, err := meter.Int64Counter("tasks.cache.hits",
			metric.WithDescription("Task cache hits by operation"))
		if err != nil {
			return nil, fmt.Errorf("create cache hits counter: %w", err)
		}
		misses, err := meter.Int64Counter("tasks.cache.misses",
			metric.WithDescription("Task cache misses by operation"))
		if err != nil {
			return nil, fmt.Errorf("create cache misses counter: %w", err)
		}
		created, err := meter.Int64Counter("tasks.created",
			metric.WithDescription("Total tasks created"))
		if err != nil {
			return nil, fmt.Errorf("create tasks created counter: %w", err)
		}

		return repository.NewCachedRepository(
			base,
			repository.NewValkeyCache(client),
			tracer,
			repository.Metrics{Hits: hits, Misses: misses, Created: created},
		), nil
	})

	do.Provide(c, func(i do.Injector) (domain.Service, error) {
		repo := do.MustInvoke[domain.Repository](i)
		return service.NewService(repo), nil
	})

	do.Provide(c, func(i do.Injector) (*httphandler.Handler, error) {
		svc := do.MustInvoke[domain.Service](i)
		return httphandler.New(svc), nil
	})

	do.Provide(c, func(i do.Injector) (*grpchandler.Server, error) {
		svc := do.MustInvoke[domain.Service](i)
		return grpchandler.NewServer(svc), nil
	})

	h, err := do.Invoke[*httphandler.Handler](c)
	if err != nil {
		return fmt.Errorf("resolve tasks http handler: %w", err)
	}
	e, err := do.Invoke[*echo.Echo](c)
	if err != nil {
		return fmt.Errorf("resolve tasks echo: %w", err)
	}
	g := e.Group("/api/v1")
	h.Register(g)

	gs, err := do.Invoke[*grpc.Server](c)
	if err != nil {
		return fmt.Errorf("resolve tasks grpc server: %w", err)
	}
	grpcHandler, err := do.Invoke[*grpchandler.Server](c)
	if err != nil {
		return fmt.Errorf("resolve tasks grpc handler: %w", err)
	}
	pb.RegisterTaskServiceServer(gs, grpcHandler)

	return nil
}
