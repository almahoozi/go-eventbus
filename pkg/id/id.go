package id

import "github.com/google/uuid"

func New() string {
	return uuid.New().String()
}
