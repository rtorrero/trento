package services

//go:generate mockery --all

type IUsersService interface {
	AuthenticateByEmailPassword(email string, password string) bool
}

type UsersService struct {
}

func NewUsersService() *UsersService {
	return &UsersService{}
}

func (s *UsersService) AuthenticateByEmailPassword(email string, password string) bool {
	return email == "banana@suse.com" && password == "potato"
}
