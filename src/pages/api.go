package pages

import (
	"encoding/json"
	"headlines"
	"net/http"
	"html"
	"strings"
)

func apiVote(w http.ResponseWriter, r *http.Request) {
	enc := json.NewEncoder(w)
	if r.Method != "POST" {
		w.WriteHeader(405)
		enc.Encode(struct {
			Error string
		}{
			"Only POST method allowed.",
		})
		return
	}
	token := r.URL.Query().Get("token")
	if token == "" {
		w.WriteHeader(400)
		enc.Encode(struct{
			Error string
		}{
			"You've provided no 'token'.",
		})
		return
	}
	var ok, votingup bool
	if r.URL.Path == "/api/voteup" {
		ok = headlines.VoteHeadlineUp(token)
		votingup = true
	} else if r.URL.Path == "/api/votedown" {
		ok = headlines.VoteHeadlineDown(token)
	} else {
		w.WriteHeader(404)
		enc.Encode(struct{
			Error string
		}{
			"Request not understood",
		})
		return
	}
	if !ok {
		w.WriteHeader(404)
		enc.Encode(struct{
			Error string
		}{
			"No such token.",
		})
		return
	}
	m := "Headline voted down!"
	if votingup {
		m = "Headline voted up!"
	}
	enc.Encode(struct{
		Message string
		Token   string
	}{
		m,
		token,
	})	
}

func Api(w http.ResponseWriter, r *http.Request) {
	if strings.HasPrefix(r.URL.Path, "/api/vote") {
		apiVote(w, r)
		return
	}
	askToCheckSeed()
	headline := headlines.GetHeadlineWithRNG(rnd)

	enc := json.NewEncoder(w)

	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	enc.Encode(struct {
		Headline string
		URLs     []string
		Token    string
	}{
		html.UnescapeString(string(headline.Title())),
		[]string{
			headline.No1.URL,
			headline.No2.URL,
		},
		headline.Token,
	})
}
