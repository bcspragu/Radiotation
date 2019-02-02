package srv

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"

	"firebase.google.com/go/auth"

	"github.com/bcspragu/Radiotation/db"
)

func (s *Srv) serveVerifyToken(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Token     string `json:"token"`
		Name      string `json:"name"`
		Anonymous bool   `json:"anonymous"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("failed to decode request: %v", err)
		return
	}

	tkn, err := s.verifyIDToken(r.Context(), req.Token)
	if err != nil {
		log.Printf("verifyIDToken(%s): %v", req.Token, err)
		return
	}

	ns := strings.Split(req.Name, " ")
	first, last := ns[0], ""
	if len(ns) > 1 {
		last = strings.Join(ns[1:], " ")
	}
	if req.Anonymous {
		first, last = "Anonymous", "User"
	}

	u := &db.User{
		ID:    db.UserID(tkn.UID),
		First: first,
		Last:  last,
	}

	// Store the information in our DB.
	s.createUser(w, u)
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{"success": true}`))
}

func (s *Srv) verifyIDToken(ctx context.Context, rawToken string) (*auth.Token, error) {
	tkn, err := s.authClient.VerifyIDToken(ctx, rawToken)
	if err != nil {
		return nil, fmt.Errorf("VerifyIDToken: %v", err)
	}

	return tkn, nil
}
