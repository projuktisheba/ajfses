package dbrepo

import (
	"github.com/jackc/pgx/v5/pgxpool"
)

// DBRepository contains all individual repositories
type DBRepository struct {
	UserRepo    *UserRepo
	InquiryRepo *InquiryRepository
	MemberRepo  *MemberRepository
	TeamRepo  *TeamRepository
	GalleryRepo  *GalleryRepository
	ClientRepo  *ClientRepository
}

// NewDBRepository initializes all repositories with a shared connection pool
func NewDBRepository(db *pgxpool.Pool) *DBRepository {
	return &DBRepository{
		UserRepo:    newUserRepo(db),
		InquiryRepo: newInquiryRepository(db),
		MemberRepo:  newMemberRepository(db),
		TeamRepo:  newTeamRepository(db),
		GalleryRepo:  newGalleryRepository(db),
		ClientRepo:  newClientRepository(db),
	}
}
