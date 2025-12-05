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

type MemberRepository struct {
	DB *pgxpool.Pool
}

// newMemberRepository creates a new instance of the repository.
func newMemberRepository(db *pgxpool.Pool) *MemberRepository {
	return &MemberRepository{DB: db}
}

// Create inserts the basic member info and returns the ID.
func (m *MemberRepository) Create(ctx context.Context, member *models.Member) (int64, error) {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	stmt := `
		INSERT INTO members (name, team, designation, contact, note, image_link, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id
	`

	var id int64
	err := m.DB.QueryRow(ctx, stmt,
		member.Name,
		member.TeamID,
		member.Designation,
		member.Contact,
		member.ImageLink,
		member.Note,
		time.Now().UTC(),
		time.Now().UTC(),
	).Scan(&id)

	if err != nil {
		return 0, fmt.Errorf("failed to insert member: %w", err)
	}

	return id, nil
}

// Update modifies all fields of an existing member, including the image link.
func (m *MemberRepository) Update(ctx context.Context, member *models.Member) error {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	stmt := `
		UPDATE members
		SET name = $1, team = $2, designation=$3, contact = $4, note = $5, image_link = $6, updated_at = $7
		WHERE id = $8
	`

	_, err := m.DB.Exec(ctx, stmt,
		member.Name,
		member.TeamID,
		member.Designation,
		member.Contact,
		member.Note,
		member.ImageLink,
		time.Now().UTC(),
		member.ID,
	)

	if err != nil {
		return fmt.Errorf("failed to update member: %w", err)
	}

	return nil
}

// UpdateImageLink updates only the image_link column for a specific member ID.
func (m *MemberRepository) UpdateImageLink(ctx context.Context, id int64, imagePath string) error {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	stmt := `
		UPDATE members 
		SET image_link = $1, updated_at = $2
		WHERE id = $3
	`

	_, err := m.DB.Exec(ctx, stmt, imagePath, time.Now().UTC(), id)
	if err != nil {
		return fmt.Errorf("failed to update image link: %w", err)
	}

	return nil
}

// Delete removes a member from the database by ID.
func (m *MemberRepository) Delete(ctx context.Context, id int64) error {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	stmt := `DELETE FROM members WHERE id = $1`

	cmdTag, err := m.DB.Exec(ctx, stmt, id)
	if err != nil {
		return fmt.Errorf("failed to delete member: %w", err)
	}

	if cmdTag.RowsAffected() == 0 {
		return fmt.Errorf("member with id %d not found", id)
	}

	return nil
}

// GetByID retrieves a single member by their ID.
func (m *MemberRepository) GetByID(ctx context.Context, id int64) (*models.Member, error) {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	stmt := `
		SELECT  m.id, m.name, m.team, t.title AS team_name, m.designation, m.contact, m.note, m.image_link, m.created_at, m.updated_at
		FROM members AS m
		LEFT JOIN teams AS t ON m.team = t.id
		WHERE m.id = $1
		ORDER BY m.created_at DESC;
	`

	var member models.Member
	err := m.DB.QueryRow(ctx, stmt, id).Scan(
		&member.ID,
		&member.Name,
		&member.TeamID,
		&member.TeamName,
		&member.Designation,
		&member.Contact,
		&member.Note,
		&member.ImageLink,
		&member.CreatedAt,
		&member.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("member not found with id: %d", id)
		}
		return nil, fmt.Errorf("failed to get member: %w", err)
	}

	return &member, nil
}

// GetAll retrieves all members, ordered by created_at descending.
func (m *MemberRepository) GetAll(ctx context.Context) ([]*models.Member, error) {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	stmt := `
		SELECT  m.id, m.name, m.team, t.title AS team_name, m.designation, m.contact, m.note, m.image_link, m.created_at, m.updated_at
		FROM members AS m
		LEFT JOIN teams AS t ON m.team = t.id
		ORDER BY m.created_at DESC;
	`

	rows, err := m.DB.Query(ctx, stmt)
	if err != nil {
		return nil, fmt.Errorf("failed to query members: %w", err)
	}
	defer rows.Close()

	var members []*models.Member

	for rows.Next() {
		var member models.Member
		err := rows.Scan(
			&member.ID,
			&member.Name,
			&member.TeamID,
			&member.TeamName,
			&member.Designation,
			&member.Contact,
			&member.Note,
			&member.ImageLink,
			&member.CreatedAt,
			&member.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan member row: %w", err)
		}
		members = append(members, &member)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating member rows: %w", err)
	}

	return members, nil
}
