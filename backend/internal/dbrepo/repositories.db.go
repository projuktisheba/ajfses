package dbrepo

import (
	"github.com/jackc/pgx/v5/pgxpool"
)

// DBRepository contains all individual repositories
type DBRepository struct {
	UserRepo    *UserRepo
	InquiryRepo *InquiryRepository
}

// NewDBRepository initializes all repositories with a shared connection pool
func NewDBRepository(db *pgxpool.Pool) *DBRepository {
	return &DBRepository{
		UserRepo:    newUserRepo(db),
		InquiryRepo: newInquiryRepository(db),
	}
}
