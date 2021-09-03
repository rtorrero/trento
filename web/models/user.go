package models

type User struct {
	Email    string `gorm:"type:varchar(40);primaryKey" json:"username,omitempty"`
	Password string `gorm:"size:255" json:"password,omitempty"`
}
