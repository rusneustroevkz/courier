package users

import (
	"net/http"
)

type Controller interface {
	GetByID(w http.ResponseWriter, r *http.Request)
}

type controller struct {
	usersService Service
}

func NewController(usersService Service) Controller {
	return &controller{
		usersService: usersService,
	}
}

func (c *controller) GetByID(w http.ResponseWriter, r *http.Request) {}
