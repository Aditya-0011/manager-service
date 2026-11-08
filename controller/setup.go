package controller

import (
	"manager/controller/portfolio"
	"manager/controller/user"

	"github.com/Aditya-0011/common/contracts/go/manager"
	"github.com/jackc/pgx/v5/pgxpool"
	"google.golang.org/grpc"
)

func Register(server grpc.ServiceRegistrar, pool *pgxpool.Pool) {
	manager.RegisterUserServiceServer(server, user.NewServer(pool))
	manager.RegisterPortfolioServiceServer(server, portfolio.NewServer(pool))
}
