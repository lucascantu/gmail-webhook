package handlers

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/dictybase/gmail-webhook/middlewares"
	"github.com/google/go-github/github"
	"golang.org/x/net/context"
	"google.golang.org/api/gmail/v1"
)

type DscClient struct {
	Gmail      *gmail.Service
	Github     *github.Client
	Label      string
	Repository string
	Owner      string
}

type user struct {
	EmailAddress string `json:"emailAddress"`
	HistoryID    uint64 `json:"historyId"`
}

func (dicty *DscClient) StockOrderHandler(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	payload, _ := ctx.Value("payload").(*middlewares.GmailPayload)
	data, err := base64.URLEncoding.DecodeString(payload.Message.Data)
	if err != nil {
		log.Printf("error in decoding base64 data %s\n", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	var u user
	if err := json.NewDecoder(bytes.NewReader(data)).Decode(&u); err != nil {
		log.Printf("error in decoding json data %s\n", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	histListCall := gmail.NewUsersHistoryService(dicty.Gmail).List("me").StartHistoryId(u.HistoryID)
	var histList []*gmail.History
	for {
		respList, err := histListCall.Do()
		if err != nil {
			log.Printf("error in making history call %s\n", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		histList = append(histList, respList.History...)
		if len(respList.NextPageToken) > 0 {
			histListCall = histListCall.PageToken(respList.NextPageToken)
			continue
		}
		break
	}
	var issues []string
	for _, h := range histList {
		for _, l := range h.LabelsAdded {
			if dicty.MatchLabel(l.LabelIds) {
				subject := parseSubject(l.Message.Payload)
				body := l.Message.Payload.Body.Data
				issue, _, err := dicty.Github.Issues.Create(
					dicty.Owner,
					dicty.Repository,
					&github.IssueRequest{
						Title: &subject,
						Body:  &body,
					},
				)
				if err != nil {
					log.Printf("error in creating github issue %s\n", err)
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}
				issues = append(issues, strconv.Itoa(*issue.Number))
			}
		}
	}
	if len(issues) > 0 {
		fmt.Fprintf(w, "created issues %s\n", strings.Join(issues, " "))
		return
	}
	w.Write([]byte("No issue created"))
}

func (dicty *DscClient) MatchLabel(labels []string) bool {
	for _, name := range labels {
		if name == dicty.Label {
			return true
		}
	}
	return false
}

func parseSubject(m *gmail.MessagePart) string {
	for _, h := range m.Headers {
		if h.Name == "Subject" {
			return h.Value
		}
	}
	return ""
}
