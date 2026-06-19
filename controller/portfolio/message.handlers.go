package portfolio

import (
	"context"
	"log/slog"
	"manager/internal/faults"
	"manager/internal/timeout"

	"github.com/Aditya-0011/common/contracts/go/manager"
)

func (ps *portfolioServer) GetMessages(c context.Context, req *manager.SimpleRequest) (*manager.GetMessagesResponse, error) {
	ctx, cancel := timeout.WithDeadline(c, timeout.Duration)
	defer cancel()

	query := `select id, name, email, messages from portfolio.message where userId = $1`

	rows, err := ps.postgres.Pool.Query(ctx, query, req.GetUserId())
	if err != nil {
		slog.LogAttrs(ctx, slog.LevelError, "error querying database",
			slog.Int64("userId", int64(req.GetUserId())),
			slog.String("error", err.Error()),
		)
		return nil, faults.ErrInternal
	}
	defer rows.Close()

	resMessages := make([]*manager.Message, 0, 16)

	for rows.Next() {
		var (
			id       string
			name     string
			email    string
			messages []string
		)
		err := rows.Scan(&id, &name, &email, &messages)
		if err != nil {
			slog.LogAttrs(ctx, slog.LevelError, "error scanning rows",
				slog.Int64("userId", int64(req.GetUserId())),
				slog.String("error", err.Error()),
			)
			return nil, faults.ErrInternal
		}

		resMessages = append(resMessages, &manager.Message{
			Id:      id,
			Name:    name,
			Email:   email,
			Message: messages,
		})
	}

	if err := rows.Err(); err != nil {
		slog.LogAttrs(ctx, slog.LevelError, "error iterating rows",
			slog.Int64("userId", int64(req.GetUserId())),
			slog.String("error", err.Error()),
		)
		return nil, faults.ErrInternal
	}

	if len(resMessages) == 0 {
		return nil, faults.ErrMessagesNotFound
	}

	return &manager.GetMessagesResponse{
		Messages: resMessages,
	}, nil
}

func (ps *portfolioServer) AddMessage(c context.Context, req *manager.AddMessageRequest) (*manager.SimpleResponse, error) {
	ctx, cancel := timeout.WithDeadline(c, timeout.Duration)
	defer cancel()

	query := `select outcode from portfolio.add_message($1, $2, $3, $4)`

	var outcode int16

	err := ps.postgres.Pool.QueryRow(ctx, query, req.GetUserId(), req.GetName(), req.GetEmail(), req.GetMessage()).Scan(&outcode)
	if err != nil {
		slog.LogAttrs(ctx, slog.LevelError, "error querying database",
			slog.Int64("userId", int64(req.GetUserId())),
			slog.String("error", err.Error()),
		)
		return nil, faults.ErrInternalCreate
	}

	switch outcode {
	case -1:
		return &manager.SimpleResponse{
			Message: "Message added successfully",
		}, nil
	case 0:
		return nil, faults.ErrMessageLimit
	default:
		return nil, faults.ErrInternalCreate
	}

}

func (ps *portfolioServer) DeleteMessage(c context.Context, req *manager.DeleteMessageRequest) (*manager.SimpleResponse, error) {
	ctx, cancel := timeout.WithDeadline(c, timeout.Duration)
	defer cancel()

	query := `select outcode from portfolio.delete_messages($1, $2)`

	var outcode int16
	err := ps.postgres.Pool.QueryRow(ctx, query, req.GetId(), req.GetUserId()).Scan(&outcode)
	if err != nil {
		slog.LogAttrs(ctx, slog.LevelError, "error querying database",
			slog.String("messageId", req.GetId()),
			slog.Int64("userId", int64(req.GetUserId())),
			slog.String("error", err.Error()),
		)
		return nil, faults.ErrInternalDelete
	}

	switch outcode {
	case -1:
		return &manager.SimpleResponse{
			Message: "Message deleted successfully",
		}, nil
	case 1:
		return nil, faults.ErrMessageNotFound
	default:
		return nil, faults.ErrInternalDelete
	}
}
