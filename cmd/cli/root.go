package main

import (
	"context"
	"log"
	"os"
	"time"

	api "github.com/krehermann/go-plugin-prom/api/v1/controller"
	"github.com/krehermann/go-plugin-prom/server"
	"github.com/urfave/cli/v2"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	var pluginName string
	var serverAddr string
	app := &cli.App{
		Commands: []*cli.Command{
			{
				Name:  "plugin",
				Usage: "plugin commands",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:  "name",
						Usage: "plugin name",
						Value: "plugin-0",
					},
					&cli.StringFlag{
						Name:  "addr",
						Usage: "server address",
						Value: "localhost:50051",
					},
				},
				Before: func(cCtx *cli.Context) error {
					pluginName = cCtx.String("name")
					serverAddr = cCtx.String("addr")
					return nil
				},
				Subcommands: []*cli.Command{
					{
						Name:  "start",
						Usage: "start the plugin",
						Action: func(cCtx *cli.Context) error {
							log.Println("Start plugin: ", pluginName)
							// Set up a connection to the server.
							conn, err := grpc.Dial(serverAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
							if err != nil {
								log.Fatalf("did not connect: %v", err)
							}
							defer conn.Close()
							c := api.NewControllerClient(conn)

							// Contact the server and print out its response.
							ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
							defer cancel()
							_, err = c.Start(ctx, &api.StartRequest{Name: pluginName})
							if err != nil {
								log.Fatalf("could not start %s: %v", pluginName, err)
							}
							return nil
						},
					},
					{
						Name:  "stop",
						Usage: "stop the plugin",
						Action: func(cCtx *cli.Context) error {
							log.Println("Stop plugin: ", pluginName)
							// Set up a connection to the server.
							conn, err := grpc.Dial(serverAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
							if err != nil {
								log.Fatalf("did not connect: %v", err)
							}
							defer conn.Close()
							c := api.NewControllerClient(conn)

							// Contact the server and print out its response.
							ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
							defer cancel()
							_, err = c.Stop(ctx, &api.StopRequest{Name: pluginName})
							if err != nil {
								log.Fatalf("could not Stop %s: %v", pluginName, err)
							}
							return nil

						},
					},
					{
						Name:  "kill",
						Usage: "kill the plugin",
						Flags: []cli.Flag{
							&cli.StringFlag{
								Name:  "name",
								Usage: "plugin name",
							},
						},
						Action: func(cCtx *cli.Context) error {
							log.Println("Kill plugin: ", pluginName)
							// Set up a connection to the server.
							conn, err := grpc.Dial(serverAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
							if err != nil {
								log.Fatalf("did not connect: %v", err)
							}
							defer conn.Close()
							c := api.NewControllerClient(conn)

							// Contact the server and print out its response.
							ctx, cancel := context.WithTimeout(context.Background(), time.Second)
							defer cancel()
							_, err = c.Kill(ctx, &api.KillRequest{Name: pluginName})
							if err != nil {
								log.Fatalf("could not Kill %s: %v", pluginName, err)
							}
							return nil
						},
					},
				},
			},

			{
				Name:  "server",
				Usage: "server commands",
				Flags: []cli.Flag{
					&cli.IntFlag{
						Name:  "port",
						Value: 50051,
						Usage: "server grpc port",
					},
					&cli.IntFlag{
						Name:  "prom-port",
						Value: 2112,
						Usage: "server prom http port",
					},
				},

				Subcommands: []*cli.Command{
					{
						Name:  "start",
						Usage: "start the server",
						Flags: []cli.Flag{
							&cli.BoolFlag{
								Name:  "enable-plugins",
								Value: false,
							},
						},
						Action: func(cCtx *cli.Context) error {
							port := cCtx.Int("port")
							promPort := cCtx.Int("prom-port")
							enablePlugins := cCtx.Bool("enable-plugins")
							log.Printf("starting server: %d, plugins enabled: %v", port, enablePlugins)
							op := server.NewOperator(server.OperatorConfig{Port: port, EnablePlugins: enablePlugins, PromPort: promPort})
							return op.Run(context.Background())

						},
					},
				},
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
