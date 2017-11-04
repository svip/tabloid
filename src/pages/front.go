package pages

import (
	"headlines"
	"html/template"
	"log"
	"math/rand"
	"net/http"
	"time"
)

// This is how simple an HTML page can be!  Woo!
const PAGE = `<!doctype html>
<html>
<head>
<title>{{.Headlines.Title}}</title>
<meta charset="utf-8" />
<link rel="stylesheet" href="/media/styles.css" />
</head>
<body>
<div id="headline" style="transform: rotate({{.Degree}}deg);">{{.Headlines.Print}}</div>
</body>
</html>`

// Global fun.
var t *template.Template
var rnd *rand.Rand
var lastSeed time.Time

func checkSeed() {
	// Every hour, we update our random seed and
	// check for new headlines.
	if time.Since(lastSeed).Hours() >= 1 {
		lastSeed = time.Now()
		rnd.Seed(lastSeed.Unix())
		headlines.UpdateHeadlines()
	}
}

// Yeah, the end rotation is calculated server side.
// *sunglasses*
func getDegree() int {
	return rnd.Intn(35*2) - 35
}

func HeadlinePage(w http.ResponseWriter, r *http.Request) {
	checkSeed()
	headline := headlines.GetHeadlineWithRNG(rnd)
	
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	
	// Build the page with the rotation and the two
	// headlines, that will be combined.
	err := t.Execute(w, struct {
		Degree    int
		Headlines headlines.NewHeadline
	}{
		getDegree(),
		headline,
	})
	if err != nil {
		log.Println(err)
	}
}

func init() {
	// Prepare our global variables.
	var err error
	t, err = template.New("page").Parse(PAGE)
	if err != nil {
		log.Fatal(err)
	}

	lastSeed = time.Now()
	rnd = rand.New(rand.NewSource(lastSeed.Unix()))

	headlines.UpdateHeadlines()
}
