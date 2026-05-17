package portfolio

import (
	"context"
	"log/slog"
	"manager/utils"

	manager "github.com/Aditya-0011/common/contracts/go/manager"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (ps *portfolioServer) GetMessages(c context.Context, req *manager.SimpleRequest) (*manager.GetMessagesResponse, error) {
	ctx, cancel := context.WithTimeout(c, utils.TimeoutDuration)
	defer cancel()

	query := `select id, name, email, messages from portfolio.message where userId = $1`

	rows, err := ps.postgres.Pool.Query(ctx, query, req.GetUserId())
	if err != nil {
		slog.Error("error querying database", "userId", req.GetUserId(), "error", err)
		return nil, status.Errorf(codes.Internal, "internal server error")
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
			return nil, status.Errorf(codes.Internal, "internal server error")
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
		return nil, status.Errorf(codes.Internal, "internal server error")
	}

	if len(resMessages) == 0 {
		return nil, status.Errorf(codes.NotFound, "messages not found")
	}

	return &manager.GetMessagesResponse{
		Messages: resMessages,
	}, nil
}

func (ps *portfolioServer) AddMessage(c context.Context, req *manager.AddMessageRequest) (*manager.SimpleResponse, error) {
	ctx, cancel := context.WithTimeout(c, utils.TimeoutDuration)
	defer cancel()

	query := `select * from portfolio.add_message($1, $2, $3, $4)`

	var outcode int8
	err := ps.postgres.Pool.QueryRow(ctx, query, req.GetUserId(), req.GetName(), req.GetEmail(), req.GetMessage()).Scan(&outcode)
	if err != nil {
		slog.Error("error querying database", "userId", req.GetUserId(), "error", err)
		return nil, status.Errorf(codes.Internal, "internal server error")
	}

	switch outcode {
	case -1:
		return &manager.SimpleResponse{
			Message: "Message added successfully",
		}, nil
	case 0:
		return nil, status.Errorf(codes.Aborted, "You have reached the limit of messages")
	default:
		return nil, status.Errorf(codes.Internal, "internal server error")
	}

}

func (ps *portfolioServer) DeleteMessage(c context.Context, req *manager.DeleteMessageRequest) (*manager.SimpleResponse, error) {
	ctx, cancel := context.WithTimeout(c, utils.TimeoutDuration)
	defer cancel()

	query := `select * from portfolio.delete_messages($1, $2)`

	var outcode int8
	err := ps.postgres.Pool.QueryRow(ctx, query, req.GetId(), req.GetUserId()).Scan(&outcode)
	if err != nil {
		slog.Error("error scanning rows", "messageId", req.GetId(), "userId", req.GetUserId(), "error", err)
		return nil, status.Errorf(codes.Internal, "internal server error")
	}

	switch outcode {
	case -1:
		return &manager.SimpleResponse{
			Message: "Message deleted successfully",
		}, nil
	case 1:
		return nil, status.Errorf(codes.NotFound, "message not found")
	default:
		return nil, status.Errorf(codes.Internal, "internal server error")
	}
}
