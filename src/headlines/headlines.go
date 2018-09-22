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
	"encoding/json"
	"os"

	"github.com/PuerkitoBio/goquery"
)

// The template we use for printing the headline.
// Who knows if it's thread safe?!
var headlineT *template.Template
var headlineshT *template.Template
var headlinestT *template.Template

const HEADLINESFILE = "headlines.json"

type Headline struct {
	Headline string
	URL      string
	Points   int
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

func (h *Headline) VoteUp() {
	h.Points += 1
}

func (h *Headline) VoteDown() {
	h.Points -= 1
}

type NewHeadline struct {
	No1       *Headline
	No2       *Headline
	Separator string
	Token     string
	Created   time.Time
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
var headlines []*Headline
var tempHeadlines []*Headline

// Storage of created headlines
var newHeadlines []NewHeadline

type newsSite int

const (
	eb newsSite = 1 + iota
	bt
	dr
)

func addHeadlines(text, href string) {
	for _, h := range headlines {
		if h.URL == href {
			// Already added.
			return
		}
	}
	// Adding headlines is only done once every hour.
	// Include "!" and "?" in headlines, but also split upon them.
	// Also split on ";", ":" and " -", all followed by a space.
	// Or rather, this doesn't split but finds the headlines we're
	// looking for.  Because you cannot split on something you wish to include.
	re := regexp.MustCompile("(.+?([!\\?;:] +| +[-–] +|[;:] +[-–] +|$))")
	reTrim := regexp.MustCompile("[ \n\t ]+") // whitespace
	reOrim := regexp.MustCompile("([;:\\.]$|^[-–\\+] | [-–]$)") // ends
	reHref := regexp.MustCompile("[^\\/\\.]+\\.[^\\/\\.]+\\/")
	reHyph := regexp.MustCompile("(\\p{L})- (\\p{L})")

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
		line = reHyph.ReplaceAllString(line, "$1$2")
		log.Println(reHref.FindString(href), line)
		tempHeadlines = append(tempHeadlines, &Headline{line, href, 0})
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
	tempHeadlines = make([]*Headline, 0)

	fillFromDomain(eb)
	fillFromDomain(bt)
	fillFromDomain(dr)

	if len(tempHeadlines) == 0 && len(headlines) == 0 {
		log.Fatal("No headlines!")
	} else if len(tempHeadlines) == 0 {
		log.Println("No headlines in update.")
	}

	headlines = append(headlines, tempHeadlines...)
	saveHeadlinesToFile()
}

func generateToken(r *rand.Rand) string {
	b := make([]byte, 16)
	r.Read(b)
	return fmt.Sprintf("%x", b)
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
	nh := NewHeadline{
		headlines[no1],
		headlines[no2],
		sep,
		generateToken(r),
		time.Now(),
	}
	newHeadlines = append(newHeadlines, nh)
	return nh
}

func GetHeadline() NewHeadline {
	rng := rand.New(rand.NewSource(time.Now().Unix()))
	return GetHeadlineWithRNG(rng)
}

func findNewHeadline(token string) (NewHeadline, bool) {
	for i, nh := range newHeadlines {
		if time.Since(nh.Created) > time.Hour * 2 {
			newHeadlines = append(newHeadlines[:i], newHeadlines[i+1:]...)
			continue
		}
		if nh.Token == token {
			return nh, true
		}
	}
	return NewHeadline{}, false
}

func VoteHeadlineUp(token string) bool {
	nh, ok := findNewHeadline(token)
	if !ok {
		return false
	}
	nh.No1.VoteUp()
	nh.No2.VoteUp()
	saveHeadlinesToFile()
	return true
}

func VoteHeadlineDown(token string) bool {
	nh, ok := findNewHeadline(token)
	if !ok {
		return false
	}
	nh.No1.VoteDown()
	nh.No2.VoteDown()
	saveHeadlinesToFile()
	return true
}

func saveHeadlinesToFile() {
	f, err := os.Create(HEADLINESFILE)
	if err != nil {
		log.Println(err)
		return
	}
	defer f.Close()
	
	enc := json.NewEncoder(f)
	if err := enc.Encode(&headlines); err != nil {
		log.Println(err)
	}
}

func loadHeadlinesFromFile() {
	f, err := os.Open(HEADLINESFILE)
	if err != nil {
		log.Println(err)
		return
	}
	defer f.Close()
	
	dec := json.NewDecoder(f)
	if err := dec.Decode(&headlines); err != nil {
		log.Println(err)
	}
}

func init() {
	loadHeadlinesFromFile()
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
