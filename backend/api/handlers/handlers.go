package handlers

import (
	"log"

	"github.com/projuktisheba/ajfses/backend/internal/dbrepo"
	"github.com/projuktisheba/ajfses/backend/internal/models"
)

type HandlerRepo struct {
	JWT      models.JWTConfig
	InfoLog  *log.Logger
	ErrorLog *log.Logger
	Auth     AuthHandler
	Inquiry  InquiryHandler
	Member   MemberHandler
	Team     TeamHandler
	Gallery  GalleryHandler
	Client   ClientHandler
}

func NewHandlerRepo(host string, db *dbrepo.DBRepository, jwt models.JWTConfig, infoLog, errorLog *log.Logger) *HandlerRepo {
	return &HandlerRepo{
		JWT:     jwt,
		InfoLog: infoLog,
		ErrorLog: errorLog,
		Auth:    newAuthHandler(db, jwt, infoLog, errorLog),
		Inquiry: newInquiryHandler(db, infoLog, errorLog),
		Member:  newMemberHandler(db, infoLog, errorLog),
		Team:    newTeamHandler(db, infoLog, errorLog),
		Gallery: newGalleryHandler(db, infoLog, errorLog),
		Client:  newClientHandler(db, infoLog, errorLog),
	}
}
