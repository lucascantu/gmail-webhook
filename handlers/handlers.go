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

	"github.com/dictybase/gmail-webhook/history"
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
	HistoryDbh *history.HistoryDb
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
	histId, err := dicty.HistoryDbh.GetCurrentHistory()
	if err != nil {
		log.Printf("error in getting current history %s\n", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	pageToken := ""
	var histList []*gmail.History
	log.Printf("current history id %d\n", histId)
	log.Printf("got history id %d\n", u.HistoryID)
	for {
		histListCall := gmail.NewUsersHistoryService(dicty.Gmail).List("me").StartHistoryId(histId)
		if pageToken != "" {
			histListCall = histListCall.PageToken(pageToken)
		}
		respList, err := histListCall.Do()
		if err != nil {
			log.Printf("error in making history call %s\n", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		histList = append(histList, respList.History...)
		log.Printf("got %d histories\n", len(respList.History))
		if respList.NextPageToken == "" {
			break
		}
		pageToken = respList.NextPageToken
	}

	log.Printf("got %d list of histories\n", len(histList))

	var messages []*gmail.Message
	var srvResp string
	for _, h := range histList {
		log.Printf("added %d messages\n", len(h.MessagesAdded))
		for _, m := range h.Messages {
			msg, err := dicty.Gmail.Users.Messages.Get("me", m.Id).Do()
			if err != nil {
				log.Printf("error in retrieving message %s %s", m.Id, err)
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			if dicty.MatchLabel(msg.LabelIds) {
				messages = append(messages, msg)
			}
		}
	}
	if len(messages) > 0 {
		log.Printf("got %d new messages\n", len(messages))
		var issues []string
		for _, msg := range messages {
			title := parseSubject(msg.Payload)
			body, err := parseBody(msg.Payload)
			if err != nil {
				log.Printf("error in parsing body %s\n", err)
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			issue, _, err := dicty.Github.Issues.Create(
				dicty.Owner,
				dicty.Repository,
				&github.IssueRequest{
					Title: &title,
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
		if len(issues) > 0 {
			log.Printf("created issues %s\n", strings.Join(issues, " "))
			fmt.Fprintf(w, "created issues %s\n", strings.Join(issues, " "))
			return
		}
		log.Println("No issues created")
		srvResp = "No issue created"
	} else {
		srvResp = "No new messages"
	}

	err = dicty.HistoryDbh.SetCurrentHistory(u.HistoryID)
	if err != nil {
		log.Printf("error in setting history %d %s\n", u.HistoryID, err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Write([]byte(srvResp))
}

func (dicty *DscClient) MatchLabel(labels []string) bool {
	for _, name := range labels {
		log.Printf("got label %s\n", name)
		if name == dicty.Label {
			return true
		}
	}
	return false
}

func parseSubject(m *gmail.MessagePart) string {
	log.Printf("got %d headers", len(m.Headers))
	for _, h := range m.Headers {
		if h.Name == "Subject" {
			return h.Value
		}
	}
	return ""
}

func parseBody(m *gmail.MessagePart) (string, error) {
	log.Printf("mimetype is %s\n", m.MimeType)
	if strings.HasPrefix(m.MimeType, "multipart") {
		for _, p := range m.Parts {
			log.Printf("part mimetype is %s\n", p.MimeType)
			if p.MimeType == "text/plain" {
				data, err := base64.URLEncoding.DecodeString(p.Body.Data)
				if err != nil {
					return "", err
				}
				return string(data), nil
			}
		}
	}
	data, err := base64.URLEncoding.DecodeString(m.Body.Data)
	if err != nil {
		return "", err
	}
	return string(data), nil
}
