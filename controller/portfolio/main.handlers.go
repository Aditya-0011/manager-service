package portfolio

import (
	"manager/db"

	manager "github.com/Aditya-0011/common/contracts/go/manager"
)

type (
	portfolioServer struct {
		manager.UnimplementedPortfolioServiceServer
		postgres *db.PostgresParams
	}

	PortfolioServerParams struct {
		Postgres *db.PostgresParams
	}
)

func NewPortfolioServer(params PortfolioServerParams) *portfolioServer {
	return &portfolioServer{
		postgres: params.Postgres,
	}
}
