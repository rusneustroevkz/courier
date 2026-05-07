package users

import "net/http"

type Controller interface {
	List(w http.ResponseWriter, r *http.Request)
	Get(w http.ResponseWriter, r *http.Request)
}

type controller struct {
	usersService Service
}

func NewController(usersService Service) Controller {
	return &controller{}
}

func (c *controller) List(w http.ResponseWriter, r *http.Request) {

}

func (c *controller) Get(w http.ResponseWriter, r *http.Request) {}
