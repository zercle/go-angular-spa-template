// HTTP mapping logic for the shared boundary errors.
package errors

import (
	"context"
	"errors"
	"net/http"
)

// HTTPError maps any error to an HTTP status code and a JSON-shaped response
// body. A nil error maps to 200 with a success body.
func HTTPError(err error) (int, map[string]any) {
	if err == nil {
		return http.StatusOK, map[string]any{"status": "ok"}
	}

	app := resolveAppError(err)

	body := map[string]any{
		"error":   app.Code,
		"message": app.Message,
	}

	return app.HTTPStatus, body
}

// resolveAppError converts err into an AppError using, in order:
//  1. direct *AppError match via errors.As,
//  2. a registered domain sentinel via errors.Is,
//  3. the standard context errors (Canceled, DeadlineExceeded),
//  4. the shared ErrInternal as a fallback.
//
// Every successful path returns a clone of the matched AppError so callers can
// never mutate shared sentinels or the AppError they passed in.
func resolveAppError(err error) *AppError {
	var app *AppError
	if errors.As(err, &app) {
		clone := *app
		clone.Cause = err
		return &clone
	}

	if app := sentinelFor(err); app != nil {
		clone := *app
		clone.Cause = err
		return &clone
	}

	if errors.Is(err, context.Canceled) {
		clone := *ErrCanceled
		clone.Cause = err
		return &clone
	}
	if errors.Is(err, context.DeadlineExceeded) {
		clone := *ErrDeadlineExceeded
		clone.Cause = err
		return &clone
	}

	clone := *ErrInternal
	clone.Cause = err
	return &clone
}
