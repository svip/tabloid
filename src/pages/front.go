package pages

import (
	"html/template"
	"headlines"
	"math/rand"
	"time"
	"net/http"
	"log"
)

const PAGE = `<!doctype html>
<html>
<head>
<title>{{.Headline1.Headline}}: {{.Headline2.Headline}}</title>
<meta charset="utf-8" />
<link rel="stylesheet" href="/media/styles.css" />
</head>
<body>
<div id="headline" style="transform: rotate({{.Degree}}deg);">{{.Headline1.Print}}: {{.Headline2.Print}}</div>
</body>
</html>`

var t *template.Template
var rnd *rand.Rand
var lastSeed time.Time

func checkSeed() {
	if time.Since(lastSeed).Hours() >= 1 {
		lastSeed = time.Now()
		rnd.Seed(lastSeed.Unix())
		headlines.UpdateHeadlines()
	}
}

func getDegree() int {
	return rnd.Intn(35*2)-35
}

func HeadlinePage(w http.ResponseWriter, r *http.Request) {
	checkSeed()
	headline1, headline2 := headlines.GetHeadline(rnd)
	err := t.Execute(w, struct{
		Degree int
		Headline1 headlines.Headline
		Headline2 headlines.Headline
	}{
		getDegree(),
		headline1,
		headline2,
	})
	if err != nil {
		log.Println(err)
	}
}

func init() {
	var err error
	t, err = template.New("page").Parse(PAGE)
	if err != nil {
		log.Fatal(err)
	}
	
	lastSeed = time.Now()
	rnd = rand.New(rand.NewSource(lastSeed.Unix()))
	
	headlines.UpdateHeadlines()
}
