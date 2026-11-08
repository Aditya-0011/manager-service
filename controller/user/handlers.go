package user

import (
	"context"
	"errors"
	"log/slog"
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
		updatedAt  time.Time
	)

	query := `select about, coverImage, updatedAt from portfolio.user where id = $1`

	err := us.pool.QueryRow(ctx, query, req.GetUserId()).Scan(&about, &coverImage, &updatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
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

func (us *userServer) EditUserDetails(ctx context.Context, req *manager.EditUserDetailsRequest) (*manager.SimpleResponse, error) {
	var outcode int16

	query := `select outcode from portfolio.edit_user($1, $2, $3)`

	err := us.pool.QueryRow(ctx, query, req.GetUserId(), req.GetAbout(), req.GetCoverImage()).Scan(&outcode)
	if err != nil {
		slog.LogAttrs(ctx, slog.LevelError, "error updating user",
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
