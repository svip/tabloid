package headlines

import (
	"html/template"
	"bytes"
	"regexp"
	"strings"
	"unicode"
	"unicode/utf8"
	"log"
	"fmt"
	"math/rand"
	
	"github.com/PuerkitoBio/goquery"
)

type Headline struct {
	Headline string
	URL string
}

func (h Headline) Print() template.HTML {
	if h.URL == "" {
		return template.HTML(h.Headline)
	} else {
		out := bytes.NewBufferString("")
		headlineT.Execute(out, h)
		return template.HTML(out.String())
	}
}

// Global variables that don't need recreating.
var headlineT *template.Template

// Storage of headlines
var headlines []Headline
var tempHeadlines []Headline

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
		tempHeadlines = append(tempHeadlines, Headline{sss, href})
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

func UpdateHeadlines() {
	tempHeadlines = make([]Headline, 0)
	
	fillFromDomain(eb)
	fillFromDomain(bt)
	fillFromDomain(dr)
	
	headlines = tempHeadlines
	
	if len(headlines) <= 0 {
		log.Fatal("No headlines!")
	}
}

func GetHeadline(r *rand.Rand) (Headline, Headline) {
	no1 := r.Intn(len(headlines))
	no2 := no1
	for no1 == no2 {
		no2 = r.Intn(len(headlines))
	}
	//log.Println(headlines[no1].Headline)
	//log.Println(headlines[no2].Headline)
	return headlines[no1], headlines[no2]
}

func init() {
	var err error 
	headlineT, err = template.New("headline").Parse(`<a href="{{.URL}}">{{.Headline}}</a>`)
	if err != nil {
		log.Fatal(err)
	}
}
