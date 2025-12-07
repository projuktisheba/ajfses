package dbrepo

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/projuktisheba/ajfses/backend/internal/models"
)

type GalleryRepository struct {
	DB *pgxpool.Pool
}

// newGalleryRepository creates a new instance of the repository.
func newGalleryRepository(db *pgxpool.Pool) *GalleryRepository {
	return &GalleryRepository{DB: db}
}

// Create inserts a new gallery item into the database.
func (m *GalleryRepository) Create(ctx context.Context, item *models.GalleryItem) (int64, error) {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	stmt := `
		INSERT INTO gallery (title, image_link, created_at, updated_at)
		VALUES ($1, $2, $3, $4)
		RETURNING id
	`

	var id int64
	err := m.DB.QueryRow(ctx, stmt,
		item.Title,
		item.ImageLink, // This might be empty string initially based on your handler
		time.Now().UTC(),
		time.Now().UTC(),
	).Scan(&id)

	if err != nil {
		return 0, fmt.Errorf("failed to insert gallery item: %w", err)
	}

	return id, nil
}

// UpdateImageLink updates only the image_link column for a specific gallery item.
// This is used during the 3-step creation process.
func (m *GalleryRepository) UpdateImageLink(ctx context.Context, id int64, imageLink string) error {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	stmt := `
		UPDATE gallery
		SET image_link = $1, updated_at = $2
		WHERE id = $3
	`

	_, err := m.DB.Exec(ctx, stmt,
		imageLink,
		time.Now().UTC(),
		id,
	)

	if err != nil {
		return fmt.Errorf("failed to update gallery image link: %w", err)
	}

	return nil
}

// GetAll retrieves all gallery items, ordered by created_at descending.
func (m *GalleryRepository) GetAll(ctx context.Context) ([]models.GalleryItem, error) {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	stmt := `
		SELECT id, title, image_link, created_at, updated_at
		FROM gallery
		ORDER BY created_at DESC
	`

	rows, err := m.DB.Query(ctx, stmt)
	if err != nil {
		return nil, fmt.Errorf("failed to query gallery: %w", err)
	}
	defer rows.Close()

	var items []models.GalleryItem

	for rows.Next() {
		var i models.GalleryItem
		err := rows.Scan(
			&i.ID,
			&i.Title,
			&i.ImageLink,
			&i.CreatedAt,
			&i.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan gallery row: %w", err)
		}
		items = append(items, i)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating gallery rows: %w", err)
	}

	return items, nil
}

// GetByID retrieves a single gallery item by ID.
func (m *GalleryRepository) GetByID(ctx context.Context, id int64) (*models.GalleryItem, error) {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	stmt := `
		SELECT id, title, image_link, created_at, updated_at
		FROM gallery
		WHERE id = $1
	`

	var item models.GalleryItem
	err := m.DB.QueryRow(ctx, stmt, id).Scan(
		&item.ID,
		&item.Title,
		&item.ImageLink,
		&item.CreatedAt,
		&item.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("gallery item not found with id: %d", id)
		}
		return nil, fmt.Errorf("failed to get gallery item: %w", err)
	}

	return &item, nil
}

// Delete removes a gallery item from the database.
func (m *GalleryRepository) Delete(ctx context.Context, id int64) error {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	stmt := `DELETE FROM gallery WHERE id = $1`

	cmdTag, err := m.DB.Exec(ctx, stmt, id)
	if err != nil {
		return fmt.Errorf("failed to delete gallery item: %w", err)
	}

	if cmdTag.RowsAffected() == 0 {
		return fmt.Errorf("gallery item with id %d not found", id)
	}

	return nil
}

// Update modifies an existing gallery item's title (general update).
func (m *GalleryRepository) Update(ctx context.Context, item *models.GalleryItem) error {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	stmt := `
		UPDATE gallery
		SET title = $1, updated_at = $2
		WHERE id = $3
	`

	_, err := m.DB.Exec(ctx, stmt,
		item.Title,
		time.Now().UTC(),
		item.ID,
	)

	if err != nil {
		return fmt.Errorf("failed to update gallery item: %w", err)
	}

	return nil
}
