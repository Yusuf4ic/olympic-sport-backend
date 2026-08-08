package postgres

import (
	"context"

	"github.com/Devesanoff/olympic-sport-backend/internal/domain"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/sync/errgroup"
)

type syncRepository struct {
	db *pgxpool.Pool
}

// NewSyncRepository creates a new instance of syncRepository.
func NewSyncRepository(db *pgxpool.Pool) domain.SyncRepository {
	return &syncRepository{db: db}
}

func (r *syncRepository) GetOfflinePackage(ctx context.Context) (*domain.OfflinePackage, error) {
	pkg := &domain.OfflinePackage{
		Categories:           make([]*domain.Category, 0),
		Zones:                make([]*domain.Zone, 0),
		CategoryAllowedZones: make([]*domain.CategoryAllowedZone, 0),
		MealSchedules:        make([]*domain.MealSchedule, 0),
		Participants:         make([]*domain.SyncParticipant, 0),
	}

	g, ctx := errgroup.WithContext(ctx)

	// Fetch Categories
	g.Go(func() error {
		rows, err := r.db.Query(ctx, "SELECT id, name, color_code, can_eat, created_at FROM categories")
		if err != nil {
			return err
		}
		defer rows.Close()

		for rows.Next() {
			var c domain.Category
			if err := rows.Scan(&c.ID, &c.Name, &c.ColorCode, &c.CanEat, &c.CreatedAt); err != nil {
				return err
			}
			pkg.Categories = append(pkg.Categories, &c)
		}
		return rows.Err()
	})

	// Fetch Zones
	g.Go(func() error {
		rows, err := r.db.Query(ctx, "SELECT id, name, code, requires_in_out, created_at FROM zones")
		if err != nil {
			return err
		}
		defer rows.Close()

		for rows.Next() {
			var z domain.Zone
			if err := rows.Scan(&z.ID, &z.Name, &z.Code, &z.RequiresInOut, &z.CreatedAt); err != nil {
				return err
			}
			pkg.Zones = append(pkg.Zones, &z)
		}
		return rows.Err()
	})

	// Fetch CategoryAllowedZones
	g.Go(func() error {
		rows, err := r.db.Query(ctx, "SELECT category_id, zone_id FROM category_allowed_zones")
		if err != nil {
			return err
		}
		defer rows.Close()

		for rows.Next() {
			var caz domain.CategoryAllowedZone
			if err := rows.Scan(&caz.CategoryID, &caz.ZoneID); err != nil {
				return err
			}
			pkg.CategoryAllowedZones = append(pkg.CategoryAllowedZones, &caz)
		}
		return rows.Err()
	})

	// Fetch MealSchedules
	g.Go(func() error {
		rows, err := r.db.Query(ctx, "SELECT id, date::text, meal_type, start_time::text, end_time::text FROM meal_schedules")
		if err != nil {
			return err
		}
		defer rows.Close()

		for rows.Next() {
			var m domain.MealSchedule
			// pgx might return date and time differently, but typically strings or time.Time work.
			// The domain has string for date, start_time, end_time.
			// We can scan them as strings if they are casted.
			if err := rows.Scan(&m.ID, &m.Date, &m.MealType, &m.StartTime, &m.EndTime); err != nil {
				// Fallback if scanning as string fails natively with pgx types, 
				// we cast in the query.
				return err
			}
			pkg.MealSchedules = append(pkg.MealSchedules, &m)
		}
		return rows.Err()
	})

	// Fetch Active Participants
	g.Go(func() error {
		rows, err := r.db.Query(ctx, "SELECT id, category_id, status, qr_token FROM participants WHERE status = 'ACTIVE'")
		if err != nil {
			return err
		}
		defer rows.Close()

		for rows.Next() {
			var p domain.SyncParticipant
			if err := rows.Scan(&p.ID, &p.CategoryID, &p.Status, &p.QRToken); err != nil {
				return err
			}
			pkg.Participants = append(pkg.Participants, &p)
		}
		return rows.Err()
	})

	if err := g.Wait(); err != nil {
		// We'll fix the meal_schedules scan query if it causes type mismatch with pgx string scan
		// To be safe, we cast to text in PostgreSQL:
		return nil, err
	}

	return pkg, nil
}

func (r *syncRepository) BulkInsertAccessLogs(ctx context.Context, logs []domain.AccessLog) error {
	if len(logs) == 0 {
		return nil
	}

	batch := &pgx.Batch{}
	query := `
		INSERT INTO access_logs (id, participant_id, zone_id, direction, status, reason, created_at)
		SELECT $1, $2, $3, $4, $5, $6, $7
		WHERE EXISTS (SELECT 1 FROM participants WHERE id = $2)
		  AND EXISTS (SELECT 1 FROM zones WHERE id = $3)
		ON CONFLICT (id) DO NOTHING
	`
	for _, l := range logs {
		batch.Queue(query, l.ID, l.ParticipantID, l.ZoneID, l.Direction, l.Status, l.Reason, l.CreatedAt)
	}

	results := r.db.SendBatch(ctx, batch)
	defer results.Close()

	for i := 0; i < len(logs); i++ {
		_, _ = results.Exec() // Safely ignore individual row constraint errors during bulk sync
	}

	return nil
}

func (r *syncRepository) BulkInsertMealLogs(ctx context.Context, logs []domain.MealLog) error {
	if len(logs) == 0 {
		return nil
	}

	batch := &pgx.Batch{}
	query := `
		INSERT INTO meal_logs (id, participant_id, meal_schedule_id, meal_type, date, status, reason, created_at)
		SELECT $1, $2, $3, $4, $5, $6, $7, $8
		WHERE EXISTS (SELECT 1 FROM participants WHERE id = $2)
		  AND ($3::int IS NULL OR EXISTS (SELECT 1 FROM meal_schedules WHERE id = $3))
		ON CONFLICT (id) DO NOTHING
	`
	for _, l := range logs {
		batch.Queue(query, l.ID, l.ParticipantID, l.MealScheduleID, l.MealType, l.Date, l.Status, l.Reason, l.CreatedAt)
	}

	results := r.db.SendBatch(ctx, batch)
	defer results.Close()

	for i := 0; i < len(logs); i++ {
		_, _ = results.Exec() // Safely ignore individual row constraint errors during bulk sync
	}

	return nil
}
