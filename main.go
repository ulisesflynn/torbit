package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/ulisesflynn/torbit/server"
	"gopkg.in/urfave/cli.v1"
	"gopkg.in/urfave/cli.v1/altsrc"
)

func main() {
	log.Println("Starting chat server...")
	app := cli.NewApp()
	app.Name = "chat server"
	app.Usage = "simple telnet chat server"
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:   "config-file",
			Value:  "config.toml",
			Usage:  "path to config file",
			EnvVar: "CFG_FILE",
		},
		altsrc.NewStringFlag(cli.StringFlag{
			Name:   "log-dir",
			Value:  "/tmp",
			Usage:  "path to server chat logs",
			EnvVar: "CHAT_LOG_DIR",
		}),
		altsrc.NewStringFlag(cli.StringFlag{
			Name:   "chat-port",
			Value:  "2000",
			Usage:  "port chat server listens to",
			EnvVar: "CHAT_PORT",
		}),
		altsrc.NewStringFlag(cli.StringFlag{
			Name:   "http-port",
			Value:  "8080",
			Usage:  "port http server listens to",
			EnvVar: "HTTP_PORT",
		}),
		altsrc.NewStringFlag(cli.StringFlag{
			Name:   "server-address",
			Value:  "127.0.0.1",
			Usage:  "server address",
			EnvVar: "SRV_ADDR",
		}),
		altsrc.NewIntFlag(cli.IntFlag{
			Name:   "max-http-body",
			Value:  1024, /// bytes
			Usage:  "max number of bytes for incoming http body",
			EnvVar: "MAX-HTTP-BODY",
		}),
	}

	app.Action = func(c *cli.Context) error {
		// create log file
		const timeFormat = "2006-01-02T15-04-05"
		logName := "client-log-" + time.Now().UTC().Format(timeFormat) + ".log"
		logPath := filepath.Join(c.String("log-dir"), logName)
		logFile, err := os.OpenFile(logPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
		if err != nil {
			panic(fmt.Sprintf("Failed to create chat log file: %s, error: %s", logPath, err))
		}
		defer logFile.Close()
		// create and run server instance
		s := server.New(logFile, c.String("chat-port"), c.String("http-port"), c.String("server-address"), c.Int("max-http-body"))
		log.Println("Chat server started")
		s.Run()
		// handle application closing
		closeChan := make(chan os.Signal, 1)
		signal.Notify(closeChan, syscall.SIGINT, syscall.SIGTERM)
		<-closeChan
		log.Println("Chat server exited")
		return nil
	}

	app.Before = altsrc.InitInputSourceWithContext(app.Flags, altsrc.NewTomlSourceFromFlagFunc("config-file"))

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
