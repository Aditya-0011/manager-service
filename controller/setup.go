package controller

import (
	"manager/controller/portfolio"
	"manager/controller/user"
	"manager/db"

	"github.com/Aditya-0011/common/contracts/go/manager"
	"google.golang.org/grpc"
)

func Setup(server *grpc.Server, database *db.DatabaseParams) {
	manager.RegisterUserServiceServer(server, user.NewUserServer(user.UserServerParams{
		Postgres: database.Postgres,
	}))

	manager.RegisterPortfolioServiceServer(server, portfolio.NewPortfolioServer(portfolio.PortfolioServerParams{
		Postgres: database.Postgres,
	}))
}
