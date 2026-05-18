package middlewares

type Auth interface {
	VerifyToken(tokenString string) (int64, error)
}
