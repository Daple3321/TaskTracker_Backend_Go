package entity

import "time"

type User struct {
	Id           int
	Username     string
	PasswordHash string
	CreatedAt    time.Time
}
