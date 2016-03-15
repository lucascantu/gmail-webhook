package handlers

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
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
	Logger     *log.Logger
}

type user struct {
	EmailAddress string `json:"emailAddress"`
	HistoryID    uint64 `json:"historyId"`
}

func (dicty *DscClient) StockOrderHandler(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	payload, _ := ctx.Value("payload").(*middlewares.GmailPayload)
	data, err := base64.URLEncoding.DecodeString(payload.Message.Data)
	if err != nil {
		dicty.Logger.Printf("error in decoding base64 data %s\n", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	var u user
	if err := json.NewDecoder(bytes.NewReader(data)).Decode(&u); err != nil {
		dicty.Logger.Printf("error in decoding json data %s\n", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	histId, err := dicty.HistoryDbh.GetCurrentHistory()
	if err != nil {
		dicty.Logger.Printf("error in getting current history %s\n", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	dicty.Logger.Printf("current history id %d\n", histId)
	dicty.Logger.Printf("mailbox history id %d\n", u.HistoryID)
	err = dicty.HistoryDbh.SetCurrentHistory(u.HistoryID)
	if err != nil {
		dicty.Logger.Printf("error in setting history %d %s\n", u.HistoryID, err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	histList, err := dicty.GetHistories(histId)
	if err != nil {
		dicty.Logger.Print(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if len(histList) == 0 {
		dicty.Logger.Println("got no history")
		w.Write([]byte("got no history"))
		return
	}
	log.Printf("got %d histories\n", len(histList))

	messages, err := dicty.GetMatchingMessages(histList)
	if err != nil {
		dicty.Logger.Print(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if len(messages) == 0 {
		dicty.Logger.Println("got no messages matching label")
		w.Write([]byte("got no messages matching label"))
		return
	}
	log.Printf("%d messages matches histories\n", len(messages))

	issues, err := dicty.GetGithubIssues(messages)
	if err != nil {
		dicty.Logger.Print(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	for _, gs := range issues {
		_, _, err := dicty.Github.Issues.Create(
			dicty.Owner,
			dicty.Repository,
			gs,
		)
		if err != nil {
			dicty.Logger.Println("error in creating github issue %s\n", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
	srvMsg := fmt.Sprintf("created %d issues", len(issues))
	log.Println(srvMsg)
	w.Write([]byte(srvMsg))
}

func (dicty *DscClient) MatchLabel(labels []string) bool {
	for _, name := range labels {
		if name == dicty.Label {
			return true
		}
	}
	return false
}

func (dicty *DscClient) GetHistories(id uint64) ([]*gmail.History, error) {
	pageToken := ""
	var histList []*gmail.History
	for {
		histListCall := gmail.NewUsersHistoryService(
			dicty.Gmail).
			List("me").
			StartHistoryId(id)
		if pageToken != "" {
			histListCall = histListCall.PageToken(pageToken)
		}
		respList, err := histListCall.Do()
		if err != nil {
			return histList, fmt.Errorf("error in making history call %s\r", err)
		}
		histList = append(histList, respList.History...)
		if respList.NextPageToken == "" {
			return histList, nil
		}
		pageToken = respList.NextPageToken
	}
	return histList, nil
}

func (dicty *DscClient) GetMatchingMessages(histList []*gmail.History) ([]*gmail.Message, error) {
	var messages []*gmail.Message
	for _, h := range histList {
		for _, m := range h.Messages {
			msg, err := dicty.Gmail.Users.Messages.Get("me", m.Id).Do()
			if err != nil {
				return messages, fmt.Errorf("error in retrieving message %s %s", m.Id, err)
			}
			if dicty.MatchLabel(msg.LabelIds) {
				messages = append(messages, msg)
			}
		}
	}
	return messages, nil
}

func (dicty *DscClient) GetGithubIssues(msgs []*gmail.Message) ([]*github.IssueRequest, error) {
	var issues []*github.IssueRequest
	for _, msg := range msgs {
		title := parseSubject(msg.Payload)
		body, err := parseBody(msg.Payload)
		if err != nil {
			return issues, fmt.Errorf("error in parsing body %s\n", err)
		}
		issues = append(issues, &github.IssueRequest{Title: &title, Body: &body})
	}
	return issues, nil
}

func parseSubject(m *gmail.MessagePart) string {
	for _, h := range m.Headers {
		if h.Name == "Subject" {
			return h.Value
		}
	}
	return ""
}

func parseBody(m *gmail.MessagePart) (string, error) {
	if strings.HasPrefix(m.MimeType, "multipart") {
		for _, p := range m.Parts {
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
