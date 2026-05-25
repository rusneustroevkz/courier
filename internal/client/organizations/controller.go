package organizations

type Controller interface{}

type controller struct {
	organizationsService Service
}

func NewController(organizationsService Service) Controller {
	return &controller{
		organizationsService: organizationsService,
	}
}
