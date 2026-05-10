package users

import (
	"net/http"
)

type Controller interface {
	GetMe(w http.ResponseWriter, r *http.Request)
}

type controller struct {
	usersService Service
}

func NewController(usersService Service) Controller {
	return &controller{
		usersService: usersService,
	}
}

func (c *controller) GetMe(w http.ResponseWriter, r *http.Request) {

}
