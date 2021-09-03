package services

import (
	"github.com/trento-project/trento/web/models"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

//go:generate mockery --all

type UsersService interface {
	AuthenticateByEmailPassword(email string, password string) bool
	CreateUserByEmailPassword(email string, password string)
}

type usersService struct {
	db *gorm.DB
}

func NewUsersService(db *gorm.DB) *usersService {
	return &usersService{db: db}
}

func (s *usersService) AuthenticateByEmailPassword(email string, password string) bool {
	var user models.User
	s.db.Where(&models.User{Email: email}).First(&user)

	err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))

	return err == nil
}

func (s *usersService) CreateUserByEmailPassword(email string, password string) {
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)

	s.db.Create(&models.User{Email: email, Password: string(hashedPassword)})
}
