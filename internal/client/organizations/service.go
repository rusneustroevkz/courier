package organizations

type Service interface{}

type service struct {
	organizationsRepo Querier
}

func NewService(organizationsRepo Querier) Service {
	return &service{
		organizationsRepo: organizationsRepo,
	}
}
