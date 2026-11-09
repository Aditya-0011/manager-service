package portfolio

import (
	"context"
	"log/slog"
	"manager/internal/faults"
	"time"

	"github.com/Aditya-0011/common/contracts/go/manager"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (ps *portfolioServer) GetTechnologies(ctx context.Context, req *manager.SimpleRequest) (*manager.GetTechnologiesResponse, error) {
	query := `select 
				id, 
				name, 
				imageUrl, 
				fallbackImageUrl, 
				category, 
				updatedAt 
			  from portfolio.technology 
			  where userId = $1
			  order by category, name`

	rows, err := ps.pool.Query(ctx, query, req.GetUserId())

	if err != nil {
		slog.LogAttrs(ctx, slog.LevelError, "error querying database",
			slog.Int("userId", int(req.GetUserId())),
			slog.String("error", err.Error()),
		)
		return nil, faults.ErrInternal
	}
	defer rows.Close()

	resTechnologies := make([]*manager.Technology, 0, 16)

	for rows.Next() {
		var (
			id               int32
			name             string
			imageUrl         string
			fallbackImageUrl string
			category         int16
			updatedAt        time.Time
		)

		err := rows.Scan(&id, &name, &imageUrl, &fallbackImageUrl, &category, &updatedAt)
		if err != nil {
			slog.LogAttrs(ctx, slog.LevelError, "error scanning rows",
				slog.Int("userId", int(req.GetUserId())),
				slog.String("error", err.Error()),
			)
			return nil, faults.ErrInternal
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
		slog.LogAttrs(ctx, slog.LevelError, "error iterating rows",
			slog.Int("userId", int(req.GetUserId())),
			slog.String("error", err.Error()),
		)
		return nil, faults.ErrInternal
	}

	return &manager.GetTechnologiesResponse{
		Technologies: resTechnologies,
	}, nil
}

func (ps *portfolioServer) CreateTechnology(ctx context.Context, req *manager.TechnologyCreateRequest) (*manager.SimpleResponse, error) {
	query := `select outcode from portfolio.edit_technology($1, $2, $3, $4, $5, $6)`

	var outcode int16

	err := ps.pool.QueryRow(ctx, query,
		-1,
		req.GetUserId(),
		req.GetName(),
		req.GetImageUrl(),
		req.GetFallbackImageUrl(),
		req.GetCategory(),
	).Scan(&outcode)
	if err != nil {
		slog.LogAttrs(ctx, slog.LevelError, "error querying database",
			slog.Int("userId", int(req.GetUserId())),
			slog.String("error", err.Error()),
		)
		return nil, faults.ErrInternalCreate
	}

	switch outcode {
	case -1:
		return &manager.SimpleResponse{
			Message: "Technology created successfully",
		}, nil
	default:
		return nil, faults.ErrInternalCreate
	}
}

func (ps *portfolioServer) UpdateTechnology(ctx context.Context, req *manager.TechnologyUpdateRequest) (*manager.SimpleResponse, error) {
	query := `select outcode from portfolio.edit_technology($1, $2, $3, $4, $5, $6)`

	var outcode int16

	err := ps.pool.QueryRow(ctx, query,
		req.GetId(),
		req.GetUserId(),
		req.GetName(),
		req.GetImageUrl(),
		req.GetFallbackImageUrl(),
		req.GetCategory(),
	).Scan(&outcode)
	if err != nil {
		slog.LogAttrs(ctx, slog.LevelError, "error querying database",
			slog.Int("technologyId", int(req.GetId())),
			slog.Int("userId", int(req.GetUserId())),
			slog.String("error", err.Error()),
		)
		return nil, faults.ErrInternalUpdate
	}

	switch outcode {
	case -1:
		return &manager.SimpleResponse{
			Message: "Technology updated successfully",
		}, nil
	case 1:
		return nil, faults.ErrTechnologyNotFound
	default:
		return nil, faults.ErrInternalUpdate
	}
}

func (ps *portfolioServer) DeleteTechnology(ctx context.Context, req *manager.DeleteRequest) (*manager.SimpleResponse, error) {
	query := `select outcode from portfolio.delete_technology($1, $2)`

	var outcode int16

	err := ps.pool.QueryRow(ctx, query, req.GetId(), req.GetUserId()).Scan(&outcode)
	if err != nil {
		slog.LogAttrs(ctx, slog.LevelError, "error querying database",
			slog.Int("technologyId", int(req.GetId())),
			slog.Int("userId", int(req.GetUserId())),
			slog.String("error", err.Error()),
		)
		return nil, faults.ErrInternalDelete
	}

	switch outcode {
	case -1:
		return &manager.SimpleResponse{
			Message: "Technology deleted successfully",
		}, nil
	case 1:
		return nil, faults.ErrTechnologyNotFound
	default:
		return nil, faults.ErrInternalDelete
	}
}
