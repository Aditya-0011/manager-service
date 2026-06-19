package portfolio

import (
	"context"
	"fmt"
	"log/slog"
	"manager/internal/faults"
	"manager/internal/model"
	"manager/internal/timeout"
	"slices"
	"time"

	"github.com/Aditya-0011/common/contracts/go/manager"
	"github.com/bytedance/sonic"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (ps *portfolioServer) GetExperiences(c context.Context, req *manager.SimpleRequest) (*manager.GetExperiencesResponse, error) {
	ctx, cancel := timeout.WithDeadline(c, timeout.Duration)
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
	where e.userId = $1
	order by e.start desc
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

	resExperiences := make([]*manager.Experience, 0, 16)

	for rows.Next() {
		var (
			id        int32
			company   string
			start     string
			end       *string
			tenure    *string
			updatedAt time.Time
			techBytes []byte
			posBytes  []byte
		)

		err := rows.Scan(&id, &company, &start, &end, &tenure, &updatedAt, &techBytes, &posBytes)
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

		var positions []*manager.Position
		if len(posBytes) > 0 {
			if err := sonic.Unmarshal(posBytes, &positions); err != nil {
				slog.LogAttrs(ctx, slog.LevelError, "error unmarshaling positions", slog.String("error", err.Error()))
			}
		}

		resEnd := ""
		if end != nil {
			resEnd = *end
		}

		resTenure := ""
		if tenure != nil && *tenure != "" {
			resTenure = *tenure
		} else if resEnd == "Present" {
			resTenure = calculateTotalTenure(positions)
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
		slog.LogAttrs(ctx, slog.LevelError, "error iterating rows",
			slog.Int64("userId", int64(req.GetUserId())),
			slog.String("error", err.Error()),
		)
		return nil, faults.ErrInternal
	}

	if len(resExperiences) == 0 {
		return nil, faults.ErrExperiencesNotFound
	}

	return &manager.GetExperiencesResponse{
		Experiences: resExperiences,
	}, nil
}

func (ps *portfolioServer) CreateExperience(c context.Context, req *manager.ExperienceCreateRequest) (*manager.SimpleResponse, error) {
	ctx, cancel := timeout.WithDeadline(c, timeout.Duration)
	defer cancel()

	query := `select outcode from portfolio.edit_experience($1, $2, $3, $4, $5)`
	for i, p := range req.GetPositions() {
		if err := validatePosition(p.GetStart(), p.End, i); err != nil {
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}
	}

	dbPositions := make([]model.PositionCreate, len(req.GetPositions()))
	for i, p := range req.GetPositions() {
		dbPositions[i] = model.PositionCreate{
			Id:       -1,
			Role:     p.GetRole(),
			Start:    p.GetStart(),
			End:      p.End,
			WorkDone: p.GetWorkDone(),
			Projects: p.GetProjects(),
		}
	}

	var outcode int16

	err := ps.postgres.Pool.QueryRow(ctx, query,
		-1,
		req.GetUserId(),
		req.GetCompany(),
		dbPositions,
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
			Message: "Experience created successfully",
		}, nil
	default:
		return nil, faults.ErrInternalCreate
	}
}

func (ps *portfolioServer) UpdateExperience(c context.Context, req *manager.ExperienceUpdateRequest) (*manager.SimpleResponse, error) {
	ctx, cancel := timeout.WithDeadline(c, timeout.Duration)
	defer cancel()

	query := `select outcode from portfolio.edit_experience($1, $2, $3, $4, $5)`
	for i, p := range req.GetPositions() {
		if err := validatePosition(p.GetStart(), p.End, i); err != nil {
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}
	}

	dbPositions := make([]model.PositionCreate, len(req.GetPositions()))
	for i, p := range req.GetPositions() {

		id := p.GetId()
		if id == 0 {
			id = -1
		}

		dbPositions[i] = model.PositionCreate{
			Id:       id,
			Role:     p.GetRole(),
			Start:    p.GetStart(),
			End:      p.End,
			WorkDone: p.GetWorkDone(),
			Projects: p.GetProjects(),
		}
	}

	var outcode int16

	err := ps.postgres.Pool.QueryRow(ctx, query,
		req.GetId(),
		req.GetUserId(),
		req.GetCompany(),
		dbPositions,
		req.GetTechnologies(),
	).Scan(&outcode)
	if err != nil {
		slog.LogAttrs(ctx, slog.LevelError, "error querying database",
			slog.Int64("experienceId", int64(req.GetId())),
			slog.Int64("userId", int64(req.GetUserId())),
			slog.String("error", err.Error()),
		)
		return nil, faults.ErrInternalUpdate
	}

	switch outcode {
	case -1:
		return &manager.SimpleResponse{
			Message: "Experience updated successfully",
		}, nil
	case 1:
		return nil, faults.ErrExperienceNotFound
	default:
		return nil, faults.ErrInternalUpdate
	}
}

func (ps *portfolioServer) DeleteExperience(c context.Context, req *manager.DeleteRequest) (*manager.SimpleResponse, error) {
	ctx, cancel := timeout.WithDeadline(c, timeout.Duration)
	defer cancel()

	query := `select outcode from portfolio.delete_experience($1, $2)`

	var outcode int16

	err := ps.postgres.Pool.QueryRow(ctx, query, req.GetId(), req.GetUserId()).Scan(&outcode)
	if err != nil {
		slog.LogAttrs(ctx, slog.LevelError, "error querying database",
			slog.Int64("experienceId", int64(req.GetId())),
			slog.Int64("userId", int64(req.GetUserId())),
			slog.String("error", err.Error()),
		)
		return nil, faults.ErrInternalDelete
	}

	switch outcode {
	case -1:
		return &manager.SimpleResponse{
			Message: "Experience deleted successfully",
		}, nil
	case 1:
		return nil, faults.ErrExperienceNotFound
	default:
		return nil, faults.ErrInternalDelete
	}
}

func calculateTotalTenure(positions []*manager.Position) string {
	if len(positions) == 0 {
		return ""
	}

	now := time.Now()

	type interval struct{ start, end time.Time }
	intervals := make([]interval, 0, len(positions))

	for _, pos := range positions {
		var start, end time.Time

		if len(pos.Start) == 7 {
			start, _ = time.Parse("2006-01", pos.Start)
		} else {
			start, _ = time.Parse("2006-01-02", pos.Start)
		}

		if pos.End == "" || pos.End == "Present" {
			end = now
		} else if len(pos.End) == 7 {
			end, _ = time.Parse("2006-01", pos.End)
		} else {
			end, _ = time.Parse("2006-01-02", pos.End)
		}

		if start.IsZero() || end.IsZero() {
			continue
		}
		if end.After(now) {
			end = now
		}
		start = time.Date(start.Year(), start.Month(), 1, 0, 0, 0, 0, time.UTC)
		end = time.Date(end.Year(), end.Month(), 1, 0, 0, 0, 0, time.UTC)
		intervals = append(intervals, interval{start, end})
	}

	if len(intervals) == 0 {
		return ""
	}

	slices.SortFunc(intervals, func(a, b interval) int {
		return a.start.Compare(b.start)
	})

	merged := intervals[:1]
	for _, iv := range intervals[1:] {
		last := &merged[len(merged)-1]

		if !iv.start.After(last.end.AddDate(0, 1, 0)) {
			if iv.end.After(last.end) {
				last.end = iv.end
			}
		} else {
			merged = append(merged, iv)
		}
	}

	totalMonths := 0
	for _, iv := range merged {
		months := (iv.end.Year()-iv.start.Year())*12 + int(iv.end.Month()-iv.start.Month()) + 1
		totalMonths += months
	}

	if totalMonths <= 0 {
		return ""
	}

	years := totalMonths / 12
	months := totalMonths % 12

	if years > 0 && months > 0 {
		return fmt.Sprintf("%d yrs %d mos", years, months)
	} else if years > 0 {
		return fmt.Sprintf("%d yrs", years)
	}
	return fmt.Sprintf("%d mos", months)
}

func validatePosition(startStr string, endStr *string, index int) error {
	var start, end time.Time
	var err error

	if len(startStr) == 7 {
		start, err = time.Parse("2006-01", startStr)
	} else {
		start, err = time.Parse("2006-01-02", startStr)
	}
	if err != nil {
		return fmt.Errorf("Invalid start date format in position %d", index+1)
	}

	if endStr != nil && *endStr != "" && *endStr != "Present" {
		if len(*endStr) == 7 {
			end, err = time.Parse("2006-01", *endStr)
		} else {
			end, err = time.Parse("2006-01-02", *endStr)
		}
		if err != nil {
			return fmt.Errorf("Invalid end date format in position %d", index+1)
		}

		if start.After(end) {
			return fmt.Errorf("Start date cannot be greater than end date in position %d", index+1)
		}
	}
	return nil
}
