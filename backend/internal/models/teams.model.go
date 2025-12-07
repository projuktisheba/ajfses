package models

import "time"

type Team struct {
	ID        int64     `json:"id"`
	Title     string    `json:"title"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// TeamData represents a team and its list of members
type TeamData struct {
	TeamID   int64     `json:"team_id"`
	TeamName string    `json:"team_name"`
	Members  []*Member `json:"members"`
}
