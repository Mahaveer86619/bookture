package models

type User struct {
	Base
	Username     string `gorm:"uniqueIndex;not null"`
	Email        string `gorm:"uniqueIndex;not null"`
	PasswordHash string 
	DisplayName  string
	Bio          string
	AvatarURL    string

	// Relations
	Books     []Book `gorm:"foreignKey:UserID"`
	Followers []User `gorm:"many2many:follows;joinForeignKey:following_id;joinReferences:follower_id"`
	Following []User `gorm:"many2many:follows;joinForeignKey:follower_id;joinReferences:following_id"`
}

func (u *User) TableName() string {
	return "users"
}
