package loghandler

import (
	"context"
	"log/slog"
)

type Handler struct {
	handler slog.Handler
	ctxKeys []string
}

func NewHandler(h slog.Handler, ctxKeys ...string) *Handler {
	return &Handler{handler: h, ctxKeys: ctxKeys}
}

func (h *Handler) Enabled(ctx context.Context, level slog.Level) bool {
	return h.handler.Enabled(ctx, level)
}

func (h *Handler) Handle(ctx context.Context, r slog.Record) error {
	for _, key := range h.ctxKeys {
		val := ctx.Value(key)
		r.AddAttrs(slog.Any(key, val))
	}
	return h.handler.Handle(ctx, r)
}

// WithAttrs implements Handler.WithAttrs.
func (h *Handler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return NewHandler(h.handler.WithAttrs(attrs))
}

// WithGroup implements Handler.WithGroup.
func (h *Handler) WithGroup(name string) slog.Handler {
	return NewHandler(h.handler.WithGroup(name))
}

// Handler returns the Handler wrapped by h.
func (h *Handler) Handler() slog.Handler {
	return h.handler
}
