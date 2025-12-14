package models

import "time"

// Client struct corresponds to the 'clients' database table.
type Client struct {
	ID          int64     `json:"id"`
	Name        string    `json:"name"`
	Area        string    `json:"area"`
	ServiceName string    `json:"service_name"`
	ServiceDate string    `json:"service_date"`
	Status      string    `json:"status"` // Can be used for filtering (e.g., "Running", "Completed")
	Note        string    `json:"note"`
	ImageLink   string    `json:"image_link"` // The filename/path on the server
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// ClientMetrics holds the statistical counts for the client data.
type ClientMetrics struct {
	TotalDistinctClients int64 `json:"total_clients"`
	ActiveProjects       int64 `json:"active_projects"`
	CompletedProjects    int64 `json:"completed_projects"`
	TotalEmployees       int64 `json:"total_employees"`
}
