package main

import (
	"os"

	"github.com/dictybase/gmail-webhook/commands"
	"gopkg.in/codegangsta/cli.v1"
)

func main() {
	app := cli.NewApp()
	app.Version = "1.0.0"
	app.Name = "gmail-webhook"
	app.Usage = "Manage gmail push notifications"
	app.Commands = []cli.Command{
		{
			Name:   "subscribe",
			Usage:  "create a new subscription to a topic",
			Action: commands.SubscribeAction,
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "key-file, k",
					Usage: "key file for authorization(required)",
				},
				cli.StringFlag{
					Name:  "project-id, id",
					Usage: "unique id of the project(required)",
				},
				cli.StringFlag{
					Name:  "topic, t",
					Usage: "Name of a topic that is already created",
				},
				cli.StringFlag{
					Name:  "name, n",
					Usage: "Name of the subscription",
				},
				cli.StringFlag{
					Name:  "endpoint, e",
					Usage: "Name of the subscription endpoint",
				},
			},
		},
		{
			Name:   "authorize",
			Usage:  "authorize gmail client",
			Action: commands.AuthGmailAction,
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "gmail-secret, gs",
					Usage: "gmail client secret json file",
				},
				cli.StringFlag{
					Name:   "cache-file, cf",
					Usage:  "location of cached gmail token file, defaults to ~/.credentials/gmail.json",
					EnvVar: "CACHE_TOKEN_FILE",
				},
			},
		},
		{
			Name:   "watch",
			Usage:  "setup watch request for subscribed topic",
			Action: commands.WatchGmailAction,
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "topic, t",
					Usage: "Name of the topic",
				},
				cli.StringFlag{
					Name:  "project, p",
					Usage: "Name of the project",
				},
				cli.StringFlag{
					Name:   "cache-file, cf",
					Usage:  "location of cached gmail token file, defaults to ~/.credentials/gmail.json",
					EnvVar: "CACHE_TOKEN_FILE",
				},
				cli.StringFlag{
					Name:  "gmail-secret, gs",
					Usage: "gmail client secret json file",
				},
				cli.StringFlag{
					Name:  "redis-address",
					Usage: "IP address of redis-server",
					Value: "redis",
				},
				cli.IntFlag{
					Name:  "redis-port",
					Usage: "Port of redis server",
					Value: 6379,
				},
			},
		},
		{
			Name:   "run",
			Usage:  "starts the webhook server for gmail push notifications",
			Action: commands.RunServer,
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "subscription, s",
					Usage: "Name of the subscription",
				},
				cli.StringFlag{
					Name:  "project, p",
					Usage: "Name of the project",
				},
				cli.StringFlag{
					Name:   "cache-file, cf",
					Usage:  "location of cached gmail token file, defaults to ~/.credentials/gmail.json",
					EnvVar: "CACHE_TOKEN_FILE",
				},
				cli.StringFlag{
					Name:  "gmail-secret, gs",
					Usage: "gmail client secret json file",
				},
				cli.StringFlag{
					Name:  "gh-token, ght",
					Usage: "github personal access token file, defaults to ~/.credentials/github.json",
				},
				cli.StringFlag{
					Name:  "log,l",
					Usage: "Name of the log file(optional), default goes to stderr",
				},
				cli.IntFlag{
					Name:  "port",
					Usage: "port on which the server listen",
					Value: 9998,
				},
				cli.StringFlag{
					Name:  "label",
					Usage: "Gmail label which will be filtered for messages",
				},
				cli.StringFlag{
					Name:  "repository, r",
					Usage: "Github repository",
				},
				cli.StringFlag{
					Name:  "owner",
					Usage: "Github repository owner",
				},
				cli.StringFlag{
					Name:  "redis-address",
					Usage: "IP address of redis-server",
					Value: "redis",
				},
				cli.IntFlag{
					Name:  "redis-port",
					Usage: "Port of redis server",
					Value: 6379,
				},
			},
		},
	}
	app.Run(os.Args)
}
