package model_auth

import "time"

type Entity struct {
	ID        string
	Name      string
	Status    string
	Email     string
	Password  string
	Created   time.Time
	UpdatedAt time.Time
}

type LoginEntity struct {
	ID       string
	Email    string
	Password string
	Status   string
}
