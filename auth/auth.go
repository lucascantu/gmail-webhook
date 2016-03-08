package auth

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/url"
	"os"
	"os/user"
	"path/filepath"

	"github.com/google/go-github/github"

	"golang.org/x/net/context"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/gmail/v1"
	"gopkg.in/codegangsta/cli.v1"
)

// getTokenFromWeb uses Config to request a Token.
// It returns the retrieved Token.
func GetTokenFromWeb(config *oauth2.Config) *oauth2.Token {
	authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	fmt.Printf("Go to the following link in your browser then type the "+
		"authorization code: \n%v\n", authURL)

	var code string
	if _, err := fmt.Scan(&code); err != nil {
		log.Fatalf("Unable to read authorization code %v", err)
	}

	tok, err := config.Exchange(oauth2.NoContext, code)
	if err != nil {
		log.Fatalf("Unable to retrieve token from web %v", err)
	}
	return tok
}

// tokenCacheFile generates credential file path/filename.
// It returns the generated credential path/filename.
func TokenCacheFile(c *cli.Context) (string, error) {
	if c.IsSet("cache-file") {
		return c.String("cache-file"), nil
	}
	usr, err := user.Current()
	if err != nil {
		return "", err
	}
	tokenCacheDir := filepath.Join(usr.HomeDir, ".credentials")
	os.MkdirAll(tokenCacheDir, 0700)
	return filepath.Join(tokenCacheDir,
		url.QueryEscape("gmail.json")), err
}

// tokenFromFile retrieves a Token from a given file path.
// It returns the retrieved Token and any read error encountered.
func TokenFromFile(file string) (*oauth2.Token, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	t := &oauth2.Token{}
	err = json.NewDecoder(f).Decode(t)
	defer f.Close()
	return t, err
}

// saveToken uses a file path to create a file and store the
// token in it.
func SaveToken(file string, token *oauth2.Token) {
	fmt.Printf("Saving credential file to: %s\n", file)
	f, err := os.Create(file)
	if err != nil {
		log.Fatalf("Unable to cache oauth token: %v", err)
	}
	defer f.Close()
	json.NewEncoder(f).Encode(token)
}

func GetGmailClient(c *cli.Context) (*gmail.Service, error) {
	var srv *gmail.Service
	cacheFile, err := TokenCacheFile(c)
	if err != nil {
		return srv, fmt.Errorf("error unable to set the token file path %s\n", err)
	}
	tok, err := TokenFromFile(cacheFile)
	if err != nil {
		return srv, fmt.Errorf("error unable to get token from cache file: possibly run the authorize-gmail command")
	}
	cont, err := ioutil.ReadFile(c.String("gmail-secret"))
	if err != nil {
		return srv, fmt.Errorf("error unable to read the secret json file %s\n", err)
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
		return srv, fmt.Errorf("error unable to create oauth config from secret file %s\n", err)
	}
	client := config.Client(context.Background(), tok)
	srv, err = gmail.New(client)
	if err != nil {
		return srv, fmt.Errorf("error unable to set gmail client %s\n", err)
	}
	return srv, nil
}

func GetGithubClient(c *cli.Context) (*github.Client, error) {
	var client *github.Client
	tok, err := ioutil.ReadFile(c.String("gh-token"))
	if err != nil {
		return client, fmt.Errorf("error cannot open token file %s\n", err)
	}
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: string(tok)},
	)
	tc := oauth2.NewClient(oauth2.NoContext, ts)
	return github.NewClient(tc), nil
}
