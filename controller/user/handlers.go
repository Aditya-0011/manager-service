package user

import (
	"context"
	"log/slog"
	"manager/db"
	"manager/internal/faults"
	"manager/internal/timeout"
	"time"

	"github.com/Aditya-0011/common/contracts/go/manager"
	"github.com/jackc/pgx/v5"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type (
	userServer struct {
		manager.UnimplementedUserServiceServer
		postgres *db.PostgresParams
	}

	UserServerParams struct {
		Postgres *db.PostgresParams
	}
)

func NewUserServer(params UserServerParams) *userServer {
	return &userServer{
		postgres: params.Postgres,
	}
}

func (us *userServer) GetUserDetails(c context.Context, req *manager.SimpleRequest) (*manager.GetUserDetailsResponse, error) {
	ctx, cancel := timeout.WithDeadline(c, timeout.Duration)
	defer cancel()

	var (
		about      string
		coverImage string
		updatedAt  time.Time
	)

	query := `select about, coverImage, updatedAt from portfolio.user where id = $1`

	err := us.postgres.Pool.QueryRow(ctx, query, req.GetUserId()).Scan(&about, &coverImage, &updatedAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			return &manager.GetUserDetailsResponse{}, nil
		}
		slog.LogAttrs(ctx, slog.LevelError, "error querying database",
			slog.Int64("userId", int64(req.GetUserId())),
			slog.String("error", err.Error()),
		)
		return nil, faults.ErrInternal
	}

	return &manager.GetUserDetailsResponse{
		About:      about,
		CoverImage: coverImage,
		UpdatedAt:  timestamppb.New(updatedAt),
	}, nil
}

func (us *userServer) EditUserDetails(c context.Context, req *manager.EditUserDetailsRequest) (*manager.SimpleResponse, error) {
	ctx, cancel := timeout.WithDeadline(c, timeout.Duration)
	defer cancel()

	var outcode int16

	query := `select outcode from portfolio.edit_user($1, $2, $3)`

	err := us.postgres.Pool.QueryRow(ctx, query, req.GetUserId(), req.GetAbout(), req.GetCoverImage()).Scan(&outcode)
	if err != nil {
		slog.LogAttrs(ctx, slog.LevelError, "error creating user",
			slog.Int64("userId", int64(req.GetUserId())),
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
