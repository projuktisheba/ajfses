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

type TeamRepository struct {
	DB *pgxpool.Pool
}
// newTeamRepository creates a new instance of the repository.
func newTeamRepository(db *pgxpool.Pool) *TeamRepository {
	return &TeamRepository{DB: db}
}

// Create inserts a new team into the database.
func (m *TeamRepository) Create(ctx context.Context, team *models.Team) (int64, error) {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	stmt := `
		INSERT INTO teams (title, created_at, updated_at)
		VALUES ($1, $2, $3)
		RETURNING id
	`

	var id int64
	err := m.DB.QueryRow(ctx, stmt,
		team.Title,
		time.Now().UTC(),
		time.Now().UTC(),
	).Scan(&id)

	if err != nil {
		return 0, fmt.Errorf("failed to insert team: %w", err)
	}

	return id, nil
}

// GetByID retrieves a single team by ID.
func (m *TeamRepository) GetByID(ctx context.Context, id int64) (*models.Team, error) {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	stmt := `
		SELECT id, title, created_at, updated_at
		FROM teams
		WHERE id = $1
	`

	var team models.Team
	err := m.DB.QueryRow(ctx, stmt, id).Scan(
		&team.ID,
		&team.Title,
		&team.CreatedAt,
		&team.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("team not found with id: %d", id)
		}
		return nil, fmt.Errorf("failed to get team: %w", err)
	}

	return &team, nil
}

// GetAll retrieves all teams, ordered by created_at descending.
func (m *TeamRepository) GetAll(ctx context.Context) ([]models.Team, error) {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	stmt := `
		SELECT id, title, created_at, updated_at
		FROM teams
		ORDER BY created_at DESC
	`

	rows, err := m.DB.Query(ctx, stmt)
	if err != nil {
		return nil, fmt.Errorf("failed to query teams: %w", err)
	}
	defer rows.Close()

	var teams []models.Team

	for rows.Next() {
		var team models.Team
		err := rows.Scan(
			&team.ID,
			&team.Title,
			&team.CreatedAt,
			&team.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan team row: %w", err)
		}
		teams = append(teams, team)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating team rows: %w", err)
	}

	return teams, nil
}

// Update modifies an existing team.
func (m *TeamRepository) Update(ctx context.Context, team *models.Team) error {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	stmt := `
		UPDATE teams
		SET title = $1, updated_at = $2
		WHERE id = $3
	`

	_, err := m.DB.Exec(ctx, stmt,
		team.Title,
		time.Now().UTC(),
		team.ID,
	)

	if err != nil {
		return fmt.Errorf("failed to update team: %w", err)
	}

	return nil
}

// Delete removes a team from the database.
func (m *TeamRepository) Delete(ctx context.Context, id int64) error {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	stmt := `DELETE FROM teams WHERE id = $1`

	cmdTag, err := m.DB.Exec(ctx, stmt, id)
	if err != nil {
		return fmt.Errorf("failed to delete team: %w", err)
	}

	if cmdTag.RowsAffected() == 0 {
		return fmt.Errorf("team with id %d not found", id)
	}

	return nil
}

func (m *MemberRepository) GetTeamsWithMembers(ctx context.Context) ([]*models.TeamData, error) {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	// Use COALESCE to handle NULLs because of the LEFT JOIN
	// If a team has no members, member columns will be NULL. We convert them to defaults.
	query := `
		SELECT 
			t.id AS team_id,
			t.title AS team_name,
			COALESCE(m.id, 0),
			COALESCE(m.name, ''),
			COALESCE(m.designation, ''),
			COALESCE(m.contact, ''),
			COALESCE(m.note, ''),
			COALESCE(m.image_link, ''),
			COALESCE(m.created_at, CURRENT_TIMESTAMP),
			COALESCE(m.updated_at, CURRENT_TIMESTAMP)
		FROM teams t
		LEFT JOIN members m ON t.id = m.team
		ORDER BY t.id ASC, m.id ASC; -- Order by Team first to group them
	`

	rows, err := m.DB.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query teams and members: %w", err)
	}
	defer rows.Close()

	// Map to keep track of pointers to TeamData by TeamID
	teamMap := make(map[int64]*models.TeamData)
	var teams []*models.TeamData

	for rows.Next() {
		var tID int64
		var tName string
		// Temp member variables
		var mID int64
		var mName, mDesignation, mContact, mNote, mImg string
		var mCreated, mUpdated time.Time

		err := rows.Scan(
			&tID, &tName,
			&mID, &mName, &mDesignation, &mContact, &mNote, &mImg, &mCreated, &mUpdated,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}

		// Check if we already processed this team
		teamData, exists := teamMap[tID]
		if !exists {
			// Create new Team entry
			teamData = &models.TeamData{
				TeamID:   tID,
				TeamName: tName,
				Members:  []*models.Member{}, // Initialize empty slice
			}
			teamMap[tID] = teamData
			teams = append(teams, teamData) // Add to result slice
		}

		// If mID is > 0, it means there is a valid member in this row
		if mID > 0 {
			member := &models.Member{
				ID:          mID,
				Name:        mName,
				TeamID:      tID,
				Designation: mDesignation,
				Contact:     mContact,
				Note:        mNote,
				ImageLink:   mImg,
				CreatedAt:   mCreated,
				UpdatedAt:   mUpdated,
			}
			teamData.Members = append(teamData.Members, member)
		}
	}

	return teams, nil
}