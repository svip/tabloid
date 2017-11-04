package headlines

import (
	"bytes"
	"fmt"
	"html/template"
	"log"
	"math/rand"
	"regexp"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/PuerkitoBio/goquery"
)

// The template we use for printing the headline.
// Who knows if it's thread safe?!
var headlineT *template.Template

type Headline struct {
	Headline string
	URL      string
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
	// Maybe these should be global as well?
	// But then again, adding headlines is only done once every hour.
	re := regexp.MustCompile("(([^!\\?]+?[!\\?]) |(.+?)(: | [-–] |: [-–] |$))")
	reTrim := regexp.MustCompile("[ \n\t]+")
	reOrim := regexp.MustCompile("(:$|^[-–] | [-–]$)")

	text = strings.TrimSpace(reTrim.ReplaceAllString(text, " "))
	s := re.FindAllString(text, -1)
	for _, line := range s {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		// Correctly uppercase the first rune.
		r, z := utf8.DecodeRuneInString(line)
		if unicode.IsLetter(r) && unicode.IsLower(r) {
			line = string(unicode.ToTitle(r)) + line[z:]
		}
		line = reTrim.ReplaceAllString(line, " ")
		line = reOrim.ReplaceAllString(line, "")
		log.Println(line)
		//log.Println(href)
		tempHeadlines = append(tempHeadlines, Headline{line, href})
	}
}

func fillFromDomain(newssite newsSite) {
	var domain string
	var selector string
	var urlprefix string
	var urlSelector string
	
	// Known sites and their special handling
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
	
	// goquery is basically jQuery for Go.
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
		
		// Sometimes the links are not relative.
		if !strings.HasPrefix(url, "http") {
			url = fmt.Sprintf("%s%s", urlprefix, url)
		}
		addHeadlines(s.Text(), url)
	})
}

func UpdateHeadlines() {
	// Work on the temporary list before copying it to the main one.
	// For some suggestion of thread-safety.  Wat?
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
	// Prepare our global variable.
	var err error
	headlineT, err = template.New("headline").Parse(`<a href="{{.URL}}">{{.Headline}}</a>`)
	if err != nil {
		log.Fatal(err)
	}
}
