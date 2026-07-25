package model

import (
	"time"

	"gorm.io/gorm"
)

/*
CREATE UNIQUE INDEX idx_users_email_unique
ON users(email)
WHERE deleted_at IS NULL;
*/

type User struct {
	ID        uint    `gorm:"primaryKey;autoIncrement"`
	Name      string  `gorm:"type:varchar(255);not null"`
	Email     string  `gorm:"type:varchar(255);not null;index:idx_users_email_unique,unique,where:deleted_at IS NULL"`
	Password  string  `gorm:"type:varchar(255);not null"`
	Phone     *string `gorm:"type:varchar(17)"`
	Photo     *string `gorm:"type:varchar(255)"`
	Address   *string `gorm:"type:text"`
	CreatedAt time.Time
	UpdatedAt *time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`
}

func (User) TableName() string {
	return "users"
}
