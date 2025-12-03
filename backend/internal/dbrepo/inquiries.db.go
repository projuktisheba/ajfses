package dbrepo

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/projuktisheba/ajfses/backend/internal/models"
)

// InquiryRepository holds the database connection pool.
type InquiryRepository struct {
	DB *pgxpool.Pool
}

// newInquiryRepository creates a new instance of the repository.
func newInquiryRepository(db *pgxpool.Pool) *InquiryRepository {
	return &InquiryRepository{DB: db}
}

// Create inserts a new inquiry into the database.
// Note: created_at, updated_at, and inquiry_date are handled by DB defaults unless specified otherwise.
func (r *InquiryRepository) Create(ctx context.Context, i *models.Inquiry) (int64, error) {
	sql := `
		INSERT INTO inquiries (name, contact, subject, message)
		VALUES ($1, $2, $3, $4)
		RETURNING id, inquiry_date, created_at, updated_at
	`

	err := r.DB.QueryRow(ctx, sql, i.Name, i.Contact, i.Subject, i.Message).
		Scan(&i.ID, &i.InquiryDate, &i.CreatedAt, &i.UpdatedAt)

	if err != nil {
		return 0, fmt.Errorf("failed to create inquiry: %w", err)
	}

	return i.ID, nil
}

// GetByID retrieves a single inquiry by its ID.
func (r *InquiryRepository) GetByID(ctx context.Context, id int64) (*models.Inquiry, error) {
	sql := `
		SELECT id, inquiry_date, name, contact, subject, message, status, created_at, updated_at
		FROM inquiries
		WHERE id = $1
	`

	var i models.Inquiry
	err := r.DB.QueryRow(ctx, sql, id).Scan(
		&i.ID,
		&i.InquiryDate,
		&i.Name,
		&i.Contact,
		&i.Subject,
		&i.Message,
		&i.Status,
		&i.CreatedAt,
		&i.UpdatedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("inquiry not found with id: %d", id)
		}
		return nil, fmt.Errorf("failed to get inquiry: %w", err)
	}

	return &i, nil
}

// GetAll retrieves all inquiries, ordered by inquiry_date descending.
func (r *InquiryRepository) GetAll(ctx context.Context) ([]models.Inquiry, error) {
	sql := `
		SELECT id, inquiry_date, name, contact, subject, message, status, created_at, updated_at
		FROM inquiries
		ORDER BY inquiry_date DESC
	`

	rows, err := r.DB.Query(ctx, sql)
	if err != nil {
		return nil, fmt.Errorf("failed to query inquiries: %w", err)
	}
	defer rows.Close()

	var inquiries []models.Inquiry

	for rows.Next() {
		var i models.Inquiry
		err := rows.Scan(
			&i.ID,
			&i.InquiryDate,
			&i.Name,
			&i.Contact,
			&i.Subject,
			&i.Message,
			&i.Status,
			&i.CreatedAt,
			&i.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan inquiry row: %w", err)
		}
		inquiries = append(inquiries, i)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating inquiry rows: %w", err)
	}

	return inquiries, nil
}
// GetStatusCounts retrieves the count of inquiries for specific statuses.
func (r *InquiryRepository) GetStatusCounts(ctx context.Context) (map[string]int, error) {
    // 1. Define SQL to count grouped by status
    sql := `
        SELECT status, COUNT(*)
        FROM inquiries
        WHERE status IN ('NEW', 'RESOLVED')
        GROUP BY status
    `

    rows, err := r.DB.Query(ctx, sql)
    if err != nil {
        return nil, fmt.Errorf("failed to query status counts: %w", err)
    }
    defer rows.Close()

    // 2. Initialize a map to hold the results
    counts := make(map[string]int)

    // 3. Iterate and scan
    for rows.Next() {
        var status string
        var count int
        if err := rows.Scan(&status, &count); err != nil {
            return nil, fmt.Errorf("failed to scan count row: %w", err)
        }
        counts[status] = count
    }

    if err = rows.Err(); err != nil {
        return nil, fmt.Errorf("error iterating count rows: %w", err)
    }

    return counts, nil
}

// Update modifies an existing inquiry.
// It automatically updates the 'updated_at' timestamp.
func (r *InquiryRepository) Update(ctx context.Context, i *models.Inquiry) error {
	sql := `
		UPDATE inquiries
		SET name = $1, contact = $2, subject = $3, message = $4, status = $5, updated_at = CURRENT_TIMESTAMP
		WHERE id = $6
		RETURNING updated_at
	`

	err := r.DB.QueryRow(ctx, sql, i.Name, i.Contact, i.Subject, i.Message, i.Status, i.ID).
		Scan(&i.UpdatedAt)

	if err != nil {
		if err == pgx.ErrNoRows {
			return fmt.Errorf("inquiry not found to update with id: %d", i.ID)
		}
		return fmt.Errorf("failed to update inquiry: %w", err)
	}

	return nil
}

// Delete removes an inquiry from the database.
func (r *InquiryRepository) Delete(ctx context.Context, id int64) error {
	sql := `DELETE FROM inquiries WHERE id = $1`

	cmdTag, err := r.DB.Exec(ctx, sql, id)
	if err != nil {
		return fmt.Errorf("failed to delete inquiry: %w", err)
	}

	if cmdTag.RowsAffected() == 0 {
		return fmt.Errorf("no inquiry found to delete with id: %d", id)
	}

	return nil
}
