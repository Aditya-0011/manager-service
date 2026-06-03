package user

import (
	"context"
	"log/slog"
	"manager/db"
	"manager/utils"
	"time"

	manager "github.com/Aditya-0011/common/contracts/go/manager"
	"github.com/jackc/pgx/v5"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
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
	ctx, cancel := context.WithTimeout(c, utils.TimeoutDuration)
	defer cancel()

	var (
		about      string
		coverImage string
		updatedAt  time.Time
	)

	query := `select about, coverImage, updatedAt from portfolio.user where id = @id`

	queryParams := pgx.NamedArgs{
		"id": req.GetUserId(),
	}

	err := us.postgres.Pool.QueryRow(ctx, query, queryParams).Scan(&about, &coverImage, &updatedAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, status.Errorf(codes.NotFound, "User not found")
		}
		slog.Error("error querying database", "userId", req.GetUserId(), "error", err)
		return nil, status.Errorf(codes.Internal, "Internal server error")
	}

	return &manager.GetUserDetailsResponse{
		About:      about,
		CoverImage: coverImage,
		UpdatedAt:  timestamppb.New(updatedAt),
	}, nil
}

func (us *userServer) EditUserDetails(c context.Context, req *manager.EditUserDetailsRequest) (*manager.SimpleResponse, error) {
	ctx, cancel := context.WithTimeout(c, utils.TimeoutDuration)
	defer cancel()

	var outcode int8

	query := `select outcode from portfolio.edit_user(@id, @about, @coverImage)`

	queryParams := pgx.NamedArgs{
		"id":         req.GetUserId(),
		"about":      req.GetAbout(),
		"coverImage": req.GetCoverImage(),
	}

	err := us.postgres.Pool.QueryRow(ctx, query, queryParams).Scan(&outcode)
	if err != nil {
		slog.Error("error creating user", "userId", req.GetUserId(), "error", err)
		return nil, status.Errorf(codes.Internal, "Error updating user")
	}

	switch outcode {
	case -1:
		return &manager.SimpleResponse{Message: "User updated successfully"}, nil
	default:
		return nil, status.Errorf(codes.Internal, "Error updating user")
	}
}
