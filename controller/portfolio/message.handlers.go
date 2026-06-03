package portfolio

import (
	"context"
	"log/slog"
	"manager/utils"

	manager "github.com/Aditya-0011/common/contracts/go/manager"
	"github.com/jackc/pgx/v5"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (ps *portfolioServer) GetMessages(c context.Context, req *manager.SimpleRequest) (*manager.GetMessagesResponse, error) {
	ctx, cancel := context.WithTimeout(c, utils.TimeoutDuration)
	defer cancel()

	query := `select id, name, email, messages from portfolio.message where userId = @userId`

	queryParams := pgx.NamedArgs{
		"userId": req.GetUserId(),
	}

	rows, err := ps.postgres.Pool.Query(ctx, query, queryParams)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, status.Errorf(codes.NotFound, "Messages not found")
		}
		slog.Error("error querying database", "userId", req.GetUserId(), "error", err)
		return nil, status.Errorf(codes.Internal, "Internal server error")
	}
	defer rows.Close()

	var resMessages []*manager.Message

	for rows.Next() {
		var (
			id       string
			name     string
			email    string
			messages []string
		)
		err := rows.Scan(&id, &name, &email, &messages)
		if err != nil {
			slog.Error("error scanning rows", "userId", req.GetUserId(), "error", err)
			return nil, status.Errorf(codes.Internal, "Internal server error")
		}

		resMessages = append(resMessages, &manager.Message{
			Id:      id,
			Name:    name,
			Email:   email,
			Message: messages,
		})
	}

	if err := rows.Err(); err != nil {
		slog.Error("error iterating rows", "userId", req.GetUserId(), "error", err)
		return nil, status.Errorf(codes.Internal, "Internal server error")
	}

	if len(resMessages) == 0 {
		return nil, status.Errorf(codes.NotFound, "Messages not found")
	}

	return &manager.GetMessagesResponse{
		Messages: resMessages,
	}, nil
}

func (ps *portfolioServer) AddMessage(c context.Context, req *manager.AddMessageRequest) (*manager.SimpleResponse, error) {
	ctx, cancel := context.WithTimeout(c, utils.TimeoutDuration)
	defer cancel()

	query := `select outcode from portfolio.add_message(@userId, @name, @email, @message)`

	queryParams := pgx.NamedArgs{
		"userId":  req.GetUserId(),
		"name":    req.GetName(),
		"email":   req.GetEmail(),
		"message": req.GetMessage(),
	}

	var outcode int8

	err := ps.postgres.Pool.QueryRow(ctx, query, queryParams).Scan(&outcode)
	if err != nil {
		slog.Error("error querying database", "userId", req.GetUserId(), "error", err)
		return nil, status.Errorf(codes.Internal, "Internal server error")
	}

	switch outcode {
	case -1:
		return &manager.SimpleResponse{
			Message: "Message added successfully",
		}, nil
	case 0:
		return nil, status.Errorf(codes.Aborted, "You have reached the limit of messages")
	default:
		return nil, status.Errorf(codes.Internal, "Internal server error")
	}

}

func (ps *portfolioServer) DeleteMessage(c context.Context, req *manager.DeleteMessageRequest) (*manager.SimpleResponse, error) {
	ctx, cancel := context.WithTimeout(c, utils.TimeoutDuration)
	defer cancel()

	query := `select outcode from portfolio.delete_messages(@id, @userId)`

	queryParams := pgx.NamedArgs{
		"id":     req.GetId(),
		"userId": req.GetUserId(),
	}

	var outcode int8
	err := ps.postgres.Pool.QueryRow(ctx, query, queryParams).Scan(&outcode)
	if err != nil {
		slog.Error("error querying database", "messageId", req.GetId(), "userId", req.GetUserId(), "error", err)
		return nil, status.Errorf(codes.Internal, "Internal server error")
	}

	switch outcode {
	case -1:
		return &manager.SimpleResponse{
			Message: "Message deleted successfully",
		}, nil
	case 1:
		return nil, status.Errorf(codes.NotFound, "Message not found")
	default:
		return nil, status.Errorf(codes.Internal, "Internal server error")
	}
}
