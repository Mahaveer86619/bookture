package models

import "gorm.io/gorm"

type User struct {
	gorm.Model

	Email        string `gorm:"uniqueIndex"`
	PasswordHash string
	DisplayName  string

	Libraries []Library `gorm:"constraint:OnDelete:CASCADE;"`
}

type Library struct {
	gorm.Model

	UserID uint `gorm:"index"`
	Name   string

	Books []Book `gorm:"constraint:OnDelete:CASCADE;"`
}
