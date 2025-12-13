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

// --- ClientRepository and Constructor ---

// ClientRepository holds the database pool connection for client operations.
type ClientRepository struct {
	DB *pgxpool.Pool
}

// newClientRepository creates a new instance of the repository.
func newClientRepository(db *pgxpool.Pool) *ClientRepository {
	return &ClientRepository{DB: db}
}

// --- CRUD Operations ---

// Create inserts the basic client info and returns the ID.
// Note: This assumes models.Client has fields: Name, Area, ServiceName, Status, Note, ImageLink, CreatedAt, UpdatedAt
func (c *ClientRepository) Create(ctx context.Context, client *models.Client) (int64, error) {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	stmt := `
		INSERT INTO clients (name, area, service_name, status, note, image_link, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id
	`

	var id int64
	err := c.DB.QueryRow(ctx, stmt,
		client.Name,
		client.Area,
		client.ServiceName,
		client.Status,
		client.Note,
		client.ImageLink,
		time.Now().UTC(),
		time.Now().UTC(),
	).Scan(&id)

	if err != nil {
		return 0, fmt.Errorf("failed to insert client: %w", err)
	}

	return id, nil
}

// Update modifies all updatable fields of an existing client, including the image link.
func (c *ClientRepository) Update(ctx context.Context, client *models.Client) error {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	stmt := `
		UPDATE clients
		SET name = $1, area = $2, service_name = $3, status = $4, note = $5, image_link = $6, updated_at = $7
		WHERE id = $8
	`

	_, err := c.DB.Exec(ctx, stmt,
		client.Name,
		client.Area,
		client.ServiceName,
		client.Status,
		client.Note,
		client.ImageLink,
		time.Now().UTC(),
		client.ID,
	)

	if err != nil {
		return fmt.Errorf("failed to update client: %w", err)
	}

	return nil
}

// UpdateImageLink updates only the image_link column for a specific client ID.
func (c *ClientRepository) UpdateImageLink(ctx context.Context, id int64, imagePath string) error {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	stmt := `
		UPDATE clients 
		SET image_link = $1, updated_at = $2
		WHERE id = $3
	`

	_, err := c.DB.Exec(ctx, stmt, imagePath, time.Now().UTC(), id)
	if err != nil {
		return fmt.Errorf("failed to update image link for client: %w", err)
	}

	return nil
}

// Delete removes a client from the database by ID.
func (c *ClientRepository) Delete(ctx context.Context, id int64) error {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	stmt := `DELETE FROM clients WHERE id = $1`

	cmdTag, err := c.DB.Exec(ctx, stmt, id)
	if err != nil {
		return fmt.Errorf("failed to delete client: %w", err)
	}

	if cmdTag.RowsAffected() == 0 {
		return fmt.Errorf("client with id %d not found", id)
	}

	return nil
}

// GetByID retrieves a single client by their ID.
func (c *ClientRepository) GetByID(ctx context.Context, id int64) (*models.Client, error) {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	stmt := `
		SELECT id, name, area, service_name, status, note, image_link, created_at, updated_at
		FROM clients
		WHERE id = $1
	`

	var client models.Client
	err := c.DB.QueryRow(ctx, stmt, id).Scan(
		&client.ID,
		&client.Name,
		&client.Area,
		&client.ServiceName,
		&client.Status,
		&client.Note,
		&client.ImageLink,
		&client.CreatedAt,
		&client.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("client not found with id: %d", id)
		}
		return nil, fmt.Errorf("failed to get client: %w", err)
	}

	return &client, nil
}

// GetAll retrieves all clients, optionally filtered by status, ordered by created_at descending.
// The optional 'status' parameter replaces the 'designation' filter from the original.
func (c *ClientRepository) GetAll(ctx context.Context, status string) ([]*models.Client, error) {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()
	
	// Prepare the WHERE clause based on the status parameter
	whClause := ""
	args := []any{}
	
	if status != "" {
		whClause = "WHERE status = $1"
		args = append(args, status)
	}
	
	stmt := fmt.Sprintf(`
		SELECT id, name, area, service_name, status, note, image_link, created_at, updated_at
		FROM clients
		%s 
		ORDER BY created_at DESC;
	`, whClause)

	var clients []*models.Client
	rows, err := c.DB.Query(ctx, stmt, args...)
	if err != nil {
		return clients, fmt.Errorf("failed to query clients: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var client models.Client
		err := rows.Scan(
			&client.ID,
			&client.Name,
			&client.Area,
			&client.ServiceName,
			&client.Status,
			&client.Note,
			&client.ImageLink,
			&client.CreatedAt,
			&client.UpdatedAt,
		)
		if err != nil {
			return clients, fmt.Errorf("failed to scan client row: %w", err)
		}
		clients = append(clients, &client)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating client rows: %w", err)
	}

	return clients, nil
}