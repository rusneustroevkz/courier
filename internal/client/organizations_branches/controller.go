package organizations_branches

import "net/http"

type Controller interface {
	Create(w http.ResponseWriter, r *http.Request)
	List(w http.ResponseWriter, r *http.Request)
	GetByID(w http.ResponseWriter, r *http.Request)
	Suggest(w http.ResponseWriter, r *http.Request)
}

type controller struct {
	service Service
}

func NewController(service Service) Controller {
	return &controller{
		service: service,
	}
}

type CreateRequest struct {
}

func (c *controller) Create(w http.ResponseWriter, r *http.Request) {

}

func (c *controller) List(w http.ResponseWriter, r *http.Request) {}

func (c *controller) GetByID(w http.ResponseWriter, r *http.Request) {}

func (c *controller) Suggest(w http.ResponseWriter, r *http.Request) {}
