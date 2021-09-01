package services

//go:generate mockery --all

type UsersService interface {
	AuthenticateByEmailPassword(email string, password string) bool
}

type usersService struct {
}

func NewUsersService() *usersService {
	return &usersService{}
}

func (s *usersService) AuthenticateByEmailPassword(email string, password string) bool {
	return email == "banana@suse.com" && password == "potato"
}
