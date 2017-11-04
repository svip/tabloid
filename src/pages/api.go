package pages

import (
	"net/http"
	"headlines"
	"encoding/json"
)

func Api(w http.ResponseWriter, r *http.Request) {
	checkSeed()
	headline := headlines.GetHeadlineWithRNG(rnd)
	
	enc := json.NewEncoder(w)
	
	enc.Encode(struct {
		Headline string
		URLs     []string
	}{
		string(headline.Title()),
		[]string{
			headline.No1.URL,
			headline.No2.URL,
		},
	})
}
