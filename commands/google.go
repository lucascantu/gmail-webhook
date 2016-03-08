package commands

import (
	"fmt"
	"io/ioutil"
	"log"

	"github.com/dictybase/gmail-webhook/auth"
	"golang.org/x/net/context"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/gmail/v1"
	"google.golang.org/cloud"
	"google.golang.org/cloud/pubsub"
	"gopkg.in/codegangsta/cli.v1"
)

func ValidateWatchOptions(c *cli.Context) error {
	for _, v := range []string{"topic", "project", "gmail-secret"} {
		if !c.IsSet(v) {
			return fmt.Errorf("missing command line argument %s\n", v)
		}
	}
	return nil
}

func ValidateGmailOptions(c *cli.Context) error {
	if !c.IsSet("gmail-secret") {
		return fmt.Errorf("missing command line argument %s\n", "gmail-secret")
	}
	return nil
}

func ValidateSubOptions(c *cli.Context) error {
	for _, v := range []string{"topic", "name", "endpoint", "key-file", "project-id"} {
		if !c.IsSet(v) {
			return fmt.Errorf("missing command line argument %s\n", v)
		}
	}
	return nil
}

func DoAuthorization(c *cli.Context) (oauth2.TokenSource, error) {
	var ts oauth2.TokenSource
	jsonKey, err := ioutil.ReadFile(c.GlobalString("key-file"))
	if err != nil {
		return ts, err
	}
	conf, err := google.JWTConfigFromJSON(jsonKey, pubsub.ScopePubSub)
	if err != nil {
		return ts, err
	}
	return conf.TokenSource(context.Background()), nil
}

func SubscribeAction(c *cli.Context) {
	if err := ValidateSubOptions(c); err != nil {
		log.Fatal(err)
	}
	ts, err := DoAuthorization(c)
	if err != nil {
		log.Fatal(err)
	}
	client, err := pubsub.NewClient(context.Background(), c.GlobalString("project-id"), cloud.WithTokenSource(ts))
	if err != nil {
		log.Fatalf("error in making new cloud client %s\n", err)
	}
	th := client.Topic(c.String("topic"))
	sh, err := th.Subscribe(context.Background(), c.String("name"), 0, &pubsub.PushConfig{Endpoint: c.String("endpoint")})
	if err != nil {
		log.Fatalf("error in creating subscription %s\n", err)
	}
	log.Printf("created subscription %s\n", sh.Name())
}

func AuthGmailAction(c *cli.Context) {
	if err := ValidateGmailOptions(c); err != nil {
		log.Fatal(err)
	}
	tokenFile, err := auth.TokenCacheFile(c)
	if err != nil {
		log.Fatalf("error unable to set the token file path %s\n", err)
	}
	cont, err := ioutil.ReadFile(c.String("gmail-secret"))
	if err != nil {
		log.Fatalf("error unable to read the secret json file %s\n", err)
	}
	config, err := google.ConfigFromJSON(
		cont,
		gmail.GmailSendScope,
		gmail.GmailComposeScope,
		gmail.GmailLabelsScope,
		gmail.GmailModifyScope,
		gmail.MailGoogleComScope,
	)
	if err != nil {
		log.Fatalf("error unable to create oauth config from secret file %s\n", err)
	}
	tok := auth.GetTokenFromWeb(config)
	auth.SaveToken(tokenFile, tok)
	log.Printf("saved gmail token to %s file\n", tokenFile)
}

func WatchGmailAction(c *cli.Context) {
	if err := ValidateWatchOptions(c); err != nil {
		log.Fatal(err)
	}
	gm, err := auth.GetGmailClient(c)
	if err != nil {
		log.Fatal(err)
	}
	resp, err := gmail.NewUsersService(gm).Watch(
		"me",
		&gmail.WatchRequest{
			TopicName: fmt.Sprintf("projects/%s/topics/%s", c.String("project"), c.String("topic")),
		},
	).Do()
	if err != nil {
		log.Fatalf("error in executing watch call %s\n", err)
	}
	log.Printf("sucessful watch call with expiration %d and history %d\n", resp.Expiration, resp.HistoryId)
}
