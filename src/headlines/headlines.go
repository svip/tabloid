package headlines

import (
	"bytes"
	"fmt"
	"html/template"
	"log"
	"math/rand"
	"regexp"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"

	"github.com/PuerkitoBio/goquery"
)

// The template we use for printing the headline.
// Who knows if it's thread safe?!
var headlineT *template.Template
var headlineshT *template.Template
var headlinestT *template.Template

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

type NewHeadline struct {
	No1       Headline
	No2       Headline
	Separator string
}

func (h NewHeadline) Print() template.HTML {
	out := bytes.NewBufferString("")
	headlineshT.Execute(out, h)
	return template.HTML(out.String())
}

func (h NewHeadline) Title() template.HTML {
	out := bytes.NewBufferString("")
	headlinestT.Execute(out, h)
	return template.HTML(out.String())
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
	// Adding headlines is only done once every hour.
	// Include "!" and "?" in headlines, but also split upon them.
	// Also split on ";", ":" and " -", all followed by a space.
	// Or rather, this doesn't split but finds the headlines we're
	// looking for.  Because you cannot split on something you wish to include.
	re := regexp.MustCompile("(.+?([!\\?;:] +| +[-–] +|[;:] +[-–] +|$))")
	reTrim := regexp.MustCompile("[ \n\t ]+")
	reOrim := regexp.MustCompile("([;:\\.]$|^[-–] | [-–]$)")
	reHref := regexp.MustCompile("[^\\/\\.]+\\.[^\\/\\.]+\\/")

	// Make sure all line breaks are turned into spaces.
	text = strings.TrimSpace(reTrim.ReplaceAllString(text, " "))
	s := re.FindAllString(text, -1)
	if len(s) < 2 {
		return // If we don't get at least two headlines from this headline,
		       // we don't care.  Because it usually means the headline isn't
		       // good material for our purpose.
	}
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
		log.Println(reHref.FindString(href), line)
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

	if len(tempHeadlines) <= 0 {
		log.Fatal("No headlines!")
	}

	headlines = tempHeadlines
}

func GetHeadlineWithRNG(r *rand.Rand) NewHeadline {
	no1 := r.Intn(len(headlines))
	no2 := no1
	for headlines[no1].URL == headlines[no2].URL {
		no2 = r.Intn(len(headlines))
	}
	// Always attempt the shortest headline slice first.
	if len(headlines[no1].Headline) > len(headlines[no2].Headline) {
		no1, no2 = no2, no1
	}
	sep := ":"
	if strings.ContainsAny(headlines[no1].Headline[len(headlines[no1].Headline)-1:], "?!") {
		sep = ""
	}
	return NewHeadline{
		headlines[no1],
		headlines[no2],
		sep,
	}
}

func GetHeadline() NewHeadline {
	rng := rand.New(rand.NewSource(time.Now().Unix()))
	return GetHeadlineWithRNG(rng)
}

func init() {
	// Prepare our global variable.
	var err error
	headlineT, err = template.New("headline").Parse(`<a href="{{.URL}}">{{.Headline}}</a>`)
	if err != nil {
		log.Fatal(err)
	}
	headlineshT, err = template.New("headlines-html").Parse(`{{.No1.Print}}{{.Separator}} {{.No2.Print}}`)
	if err != nil {
		log.Fatal(err)
	}
	headlinestT, err = template.New("headlines-title").Parse(`{{.No1.Headline}}{{.Separator}} {{.No2.Headline}}`)
	if err != nil {
		log.Fatal(err)
	}
}
