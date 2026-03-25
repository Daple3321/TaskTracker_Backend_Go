package entity

import "time"

type Task struct {
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Tags        []Tag     `json:"tags"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
	Id          int       `json:"id"`
	UserId      int       `json:"userId"`
}

type PaginatedResponse struct {
	Items      []Task `json:"tasks"`
	Page       int    `json:"page"`
	Limit      int    `json:"limit"`
	TotalItems int    `json:"total_items"`
	TotalPages int    `json:"total_pages"`
}
