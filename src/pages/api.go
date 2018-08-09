package pages

import (
	"encoding/json"
	"headlines"
	"net/http"
	"html"
)

func Api(w http.ResponseWriter, r *http.Request) {
	askToCheckSeed()
	headline := headlines.GetHeadlineWithRNG(rnd)

	enc := json.NewEncoder(w)

	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	enc.Encode(struct {
		Headline string
		URLs     []string
	}{
		html.UnescapeString(string(headline.Title())),
		[]string{
			headline.No1.URL,
			headline.No2.URL,
		},
	})
}
