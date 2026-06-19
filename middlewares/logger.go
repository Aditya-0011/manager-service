package middlewares

import (
	"context"
	"log/slog"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func LoggingInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		start := time.Now()

		resp, err := handler(ctx, req)

		duration := time.Since(start)

		if err != nil {
			st, _ := status.FromError(err)
			code := st.Code()

			level := slog.LevelError
			if st.Code() < codes.Internal {
				level = slog.LevelWarn
			}

			slog.LogAttrs(ctx, level, "gRPC error",
				slog.String("method", info.FullMethod),
				slog.String("status", code.String()),
				slog.Duration("duration", duration),
				slog.String("error", err.Error()),
			)
		} else {
			slog.LogAttrs(ctx, slog.LevelInfo, "gRPC success",
				slog.String("method", info.FullMethod),
				slog.String("status", "OK"),
				slog.Duration("duration", duration),
			)
		}

		return resp, err
	}
}
