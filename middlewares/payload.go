package middlewares

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/cyclopsci/apollo"
	"golang.org/x/net/context"
)

type GmailPayload struct {
	Message struct {
		Data      string `json:"data"`
		MessageID int    `json:"message_id"`
	} `json:"message"`
	Subscription string `json:"subscription"`
}

func DecodeMiddleware(h apollo.Handler) apollo.Handler {
	fn := func(ctx context.Context, w http.ResponseWriter, r *http.Request) {
		var payload GmailPayload
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			log.Printf("error in payload decoding %s\n", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		h.ServeHTTP(context.WithValue(ctx, "payload", payload), w, r)
	}
	return apollo.HandlerFunc(fn)
}
