package portfolio

import (
	"context"
	"log/slog"
	"manager/utils"
	"time"

	manager "github.com/Aditya-0011/common/contracts/go/manager"
	"github.com/jackc/pgx/v5"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (ps *portfolioServer) GetProjects(c context.Context, req *manager.SimpleRequest) (*manager.GetProjectsResponse, error) {
	ctx, cancel := context.WithTimeout(c, utils.TimeoutDuration)
	defer cancel()

	query := `
	select 
		p.id, p.name, p.description, p.imageUrl, p.projectUrl, p.githubUrl, p.featured, p.updatedAt,
		coalesce(
			jsonb_agg(
				jsonb_build_object(
					'id', t.id,
					'name', t.name,
					'image_url', t.imageUrl,
					'fallback_image_url', t.fallbackImageUrl
				)
			) filter (where t.id is not null), '[]'::jsonb
		) as technologies
	from portfolio.project p
	left join portfolio.technology_project tp on p.id = tp.projectId
	left join portfolio.technology t on tp.technologyId = t.id
	where p.userId = @userId
	group by p.id
	`

	queryParams := pgx.NamedArgs{
		"userId": req.GetUserId(),
	}

	rows, err := ps.postgres.Pool.Query(ctx, query, queryParams)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, status.Errorf(codes.NotFound, "projects not found")
		}
		slog.Error("error querying database", "userId", req.GetUserId(), "error", err)
		return nil, status.Errorf(codes.Internal, "internal server error")
	}
	defer rows.Close()

	var resProjects []*manager.Project

	for rows.Next() {
		var (
			id           int32
			name         string
			description  string
			imageUrl     string
			projectUrl   *string
			githubUrl    *string
			featured     bool
			updatedAt    time.Time
			technologies []*manager.TechnologySummary
		)

		err := rows.Scan(&id, &name, &description, &imageUrl, &projectUrl, &githubUrl, &featured, &updatedAt, &technologies)
		if err != nil {
			slog.Error("error scanning rows", "userId", req.GetUserId(), "error", err)
			return nil, status.Errorf(codes.Internal, "internal server error")
		}

		resProjects = append(resProjects, &manager.Project{
			Id:           id,
			Name:         name,
			Description:  description,
			ImageUrl:     imageUrl,
			ProjectUrl:   projectUrl,
			GithubUrl:    githubUrl,
			Featured:     featured,
			Technologies: technologies,
			UpdatedAt:    timestamppb.New(updatedAt),
		})
	}

	if err := rows.Err(); err != nil {
		slog.Error("error iterating rows", "userId", req.GetUserId(), "error", err)
		return nil, status.Errorf(codes.Internal, "internal server error")
	}

	if len(resProjects) == 0 {
		return nil, status.Errorf(codes.NotFound, "technologies not found")
	}

	return &manager.GetProjectsResponse{
		Projects: resProjects,
	}, nil
}

func (ps *portfolioServer) CreateProject(c context.Context, req *manager.ProjectCreateRequest) (*manager.SimpleResponse, error) {
	ctx, cancel := context.WithTimeout(c, utils.TimeoutDuration)
	defer cancel()

	query := `select outcode from portfolio.edit_project(
				@id, 
				@userId, 
				@name, 
				@description, 
				@imageUrl, 
				@projectUrl, 
				@githubUrl, 
				@featured,
				@technologies
			  )`

	queryParams := pgx.NamedArgs{
		"id":           -1,
		"userId":       req.GetUserId(),
		"name":         req.GetName(),
		"description":  req.GetDescription(),
		"imageUrl":     req.GetImageUrl(),
		"projectUrl":   req.GetProjectUrl(),
		"githubUrl":    req.GetGithubUrl(),
		"featured":     req.GetFeatured(),
		"technologies": req.GetTechnologies(),
	}

	var outcode int8

	err := ps.postgres.Pool.QueryRow(ctx, query, queryParams).Scan(&outcode)
	if err != nil {
		slog.Error("error querying database", "userId", req.GetUserId(), "error", err)
		return nil, status.Errorf(codes.Internal, "internal server error")
	}

	switch outcode {
	case -1:
		return &manager.SimpleResponse{
			Message: "Project created successfully",
		}, nil
	default:
		return nil, status.Errorf(codes.Internal, "internal server error")
	}
}

func (ps *portfolioServer) UpdateProject(c context.Context, req *manager.ProjectUpdateRequest) (*manager.SimpleResponse, error) {
	ctx, cancel := context.WithTimeout(c, utils.TimeoutDuration)
	defer cancel()

	query := `select outcode from portfolio.edit_project(
				@id, 
				@userId, 
				@name, 
				@description, 
				@imageUrl, 
				@projectUrl, 
				@githubUrl, 
				@featured,
				@technologies
			  )`

	queryParams := pgx.NamedArgs{
		"id":           req.GetId(),
		"userId":       req.GetUserId(),
		"name":         req.GetName(),
		"description":  req.GetDescription(),
		"imageUrl":     req.GetImageUrl(),
		"projectUrl":   req.GetProjectUrl(),
		"githubUrl":    req.GetGithubUrl(),
		"featured":     req.GetFeatured(),
		"technologies": req.GetTechnologies(),
	}

	var outcode int8

	err := ps.postgres.Pool.QueryRow(ctx, query, queryParams).Scan(&outcode)
	if err != nil {
		slog.Error("error querying database", "projectId", req.GetId(), "userId", req.GetUserId(), "error", err)
		return nil, status.Errorf(codes.Internal, "internal server error")
	}

	switch outcode {
	case -1:
		return &manager.SimpleResponse{
			Message: "Project updated successfully",
		}, nil
	case 1:
		return nil, status.Errorf(codes.NotFound, "project not found")
	default:
		return nil, status.Errorf(codes.Internal, "internal server error")
	}
}

func (ps *portfolioServer) DeleteProject(c context.Context, req *manager.DeleteRequest) (*manager.SimpleResponse, error) {
	ctx, cancel := context.WithTimeout(c, utils.TimeoutDuration)
	defer cancel()

	query := `select outcode from portfolio.delete_project(@id, @userId)`

	queryParams := pgx.NamedArgs{
		"id":     req.GetId(),
		"userId": req.GetUserId(),
	}

	var outcode int8

	err := ps.postgres.Pool.QueryRow(ctx, query, queryParams).Scan(&outcode)
	if err != nil {
		slog.Error("error querying database", "projectId", req.GetId(), "userId", req.GetUserId(), "error", err)
		return nil, status.Errorf(codes.Internal, "internal server error")
	}

	switch outcode {
	case -1:
		return &manager.SimpleResponse{
			Message: "Project deleted successfully",
		}, nil
	case 1:
		return nil, status.Errorf(codes.NotFound, "project not found")
	default:
		return nil, status.Errorf(codes.Internal, "internal server error")
	}
}
