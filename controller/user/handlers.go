package user

import (
	"context"
	"errors"
	"log/slog"
	"manager/internal/chatmodels"
	"manager/internal/faults"
	"time"

	"github.com/Aditya-0011/common/contracts/go/manager"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type userServer struct {
	manager.UnimplementedUserServiceServer
	pool *pgxpool.Pool
}

func NewServer(pool *pgxpool.Pool) manager.UserServiceServer {
	return &userServer{pool: pool}
}

func (us *userServer) GetUserDetails(ctx context.Context, req *manager.SimpleRequest) (*manager.GetUserDetailsResponse, error) {
	var (
		about      string
		coverImage string
		chatModel  string
		updatedAt  time.Time
	)

	query := `select about, coverImage, chatModel, updatedAt from portfolio.user where id = $1`

	err := us.pool.QueryRow(ctx, query, req.GetUserId()).Scan(&about, &coverImage, &chatModel, &updatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return &manager.GetUserDetailsResponse{}, nil
		}
		slog.LogAttrs(ctx, slog.LevelError, "error querying database",
			slog.Int("userId", int(req.GetUserId())),
			slog.String("error", err.Error()),
		)
		return nil, faults.ErrInternal
	}

	return &manager.GetUserDetailsResponse{
		About:      about,
		CoverImage: coverImage,
		ChatModel:  chatModel,
		UpdatedAt:  timestamppb.New(updatedAt),
	}, nil
}

func (us *userServer) EditUserDetails(ctx context.Context, req *manager.EditUserDetailsRequest) (*manager.SimpleResponse, error) {
	if err := us.checkChatModel(ctx, req); err != nil {
		return nil, err
	}

	var outcode int16

	query := `select outcode from portfolio.edit_user($1, $2, $3, $4)`

	err := us.pool.QueryRow(ctx, query, req.GetUserId(), req.GetAbout(), req.GetCoverImage(), req.GetChatModel()).Scan(&outcode)
	if err != nil {
		slog.LogAttrs(ctx, slog.LevelError, "error updating user",
			slog.Int("userId", int(req.GetUserId())),
			slog.String("error", err.Error()),
		)
		return nil, faults.ErrInternalUpdate
	}

	switch outcode {
	case -1:
		return &manager.SimpleResponse{Message: "User updated successfully"}, nil
	default:
		return nil, faults.ErrInternalUpdate
	}
}

func (us *userServer) GetChatModels(ctx context.Context, req *manager.SimpleRequest) (*manager.GetChatModelsResponse, error) {
	models, err := chatmodels.ListAvailable(ctx)
	if err != nil {
		slog.LogAttrs(ctx, slog.LevelError, "error listing chat models",
			slog.Int("userId", int(req.GetUserId())),
			slog.String("error", err.Error()),
		)
		return nil, faults.ErrChatModelsUnavailable
	}

	res := make([]*manager.ChatModel, 0, len(models))
	for _, model := range models {
		res = append(res, &manager.ChatModel{Id: model.ID, Name: model.Name})
	}
	return &manager.GetChatModelsResponse{Models: res}, nil
}

func (us *userServer) checkChatModel(ctx context.Context, req *manager.EditUserDetailsRequest) error {
	model := req.GetChatModel()
	if model == "" {
		return nil
	}

	var current string
	err := us.pool.QueryRow(ctx, `select chatModel from portfolio.user where id = $1`, req.GetUserId()).Scan(&current)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		slog.LogAttrs(ctx, slog.LevelError, "error querying database",
			slog.Int("userId", int(req.GetUserId())),
			slog.String("error", err.Error()),
		)
		return faults.ErrInternalUpdate
	}
	if model == current {
		return nil
	}

	models, err := chatmodels.ListAvailable(ctx)
	if err != nil {
		slog.LogAttrs(ctx, slog.LevelError, "error listing chat models",
			slog.Int("userId", int(req.GetUserId())),
			slog.String("error", err.Error()),
		)
		return faults.ErrChatModelsUnavailable
	}
	for _, available := range models {
		if available.ID == model {
			return nil
		}
	}
	return faults.ErrChatModelNotAllowed
}
