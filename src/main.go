package main

import (
	"net/http"
	"log"
	"html/template"
	"math/rand"
	"time"
	"regexp"
	"fmt"
	"strings"
	"bytes"
	"unicode"
	"unicode/utf8"
	
	"github.com/PuerkitoBio/goquery"
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
var headlineT *template.Template
var r *rand.Rand
var lastSeed time.Time

func checkSeed() {
	if time.Since(lastSeed).Hours() >= 1 {
		lastSeed = time.Now()
		r.Seed(lastSeed.Unix())
		fillHeadlines()
	}
}

func getDegree() int {
	return r.Intn(35*2)-35
}

type headline struct {
	Headline string
	URL string
}

func (h headline) Print() template.HTML {
	if h.URL == "" {
		return template.HTML(h.Headline)
	} else {
		out := bytes.NewBufferString("")
		headlineT.Execute(out, h)
		return template.HTML(out.String())
	}
}

var headlines []headline
var tempHeadlines []headline

type newsSite int

const (
	eb newsSite = 1 + iota
	bt
	dr
)

func addHeadlines(text, href string) {
	re := regexp.MustCompile("(: | - )")
	reTrim := regexp.MustCompile("[ \n\t]+")
	
	text = strings.TrimSpace(reTrim.ReplaceAllString(text, " "))
	s := re.Split(text, -1)
	for _, sss := range s {
		sss = strings.TrimSpace(sss)
		if sss == "" {
			continue
		}
		r, z := utf8.DecodeRuneInString(sss)
		if unicode.IsLetter(r) && unicode.IsLower(r) {
			sss = string(unicode.ToTitle(r)) + sss[z:]
		}
		sss = reTrim.ReplaceAllString(sss, " ")
		log.Println(sss)
		//log.Println(href)
		tempHeadlines = append(tempHeadlines, headline{sss, href})
	}
}

func fillFromDomain(newssite newsSite) {
	var domain string
	var selector string
	var urlprefix string
	var urlSelector string
	switch newssite {
	case eb:
		domain = "http://www.ekstrabladet.dk/"
		selector = ".df-article"
		urlprefix = "http://www.ekstrabladet.dk/"
		urlSelector = "a"
	case bt:
		domain = "http://bt.dk/"
		selector = "a.teaser-link"
		urlprefix = "http://bt.dk"
	case dr:
		domain = "https://www.dr.dk/nyheder/allenyheder/"
		selector = "article h3 a"
		urlprefix = "https://www.dr.dk"
	default:
		return
	}
	
	doc, err := goquery.NewDocument(domain)
	if err != nil {
		log.Println(err)
		return
	}
	
	doc.Find(selector).Each(func(i int, s *goquery.Selection) {
		var url string
		if urlSelector != "" {
			url = s.Find(urlSelector).AttrOr("href", "")
		} else {
			url = s.AttrOr("href", "")
		}
		if !strings.HasPrefix(url, "http") {
			url = fmt.Sprintf("%s%s", urlprefix, url)
		}
		addHeadlines(s.Text(), url)
	})
}	

func fillHeadlines() {
	tempHeadlines = make([]headline, 0)
	
	fillFromDomain(eb)
	fillFromDomain(bt)
	fillFromDomain(dr)
	
	headlines = tempHeadlines
	
	if len(headlines) <= 0 {
		log.Fatal("No headlines!")
	}
}

func getHeadline() (headline, headline) {
	no1 := r.Intn(len(headlines))
	no2 := no1
	for no1 == no2 {
		no2 = r.Intn(len(headlines))
	}
	//log.Println(headlines[no1].Headline)
	//log.Println(headlines[no2].Headline)
	return headlines[no1], headlines[no2]
}

func headlinePage(w http.ResponseWriter, r *http.Request) {
	checkSeed()
	headline1, headline2 := getHeadline()
	err := t.Execute(w, struct{
		Degree int
		Headline1 headline
		Headline2 headline
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
	
	headlineT, err = template.New("headline").Parse(`<a href="{{.URL}}">{{.Headline}}</a>`)
	if err != nil {
		log.Fatal(err)
	}
	
	lastSeed = time.Now()
	r = rand.New(rand.NewSource(lastSeed.Unix()))
	
	fillHeadlines()
	
	http.HandleFunc("/", headlinePage)
	
	log.Fatal(http.ListenAndServe(":8070", nil))
}
