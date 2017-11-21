package srv

import (
	"context"
	"log"
	"net/http"

	"github.com/bcspragu/Radiotation/db"
	oidc "github.com/coreos/go-oidc"
)

func (s *Srv) serveVerifyToken(w http.ResponseWriter, r *http.Request) {
	token := r.PostFormValue("token")
	ti, err := s.verifyIdToken(token)
	if err != nil {
		log.Printf("verifyIdToken(%s): %v", token, err)
		return
	}

	var name struct {
		First string `json:"given_name"`
		Last  string `json:"family_name"`
	}
	if err := ti.Claims(&name); err != nil {
		log.Printf("token.Claims: %v", err)
		return
	}

	// If the token is good, store the information in the user's encrypted cookie
	u := db.GoogleUser(ti.Subject, name.First, name.Last)
	s.createUser(w, u)
	w.Write([]byte("success"))
}

func (s *Srv) verifyIdToken(rawIDToken string) (*oidc.IDToken, error) {
	// Verify the token
	idToken, err := s.googleVerifier.Verify(context.Background(), rawIDToken)
	if err != nil {
		return nil, err
	}

	return idToken, err
}
