package portfolio

import (
	"context"
	"log/slog"
	"manager/utils"
	"time"

	"github.com/Aditya-0011/common/contracts/go/manager"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (ps *portfolioServer) GetTechnologies(c context.Context, req *manager.SimpleRequest) (*manager.GetTechnologiesResponse, error) {
	ctx, cancel := context.WithTimeout(c, utils.TimeoutDuration)
	defer cancel()

	query := `select id, name, imageUrl, fallbackImageUrl, category, updatedAt from portfolio.technology where userId = $1`

	rows, err := ps.postgres.Pool.Query(ctx, query, req.GetUserId())

	if err != nil {
		slog.Error("error querying database", "userId", req.GetUserId(), "error", err)
		return nil, status.Errorf(codes.Internal, "internal server error")
	}
	defer rows.Close()

	var resTechnologies []*manager.Technology

	for rows.Next() {
		var (
			id               int32
			name             string
			imageUrl         string
			fallbackImageUrl string
			category         int8
			updatedAt        time.Time
		)

		err := rows.Scan(&id, &name, &imageUrl, &fallbackImageUrl, &category, &updatedAt)
		if err != nil {
			slog.Error("error scanning rows", "userId", req.GetUserId(), "error", err)
			return nil, status.Errorf(codes.Internal, "internal server error")
		}

		resTechnologies = append(resTechnologies, &manager.Technology{
			Id:               id,
			Name:             name,
			ImageUrl:         imageUrl,
			FallbackImageUrl: fallbackImageUrl,
			Category:         manager.TechnologyCategory(category),
			UpdatedAt:        timestamppb.New(updatedAt),
		})
	}

	if err := rows.Err(); err != nil {
		slog.Error("error iterating rows", "userId", req.GetUserId(), "error", err)
		return nil, status.Errorf(codes.Internal, "internal server error")
	}

	if len(resTechnologies) == 0 {
		return nil, status.Errorf(codes.NotFound, "technologies not found")
	}

	return &manager.GetTechnologiesResponse{
		Technologies: resTechnologies,
	}, nil
}

func (ps *portfolioServer) CreateTechnology(c context.Context, req *manager.TechnologyCreateRequest) (*manager.SimpleResponse, error) {
	ctx, cancel := context.WithTimeout(c, utils.TimeoutDuration)
	defer cancel()

	query := `select * from portfolio.edit_technology($1, $2, $3, $4, $5, $6)`

	var outcode int8

	err := ps.postgres.Pool.QueryRow(ctx, query, -1, req.GetUserId(), req.GetName(), req.GetImageUrl(), req.GetFallbackImageUrl(), req.GetCategory()).Scan(&outcode)
	if err != nil {
		slog.Error("error querying database", "userId", req.GetUserId(), "error", err)
		return nil, status.Errorf(codes.Internal, "internal server error")
	}

	switch outcode {
	case -1:
		return &manager.SimpleResponse{
			Message: "Technology created successfully",
		}, nil
	default:
		return nil, status.Errorf(codes.Internal, "internal server error")
	}
}

func (ps *portfolioServer) UpdateTechnology(c context.Context, req *manager.TechnologyUpdateRequest) (*manager.SimpleResponse, error) {
	ctx, cancel := context.WithTimeout(c, utils.TimeoutDuration)
	defer cancel()

	query := `select * from portfolio.edit_technology($1, $2, $3, $4, $5, $6)`

	var outcode int8

	err := ps.postgres.Pool.QueryRow(ctx, query, req.GetId(), req.GetUserId(), req.GetName(), req.GetImageUrl(), req.GetFallbackImageUrl(), req.GetCategory()).Scan(&outcode)
	if err != nil {
		slog.Error("error querying database", "technologyId", req.GetId(), "userId", req.GetUserId(), "error", err)
		return nil, status.Errorf(codes.Internal, "internal server error")
	}

	switch outcode {
	case -1:
		return &manager.SimpleResponse{
			Message: "Technology updated successfully",
		}, nil
	case 1:
		return nil, status.Errorf(codes.NotFound, "technology not found")
	default:
		return nil, status.Errorf(codes.Internal, "internal server error")
	}
}

func (ps *portfolioServer) DeleteTechnology(c context.Context, req *manager.DeleteRequest) (*manager.SimpleResponse, error) {
	ctx, cancel := context.WithTimeout(c, utils.TimeoutDuration)
	defer cancel()

	query := `select * from portfolio.delete_technology($1, $2)`

	var outcode int8

	err := ps.postgres.Pool.QueryRow(ctx, query, req.GetId(), req.GetUserId()).Scan(&outcode)
	if err != nil {
		slog.Error("error querying database", "technologyId", req.GetId(), "userId", req.GetUserId(), "error", err)
		return nil, status.Errorf(codes.Internal, "internal server error")
	}

	switch outcode {
	case -1:
		return &manager.SimpleResponse{
			Message: "Technology deleted successfully",
		}, nil
	case 1:
		return nil, status.Errorf(codes.NotFound, "technology not found")
	default:
		return nil, status.Errorf(codes.Internal, "internal server error")
	}
}
