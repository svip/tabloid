package main

import (
	"net/http"
	"log"
	"html/template"
	"math/rand"
	"time"
	"headlines"
)

const PAGE = `<!doctype html>
<html>
<head>
<title>{{.Headline1.Headline}}: {{.Headline2.Headline}}</title>
<meta charset="utf-8" />
<style>
body {
	font-family: sans-serif;
	display: flex;
	justify-content: center;
	align-items: center;
	height: 100ch;
	overflow: hidden;
	background: red;
}
a {
	color: #000;
	text-decoration: none;
}
@keyframes headline {
	from {
		transform: rotate(-250deg);
	}
}
#headline {
	font-size: 30pt;
	font-weight: bold;
	animation-duration: 0.5s;
	animation-name: headline;
	transform: rotate({{.Degree}}deg);
	background: orange;
	padding: 5px 10px;
	margin-top: -5ch;
}
</style>
</head>
<body>
<div id="headline">{{.Headline1.Print}}: {{.Headline2.Print}}</div>
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

func headlinePage(w http.ResponseWriter, r *http.Request) {
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

func main() {
	var err error
	t, err = template.New("page").Parse(PAGE)
	if err != nil {
		log.Fatal(err)
	}
	
	lastSeed = time.Now()
	rnd = rand.New(rand.NewSource(lastSeed.Unix()))
	
	headlines.UpdateHeadlines()
	
	http.HandleFunc("/", headlinePage)
	
	log.Fatal(http.ListenAndServe(":8070", nil))
}
