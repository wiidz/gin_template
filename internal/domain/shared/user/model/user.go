package model

type User struct {
	ID           uint64
	LoginID      string
	Nickname     string
	PasswordHash string
}
