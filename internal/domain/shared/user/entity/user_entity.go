package entity

import "time"

type UserEntity struct {
	ID           uint64 `gorm:"primaryKey;autoIncrement"`
	LoginID      string `gorm:"uniqueIndex;size:128;not null"`
	Nickname     string `gorm:"size:128"`
	PasswordHash string `gorm:"size:256;not null"`
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

func EntitiesForMigrate() []interface{} {
	return []interface{}{&UserEntity{}}
}
