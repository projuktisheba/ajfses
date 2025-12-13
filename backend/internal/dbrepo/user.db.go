package dbrepo

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/projuktisheba/ajfses/backend/internal/models"
)

// ============================== User Repository ==============================
type UserRepo struct {
	db *pgxpool.Pool
}

func newUserRepo(db *pgxpool.Pool) *UserRepo {
	return &UserRepo{db: db}
}

// CreateUser inserts a new user
func (r *UserRepo) CreateUser(ctx context.Context, e *models.User) error {
	query := `
		INSERT INTO users 
		(name, role, status, mobile, email, password, address, avatar_link, created_at, updated_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,CURRENT_TIMESTAMP,CURRENT_TIMESTAMP)
		RETURNING id, created_at, updated_at
	`

	row := r.db.QueryRow(ctx, query,
		e.Name, e.Role, e.Status, e.Mobile, e.Email, e.Password,
		e.Address, e.AvatarLink,
	)

	err := row.Scan(&e.ID, &e.CreatedAt, &e.UpdatedAt)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" { // unique_violation
			switch pgErr.ConstraintName {
			case "users_mobile_key":
				return errors.New("this mobile is already associated with another account")
			case "users_email_key":
				return errors.New("this email is already associated with another account")
			}
		}
		if err == pgx.ErrNoRows {
			return errors.New("failed to insert user")
		}
		return err
	}

	return nil
}

// GetUserByID fetches a user by ID
func (r *UserRepo) GetUserByID(ctx context.Context, id int64) (*models.User, error) {
	query := `
		SELECT id, name, role, status, mobile, email, password, address, avatar_link, joining_date, created_at, updated_at
		FROM users WHERE id = $1
	`
	e := &models.User{}
	err := r.db.QueryRow(ctx, query, id).Scan(
		&e.ID, &e.Name, &e.Role, &e.Status, &e.Mobile, &e.Email,
		&e.Password, &e.Address, &e.AvatarLink, &e.JoiningDate,
		&e.CreatedAt, &e.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, errors.New("no user found")
		}
		return nil, err
	}
	return e, nil
}

// GetUserByUsername fetches a user by mobile or email
func (r *UserRepo) GetUserByUsername(ctx context.Context, username string) (*models.User, error) {
	query := `
		SELECT id, name, role, status, mobile, email, password, address, avatar_link, joining_date, created_at, updated_at
		FROM users
		WHERE mobile = $1 OR email = $1
		LIMIT 1
	`
	e := &models.User{}
	err := r.db.QueryRow(ctx, query, username).Scan(
		&e.ID, &e.Name, &e.Role, &e.Status, &e.Mobile, &e.Email,
		&e.Password, &e.Address, &e.AvatarLink, &e.JoiningDate,
		&e.CreatedAt, &e.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, errors.New("no user found")
		}
		return nil, err
	}
	return e, nil
}

// UpdateUser updates user details
func (r *UserRepo) UpdateUser(ctx context.Context, e *models.User) error {
	query := `
		UPDATE users
		SET 
			name = $2,
			role = $3,
			status = $4,
			mobile = $5,
			email = $6,
			address = $7,
			avatar_link = $8,
			updated_at = CURRENT_TIMESTAMP
		WHERE id = $1
		RETURNING updated_at
	`

	row := r.db.QueryRow(ctx, query,
		e.ID, e.Name, e.Role, e.Status, e.Mobile, e.Email,
		e.Address, e.AvatarLink,
	)

	err := row.Scan(&e.UpdatedAt)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			if pgErr.Code == "23505" {
				switch pgErr.ConstraintName {
				case "users_mobile_key":
					return errors.New("this mobile is already associated with another user")
				case "users_email_key":
					return errors.New("this email is already associated with another user")
				default:
					return fmt.Errorf("unique constraint violation: %s", pgErr.Message)
				}
			}
			return fmt.Errorf("database error: %s", pgErr.Message)
		}
		if err == pgx.ErrNoRows {
			return errors.New("no user found with the given id")
		}
		return err
	}

	return nil
}
// UpdatePassword updates only the password hash for a specific user ID.
func (r *UserRepo) UpdatePassword(ctx context.Context, userID int64, hashedPassword string) error {
	const sqlUpdatePassword = `
        UPDATE users SET
            password = $2,
            updated_at = CURRENT_TIMESTAMP
        WHERE id = $1
    `
	
	_, err :=  r.db.Exec(ctx, sqlUpdatePassword, userID, hashedPassword)
	
	if err != nil {
		return fmt.Errorf("failed to update password: %w", err)
	}

	return nil
}
// UpdateUserAvatarLink updates only the user's avatar link
func (r *UserRepo) UpdateUserAvatarLink(ctx context.Context, id int64, avatarLink string) error {
	query := `
		UPDATE users
		SET avatar_link=$1, updated_at=CURRENT_TIMESTAMP
		WHERE id=$2
	`
	_, err := r.db.Exec(ctx, query, avatarLink, id)
	return err
}

// UpdateUserStatus updates user role and status
func (r *UserRepo) UpdateUserStatus(ctx context.Context, id int64, role, status string) error {
	query := `
		UPDATE users
		SET role=$1, status=$2, updated_at=CURRENT_TIMESTAMP
		WHERE id=$3
	`
	_, err := r.db.Exec(ctx, query, role, status, id)
	return err
}

// PaginatedUserList returns paginated list of users with optional filters
func (r *UserRepo) PaginatedUserList(ctx context.Context, page, limit int, role, status, sortBy, sortOrder string) ([]*models.User, int, error) {
	query := `
		SELECT id, name, role, status, mobile, email, address, avatar_link, joining_date, created_at, updated_at
		FROM users
		WHERE 1=1
	`
	countQuery := `SELECT COUNT(*) FROM users WHERE 1=1`
	args := []interface{}{}
	countArgs := []interface{}{}
	argIdx := 1

	// Filtering
	if role != "" {
		query += fmt.Sprintf(" AND role = $%d", argIdx)
		countQuery += fmt.Sprintf(" AND role = $%d", argIdx)
		args = append(args, role)
		countArgs = append(countArgs, role)
		argIdx++
	}
	if status != "" {
		query += fmt.Sprintf(" AND status = $%d", argIdx)
		countQuery += fmt.Sprintf(" AND status = $%d", argIdx)
		args = append(args, status)
		countArgs = append(countArgs, status)
		argIdx++
	}

	// Sorting
	if sortBy == "" {
		sortBy = "created_at"
	}
	if sortOrder != "ASC" && sortOrder != "DESC" {
		sortOrder = "DESC"
	}
	query += fmt.Sprintf(" ORDER BY %s %s", sortBy, sortOrder)

	// Pagination
	if page > 0 && limit > 0 {
		offset := (page - 1) * limit
		query += fmt.Sprintf(" LIMIT $%d OFFSET $%d", argIdx, argIdx+1)
		args = append(args, limit, offset)
	}

	// Count total
	var total int
	if err := r.db.QueryRow(ctx, countQuery, countArgs...).Scan(&total); err != nil {
		return nil, 0, err
	}

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var users []*models.User
	for rows.Next() {
		var u models.User
		if err := rows.Scan(
			&u.ID, &u.Name, &u.Role, &u.Status, &u.Mobile, &u.Email,
			&u.Address, &u.AvatarLink, &u.JoiningDate, &u.CreatedAt, &u.UpdatedAt,
		); err != nil {
			return nil, 0, err
		}
		users = append(users, &u)
	}

	return users, total, nil
}
