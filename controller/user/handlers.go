package user

import (
	"context"
	"log/slog"
	"manager/db"
	"manager/utils"
	"time"

	manager "github.com/Aditya-0011/common/contracts/go/manager"
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

	query := `select about, coverImage, updatedAt from portfolio.user where id = $1`

	err := us.postgres.Pool.QueryRow(ctx, query, req.GetUserId()).Scan(&about, &coverImage, &updatedAt)
	if err != nil {
		slog.Error("error querying database", "userId", req.GetUserId(), "error", err)
		return nil, status.Errorf(codes.Internal, "internal server error")
	}

	if about == "" || coverImage == "" || updatedAt.IsZero() {
		return nil, status.Errorf(codes.NotFound, "data not found")
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

	err := us.postgres.Pool.QueryRow(ctx, "select * from portfolio.edit_user($1, $2, $3)", req.GetUserId(), req.GetAbout(), req.GetCoverImage()).Scan(&outcode)
	if err != nil {
		slog.Error("error creating user", "userId", req.GetUserId(), "error", err)
		return nil, status.Errorf(codes.Internal, "error updating user")
	}

	switch outcode {
	case -1:
		return &manager.SimpleResponse{Message: "User updated successfully"}, nil
	default:
		return nil, status.Errorf(codes.Internal, "error updating user")
	}
}
