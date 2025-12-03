package models

import "time"

// Inquiry represents the data structure for the inquiries table.
type Inquiry struct {
	ID          int64     `json:"id"`
	InquiryDate time.Time `json:"inquiry_date"`
	Name        string    `json:"name"`
	Contact     string    `json:"contact"`
	Subject     string    `json:"subject"`
	Message     string    `json:"message"`
	Status      string    `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}
