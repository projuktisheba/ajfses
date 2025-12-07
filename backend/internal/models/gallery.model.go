package models

import "time"

type GalleryItem struct {
	ID        int64     `json:"id"`
	Title     string    `json:"title"`
	ImageLink string    `json:"image_link"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
