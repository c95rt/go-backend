package main

import (
	"log"
	"os"
	"time"

	"bitbucket.org/parqueoasis/backend/api"
	"bitbucket.org/parqueoasis/backend/server"
	"github.com/joho/godotenv"
	"github.com/urfave/cli"
)

// @title backend API
// @version 0.1
// @description Api for all payment logic.

// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html

// @host api.parqueoasis.cl
// @BasePath /
// @schemes http https

// @securityDefinitions.apiKey ApiKeyAuth
// @in header
// @name Authorization

func main() {
	_ = godotenv.Load("dev.env")

	app := cli.NewApp()
	app.Name = "Go Auth Service"
	app.Version = "1.00"
	app.Compiled = time.Now()
	app.Authors = []cli.Author{
		{
			Name:  "César Reyes",
			Email: "cesar95rt@gmail.com",
		},
	}
	app.Copyright = "(c) Routeland CORP"
	app.Commands = []cli.Command{
		{
			Name:  "backend-up",
			Usage: "This command starts the backend service",
			Action: func(c *cli.Context) error {
				StartServer(api.GetRoutes())
				return nil
			},
		},
	}
	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}

func StartServer(routes []*server.Route) {
	ctx := server.GetAppContext()
	ctx.CreateMySQLConnection()
	ctx.CreateSMTPConnection()
	ctx.CreateMercadoPagoIntegration()
	ctx.CreateNewSessionS3()

	server.UpServer(routes, ctx)
}
