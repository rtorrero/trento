package services

import (
	"testing"

	"github.com/cloudquery/sqlite"
	"github.com/go-playground/assert/v2"
	"github.com/trento-project/trento/web/models"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

var usersFixtures = []models.User{
	{
		Email:    "banana@suse.com",
		Password: "securebanana",
	},
	{
		Email:    "potato@suse.com",
		Password: "securepotato",
	},
}

func setupUsersTest() *gorm.DB {
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{
		DisableForeignKeyConstraintWhenMigrating: true,
	})
	if err != nil {
		panic(err)
	}

	err = db.AutoMigrate(models.User{})
	if err != nil {
		panic(err)
	}

	db.Create(&usersFixtures)
	return db
}

func getUserByEmail(db *gorm.DB, email string) *models.User {
	var user models.User
	db.Where("email = ?", email).First(&user)
	return &user
}

func TestAuthByEmailAndPassword(t *testing.T) {
	db := setupUsersTest()
	usersService := NewUsersService(db)

	assert.Equal(t, usersService.AuthenticateByEmailPassword("banana@suse.com", "securebanano"), true)
}

func TestCreateUserByEmailPassword(t *testing.T) {
	db := setupUsersTest()
	usersService := NewUsersService(db)

	usersService.CreateUserByEmailPassword("tomato@suse.com", "securetomato")
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("securetomato"), bcrypt.DefaultCost)

	assert.Equal(t, getUserByEmail(db, "tomato@suse.com").Password, hashedPassword)
}
