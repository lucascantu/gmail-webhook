package commands

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"golang.org/x/net/context"

	"github.com/cyclopsci/apollo"
	"github.com/dictybase/gmail-webhook/auth"
	"github.com/dictybase/gmail-webhook/handlers"
	"github.com/dictybase/gmail-webhook/middlewares"
	"gopkg.in/codegangsta/cli.v1"
)

func ValidateServerOptions(c *cli.Context) error {
	for _, v := range []string{"subscription", "project", "gmail-secret", "gh-token"} {
		if !c.IsSet(v) {
			return fmt.Errorf("missing command line argument %s\n", v)
		}
	}
	return nil
}

func RunServer(c *cli.Context) {
	if err := ValidateServerOptions(c); err != nil {
		log.Fatal(err)
	}
	var logMw *middlewares.Logger
	if c.IsSet("log") {
		w, err := os.Create(c.String("log"))
		if err != nil {
			log.Fatalf("cannot open log file %q\n", err)
		}
		defer w.Close()
		logMw = middlewares.NewFileLogger(w)
	} else {
		logMw = middlewares.NewLogger()
	}
	gmClient, err := auth.GetGmailClient(c)
	if err != nil {
		log.Fatal(err)
	}
	ghClient, err := auth.GetGithubClient(c)
	if err != nil {
		log.Fatal(err)
	}

	mux := http.NewServeMux()
	valMw := &middlewares.GmailSubscription{
		fmt.Sprintf(
			"projects/%s/subscriptions/%s",
			c.String("project"),
			c.String("subscription"),
		),
	}
	dsc := &handlers.DscClient{
		Gmail:      gmClient,
		Github:     ghClient,
		Label:      c.String("label"),
		Repository: c.String("repository"),
		Owner:      c.String("owner"),
	}
	dscChain := apollo.New(
		apollo.Wrap(logMw.LoggerMiddleware),
		middlewares.DecodeMiddleware,
		valMw.ValidateMiddleware,
	).With(context.Background()).ThenFunc(dsc.StockOrderHandler)

	mux.Handle("/gmail/order", dscChain)
	log.Printf("Starting web server on port %d\n", c.Int("port"))
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", c.Int("port")), mux))
}
