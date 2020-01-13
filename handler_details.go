package emperror

import (
	"context"

	"emperror.dev/errors"
)

// WithDetails returns a new error handler that annotates every error with a set of key-value pairs.
func WithDetails(handler ErrorHandler, details ...interface{}) ErrorHandlerSet {
	handlerSet := ensureErrorHandlerSet(handler)

	if len(details) == 0 {
		return handlerSet
	}

	d, handlerSet := extractHandlerDetails(handlerSet)

	d = append(d, details...)

	if len(d)%2 != 0 {
		d = append(d, nil)
	}

	// Limiting the capacity of the stored keyvals ensures that a new
	// backing array is created if the slice must grow in HandlerWith.
	// Using the extra capacity without copying risks a data race.
	return newWithDetails(handlerSet, d[:len(d):len(d)])
}

// HandlerWithDetails returns a new error handler annotated with key-value pairs.
//
// The created handler will add it's own details to the handled errors.
// Deprecated: use WithDetails instead.
func HandlerWithDetails(handler Handler, details ...interface{}) Handler {
	return WithDetails(handler, details...)
}

// extractHandlerDetails extracts the context and optionally the wrapped handler when it's the same container.
func extractHandlerDetails(handler ErrorHandlerSet) ([]interface{}, ErrorHandlerSet) {
	var d []interface{}

	if c, ok := handler.(*withDetails); ok {
		handler = c.handler
		d = c.details[:] // nolint: gocritic
	}

	return d, handler
}

// withDetails is a Handler implementation returned by HandlerWith or HandlerWithPrefix.
//
// It wraps an error handler and a holds keyvals as the context.
type withDetails struct {
	handler ErrorHandlerSet
	details []interface{}
}

// newWithDetails creates a new handler annotated with a set of key-value pairs.
func newWithDetails(handler ErrorHandlerSet, details []interface{}) ErrorHandlerSet {
	return &withDetails{
		handler: handler,
		details: details,
	}
}

// Handle adds the handler's details to err and delegates the call to the underlying handler.
func (h *withDetails) Handle(err error) {
	// TODO: prepend details so the error takes higher precedence
	err = errors.WithDetails(err, h.details...)

	h.handler.Handle(err)
}

// HandleContext adds the handler's details to err and delegates the call to the underlying handler.
func (h *withDetails) HandleContext(ctx context.Context, err error) {
	// TODO: prepend details so the error takes higher precedence
	err = errors.WithDetails(err, h.details...)

	h.handler.HandleContext(ctx, err)
}
