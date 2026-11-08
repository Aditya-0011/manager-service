package portfolio

import (
	"github.com/Aditya-0011/common/contracts/go/manager"
	"github.com/jackc/pgx/v5/pgxpool"
)

type portfolioServer struct {
	manager.UnimplementedPortfolioServiceServer
	pool *pgxpool.Pool
}

func NewServer(pool *pgxpool.Pool) manager.PortfolioServiceServer {
	return &portfolioServer{pool: pool}
}
