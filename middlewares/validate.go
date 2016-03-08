package middlewares

import (
	"fmt"
	"log"
	"net/http"

	"github.com/cyclopsci/apollo"
	"golang.org/x/net/context"
)

type GmailSubscription struct {
	Name string
}

func (s *GmailSubscription) ValidateMiddleware(h apollo.Handler) apollo.Handler {
	fn := func(ctx context.Context, w http.ResponseWriter, r *http.Request) {
		payload, ok := ctx.Value("payload").(*GmailPayload)
		if !ok {
			log.Println("error in retrieving payload from *payload*")
			http.Error(w, "unable to retrieve payload from *payload*", http.StatusInternalServerError)
			return
		}
		if s.Name != payload.Subscription {
			log.Printf("Expected subscription %s does not match with existing subscription %s\n", s.Name, payload.Subscription)
			http.Error(
				w,
				fmt.Sprintf("Expected subscription %s does not match with existing subscription %s\n", s.Name, payload.Subscription),
				http.StatusInternalServerError,
			)
			return
		}
		h.ServeHTTP(ctx, w, r)
	}
	return apollo.HandlerFunc(fn)
}
