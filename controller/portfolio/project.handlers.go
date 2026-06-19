package portfolio

import (
	"context"
	"log/slog"
	"manager/internal/faults"
	"manager/internal/timeout"
	"time"

	"github.com/Aditya-0011/common/contracts/go/manager"
	"github.com/bytedance/sonic"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (ps *portfolioServer) GetProjects(c context.Context, req *manager.SimpleRequest) (*manager.GetProjectsResponse, error) {
	ctx, cancel := timeout.WithDeadline(c, timeout.Duration)
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
	where p.userId = $1
	group by p.id
	`

	rows, err := ps.postgres.Pool.Query(ctx, query, req.GetUserId())

	if err != nil {
		slog.LogAttrs(ctx, slog.LevelError, "error querying database",
			slog.Int64("userId", int64(req.GetUserId())),
			slog.String("error", err.Error()),
		)
		return nil, faults.ErrInternal
	}
	defer rows.Close()

	resProjects := make([]*manager.Project, 0, 16)

	for rows.Next() {
		var (
			id          int32
			name        string
			description string
			imageUrl    string
			projectUrl  *string
			githubUrl   *string
			featured    bool
			updatedAt   time.Time
			techBytes   []byte
		)

		err := rows.Scan(&id, &name, &description, &imageUrl, &projectUrl, &githubUrl, &featured, &updatedAt, &techBytes)
		if err != nil {
			slog.LogAttrs(ctx, slog.LevelError, "error scanning rows",
				slog.Int64("userId", int64(req.GetUserId())),
				slog.String("error", err.Error()),
			)
			return nil, faults.ErrInternal
		}

		var technologies []*manager.TechnologySummary
		if len(techBytes) > 0 {
			if err := sonic.Unmarshal(techBytes, &technologies); err != nil {
				slog.LogAttrs(ctx, slog.LevelError, "error unmarshaling technologies", slog.String("error", err.Error()))
			}
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
		slog.LogAttrs(ctx, slog.LevelError, "error iterating rows",
			slog.Int64("userId", int64(req.GetUserId())),
			slog.String("error", err.Error()),
		)
		return nil, faults.ErrInternal
	}

	if len(resProjects) == 0 {
		return nil, faults.ErrProjectsNotFound
	}

	return &manager.GetProjectsResponse{
		Projects: resProjects,
	}, nil
}

func (ps *portfolioServer) CreateProject(c context.Context, req *manager.ProjectCreateRequest) (*manager.SimpleResponse, error) {
	ctx, cancel := timeout.WithDeadline(c, timeout.Duration)
	defer cancel()

	query := `select outcode from portfolio.edit_project($1, $2, $3, $4, $5, $6, $7, $8, $9)`

	var outcode int16

	err := ps.postgres.Pool.QueryRow(ctx, query,
		-1,
		req.GetUserId(),
		req.GetName(),
		req.GetDescription(),
		req.GetImageUrl(),
		req.GetProjectUrl(),
		req.GetGithubUrl(),
		req.GetFeatured(),
		req.GetTechnologies(),
	).Scan(&outcode)
	if err != nil {
		slog.LogAttrs(ctx, slog.LevelError, "error querying database",
			slog.Int64("userId", int64(req.GetUserId())),
			slog.String("error", err.Error()),
		)
		return nil, faults.ErrInternalCreate
	}

	switch outcode {
	case -1:
		return &manager.SimpleResponse{
			Message: "Project created successfully",
		}, nil
	default:
		return nil, faults.ErrInternalCreate
	}
}

func (ps *portfolioServer) UpdateProject(c context.Context, req *manager.ProjectUpdateRequest) (*manager.SimpleResponse, error) {
	ctx, cancel := timeout.WithDeadline(c, timeout.Duration)
	defer cancel()

	query := `select outcode from portfolio.edit_project($1, $2, $3, $4, $5, $6, $7, $8, $9)`

	var outcode int16

	err := ps.postgres.Pool.QueryRow(ctx, query,
		req.GetId(),
		req.GetUserId(),
		req.GetName(),
		req.GetDescription(),
		req.GetImageUrl(),
		req.GetProjectUrl(),
		req.GetGithubUrl(),
		req.GetFeatured(),
		req.GetTechnologies(),
	).Scan(&outcode)
	if err != nil {
		slog.LogAttrs(ctx, slog.LevelError, "error querying database",
			slog.Int64("projectId", int64(req.GetId())),
			slog.Int64("userId", int64(req.GetUserId())),
			slog.String("error", err.Error()),
		)
		return nil, faults.ErrInternalUpdate
	}

	switch outcode {
	case -1:
		return &manager.SimpleResponse{
			Message: "Project updated successfully",
		}, nil
	case 1:
		return nil, faults.ErrProjectNotFound
	default:
		return nil, faults.ErrInternalUpdate
	}
}

func (ps *portfolioServer) DeleteProject(c context.Context, req *manager.DeleteRequest) (*manager.SimpleResponse, error) {
	ctx, cancel := timeout.WithDeadline(c, timeout.Duration)
	defer cancel()

	query := `select outcode from portfolio.delete_project($1, $2)`

	var outcode int16

	err := ps.postgres.Pool.QueryRow(ctx, query, req.GetId(), req.GetUserId()).Scan(&outcode)
	if err != nil {
		slog.LogAttrs(ctx, slog.LevelError, "error querying database",
			slog.Int64("projectId", int64(req.GetId())),
			slog.Int64("userId", int64(req.GetUserId())),
			slog.String("error", err.Error()),
		)
		return nil, faults.ErrInternalDelete
	}

	switch outcode {
	case -1:
		return &manager.SimpleResponse{
			Message: "Project deleted successfully",
		}, nil
	case 1:
		return nil, faults.ErrProjectNotFound
	default:
		return nil, faults.ErrInternalDelete
	}
}
