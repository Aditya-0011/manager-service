package portfolio

import (
	"context"
	"log/slog"
	"manager/lib"
	"manager/utils"
	"time"

	manager "github.com/Aditya-0011/common/contracts/go/manager"
	"github.com/jackc/pgx/v5"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (ps *portfolioServer) GetExperiences(c context.Context, req *manager.SimpleRequest) (*manager.GetExperiencesResponse, error) {
	ctx, cancel := context.WithTimeout(c, utils.TimeoutDuration)
	defer cancel()

	query := `
	select 
		e.id, e.company, e.start, e."end", e.tenure, e.updatedAt,
		coalesce((
			select jsonb_agg(
				jsonb_build_object(
					'id', t.id,
					'name', t.name,
					'image_url', t.imageUrl,
					'fallback_image_url', t.fallbackImageUrl
				)
			)
			from portfolio.technology_experience te
			join portfolio.technology t on te.technologyId = t.id
			where te.experienceId = e.id
		), '[]'::jsonb) as technologies,
		coalesce((
			select jsonb_agg(
				jsonb_build_object(
					'id', pos.id,
					'role', pos.role,
					'start', pos.start,
					'end', pos."end",
					'work_done', pos.workDone,
					'projects', coalesce((
						select jsonb_agg(
							jsonb_build_object(
								'id', pr.id,
								'name', pr.name
							)
						)
						from portfolio.project_position pp
						join portfolio.project pr on pp.projectId = pr.id
						where pp.positionId = pos.id
					), '[]'::jsonb)
				)
			)
			from portfolio.position pos
			where pos.experienceId = e.id
		), '[]'::jsonb) as positions
	from portfolio.experience e
	where e.userId = @userId
	order by e.start desc
	`

	queryParams := pgx.NamedArgs{
		"userId": req.GetUserId(),
	}

	rows, err := ps.postgres.Pool.Query(ctx, query, queryParams)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, status.Errorf(codes.NotFound, "experiences not found")
		}
		slog.Error("error querying database", "userId", req.GetUserId(), "error", err)
		return nil, status.Errorf(codes.Internal, "internal server error")
	}
	defer rows.Close()

	var resExperiences []*manager.Experience

	for rows.Next() {
		var (
			id           int32
			company      string
			start        string
			end          *string
			tenure       *string
			updatedAt    time.Time
			technologies []*manager.TechnologySummary
			positions    []*manager.Position
		)

		err := rows.Scan(&id, &company, &start, &end, &tenure, &updatedAt, &technologies, &positions)
		if err != nil {
			slog.Error("error scanning rows", "userId", req.GetUserId(), "error", err)
			return nil, status.Errorf(codes.Internal, "internal server error")
		}

		resEnd := ""
		if end != nil {
			resEnd = *end
		}

		resTenure := ""
		if tenure != nil {
			resTenure = *tenure
		}

		resExperiences = append(resExperiences, &manager.Experience{
			Id:           id,
			Company:      company,
			Start:        start,
			End:          resEnd,
			Tenure:       resTenure,
			Technologies: technologies,
			Positions:    positions,
			UpdatedAt:    timestamppb.New(updatedAt),
		})
	}

	if err := rows.Err(); err != nil {
		slog.Error("error iterating rows", "userId", req.GetUserId(), "error", err)
		return nil, status.Errorf(codes.Internal, "internal server error")
	}

	if len(resExperiences) == 0 {
		return nil, status.Errorf(codes.NotFound, "experiences not found")
	}

	return &manager.GetExperiencesResponse{
		Experiences: resExperiences,
	}, nil
}

func (ps *portfolioServer) CreateExperience(c context.Context, req *manager.ExperienceCreateRequest) (*manager.SimpleResponse, error) {
	ctx, cancel := context.WithTimeout(c, utils.TimeoutDuration)
	defer cancel()

	query := `select outcode from portfolio.edit_experience(
				@id, 
				@userId, 
				@company, 
				@positions, 
				@technologies
			  )`

	dbPositions := make([]lib.PositionCreate, len(req.GetPositions()))
	for i, p := range req.GetPositions() {
		dbPositions[i] = lib.PositionCreate{
			Id:       -1,
			Role:     p.GetRole(),
			Start:    p.GetStart(),
			End:      p.End,
			WorkDone: p.GetWorkDone(),
			Projects: p.GetProjects(),
		}
	}

	queryParams := pgx.NamedArgs{
		"id":           -1,
		"userId":       req.GetUserId(),
		"company":      req.GetCompany(),
		"positions":    dbPositions,
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
			Message: "Experience created successfully",
		}, nil
	default:
		return nil, status.Errorf(codes.Internal, "internal server error")
	}
}

func (ps *portfolioServer) UpdateExperience(c context.Context, req *manager.ExperienceUpdateRequest) (*manager.SimpleResponse, error) {
	ctx, cancel := context.WithTimeout(c, utils.TimeoutDuration)
	defer cancel()

	query := `select outcode from portfolio.edit_experience(
				@id, 
				@userId, 
				@company, 
				@positions, 
				@technologies
			  )`

	dbPositions := make([]lib.PositionCreate, len(req.GetPositions()))
	for i, p := range req.GetPositions() {

		id := p.GetId()
		if id == 0 {
			id = -1
		}

		dbPositions[i] = lib.PositionCreate{
			Id:       id,
			Role:     p.GetRole(),
			Start:    p.GetStart(),
			End:      p.End,
			WorkDone: p.GetWorkDone(),
			Projects: p.GetProjects(),
		}
	}

	queryParams := pgx.NamedArgs{
		"id":           req.GetId(),
		"userId":       req.GetUserId(),
		"company":      req.GetCompany(),
		"positions":    dbPositions,
		"technologies": req.GetTechnologies(),
	}

	var outcode int8

	err := ps.postgres.Pool.QueryRow(ctx, query, queryParams).Scan(&outcode)
	if err != nil {
		slog.Error("error querying database", "experienceId", req.GetId(), "userId", req.GetUserId(), "error", err)
		return nil, status.Errorf(codes.Internal, "internal server error")
	}

	switch outcode {
	case -1:
		return &manager.SimpleResponse{
			Message: "Experience updated successfully",
		}, nil
	case 1:
		return nil, status.Errorf(codes.NotFound, "experience not found")
	default:
		return nil, status.Errorf(codes.Internal, "internal server error")
	}
}

func (ps *portfolioServer) DeleteExperience(c context.Context, req *manager.DeleteRequest) (*manager.SimpleResponse, error) {
	ctx, cancel := context.WithTimeout(c, utils.TimeoutDuration)
	defer cancel()

	query := `select outcode from portfolio.delete_experience(@id, @userId)`

	queryParams := pgx.NamedArgs{
		"id":     req.GetId(),
		"userId": req.GetUserId(),
	}

	var outcode int8

	err := ps.postgres.Pool.QueryRow(ctx, query, queryParams).Scan(&outcode)
	if err != nil {
		slog.Error("error querying database", "experienceId", req.GetId(), "userId", req.GetUserId(), "error", err)
		return nil, status.Errorf(codes.Internal, "internal server error")
	}

	switch outcode {
	case -1:
		return &manager.SimpleResponse{
			Message: "Experience deleted successfully",
		}, nil
	case 1:
		return nil, status.Errorf(codes.NotFound, "experience not found")
	default:
		return nil, status.Errorf(codes.Internal, "internal server error")
	}
}
