package models

import "time"

type Member struct {
	ID             int64     `json:"id"`
	Name           string    `json:"name"`
	TeamID         int64     `json:"team_id"`
	TeamName       string    `json:"team_name"`
	Designation    string    `json:"designation"`
	Contact        string    `json:"contact"`
	Note           string    `json:"note"`
	ImageLink      string    `json:"image_link"`
	ShowOnHomepage bool      `json:"show_on_homepage"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}
