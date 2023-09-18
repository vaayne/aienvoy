package logger

import (
	"context"
	"log/slog"

	"github.com/Vaayne/aienvoy/internal/pkg/config"
	"github.com/Vaayne/aienvoy/internal/pkg/ctxutils"
)

type Handler struct {
	handler slog.Handler
}

func NewHandler(h slog.Handler) *Handler {
	return &Handler{handler: h}
}

func (h *Handler) Enabled(ctx context.Context, level slog.Level) bool {
	return h.handler.Enabled(ctx, level)
}

func (h *Handler) Handle(ctx context.Context, r slog.Record) error {
	r.AddAttrs(
		slog.Attr{
			Key:   config.ContextKeyUserId,
			Value: slog.StringValue(ctxutils.GetUserId(ctx)),
		},
		// slog.Attr{
		// 	Key:   config.ContextKeyApiKey,
		// 	Value: slog.StringValue(ctxutils.GetApiKey(ctx)),
		// },
		slog.Attr{
			Key:   config.ContextKeyRequestId,
			Value: slog.StringValue(ctxutils.GetRequestId(ctx)),
		},
	)
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
