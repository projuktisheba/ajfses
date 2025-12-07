package handlers

import (
	"log"

	"github.com/projuktisheba/ajfses/backend/internal/dbrepo"
	"github.com/projuktisheba/ajfses/backend/internal/models"
)

type HandlerRepo struct {
	Auth    AuthHandler
	Inquiry InquiryHandler
	Member  MemberHandler
	Team    TeamHandler
	Gallery GalleryHandler
}

func NewHandlerRepo(host string, db *dbrepo.DBRepository, JWT models.JWTConfig, infoLog, errorLog *log.Logger) *HandlerRepo {
	return &HandlerRepo{
		Auth:    newAuthHandler(db, JWT, infoLog, errorLog),
		Inquiry: newInquiryHandler(db, infoLog, errorLog),
		Member:  newMemberHandler(db, infoLog, errorLog),
		Team:    newTeamHandler(db, infoLog, errorLog),
		Gallery: newGalleryHandler(db, infoLog, errorLog),
	}
}
