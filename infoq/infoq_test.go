package infoq_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/tamnd/infoq-cli/infoq"
)

// rssXML wraps items in a minimal RSS 2.0 envelope.
func rssXML(items string) string {
	return `<?xml version="1.0" encoding="UTF-8"?>
<rss xmlns:dc="http://purl.org/dc/elements/1.1/" version="2.0">
  <channel>
` + items + `
  </channel>
</rss>`
}

func singleItem(title, link, pubDate, creator, category, description string) string {
	return `<item>
  <title>` + title + `</title>
  <link>` + link + `</link>
  <pubDate>` + pubDate + `</pubDate>
  <dc:creator><![CDATA[` + creator + `]]></dc:creator>
  <category><![CDATA[` + category + `]]></category>
  <description><![CDATA[` + description + `]]></description>
</item>`
}

func newTestClient(ts *httptest.Server) *infoq.Client {
	cfg := infoq.DefaultConfig()
	cfg.BaseURL = ts.URL
	cfg.Rate = 0
	return infoq.NewClient(cfg)
}

func TestLatestParsesTitle(t *testing.T) {
	body := rssXML(singleItem(
		"Microservices at Scale",
		"https://www.infoq.com/news/2024/01/microservices-scale",
		"Mon, 15 Jan 2024 12:00:00 GMT",
		"Jane Smith",
		"Architecture & Design",
		"<p>A short summary about microservices.</p>",
	))
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(body))
	}))
	defer ts.Close()

	arts, err := newTestClient(ts).Latest(context.Background(), 0)
	if err != nil {
		t.Fatal(err)
	}
	if len(arts) != 1 {
		t.Fatalf("got %d articles, want 1", len(arts))
	}
	if arts[0].Title != "Microservices at Scale" {
		t.Errorf("Title = %q", arts[0].Title)
	}
}

func TestLatestParsesAuthor(t *testing.T) {
	body := rssXML(singleItem(
		"Go 1.22 Features",
		"https://www.infoq.com/news/2024/01/go-1-22",
		"Wed, 10 Jan 2024 09:00:00 GMT",
		"John Doe",
		"Development",
		"<p>Body text.</p>",
	))
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(body))
	}))
	defer ts.Close()

	arts, err := newTestClient(ts).Latest(context.Background(), 0)
	if err != nil {
		t.Fatal(err)
	}
	if arts[0].Author != "John Doe" {
		t.Errorf("Author = %q", arts[0].Author)
	}
}

func TestLatestParsesURL(t *testing.T) {
	wantURL := "https://www.infoq.com/news/2024/01/devops-pipeline"
	body := rssXML(singleItem(
		"DevOps Pipeline Guide",
		wantURL,
		"Fri, 12 Jan 2024 15:30:00 GMT",
		"Alice Green",
		"DevOps",
		"<p>Summary here.</p>",
	))
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(body))
	}))
	defer ts.Close()

	arts, err := newTestClient(ts).Latest(context.Background(), 0)
	if err != nil {
		t.Fatal(err)
	}
	if arts[0].URL != wantURL {
		t.Errorf("URL = %q, want %q", arts[0].URL, wantURL)
	}
}

func TestLatestParsesDate(t *testing.T) {
	body := rssXML(singleItem(
		"Security Update",
		"https://www.infoq.com/news/2024/03/security-update",
		"Thu, 07 Mar 2024 18:00:00 GMT",
		"Bob Security",
		"Security",
		"<p>Details.</p>",
	))
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(body))
	}))
	defer ts.Close()

	arts, err := newTestClient(ts).Latest(context.Background(), 0)
	if err != nil {
		t.Fatal(err)
	}
	if arts[0].Published != "2024-03-07" {
		t.Errorf("Published = %q, want %q", arts[0].Published, "2024-03-07")
	}
}

func TestLatestStripsSummaryHTML(t *testing.T) {
	body := rssXML(singleItem(
		"Cloud Native Update",
		"https://www.infoq.com/news/2024/01/cloud-native",
		"Sat, 20 Jan 2024 10:00:00 GMT",
		"Carol Dev",
		"Architecture & Design",
		"<p>This is the <b>summary</b> text.</p>",
	))
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(body))
	}))
	defer ts.Close()

	arts, err := newTestClient(ts).Latest(context.Background(), 0)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(arts[0].Summary, "<") || strings.Contains(arts[0].Summary, ">") {
		t.Errorf("Summary contains HTML tags: %q", arts[0].Summary)
	}
	if !strings.Contains(arts[0].Summary, "summary") {
		t.Errorf("Summary text missing: %q", arts[0].Summary)
	}
}

func TestLatestTruncatesSummary(t *testing.T) {
	long := strings.Repeat("x", 300)
	body := rssXML(singleItem(
		"Long Article",
		"https://www.infoq.com/news/2024/01/long",
		"Mon, 01 Jan 2024 00:00:00 GMT",
		"Author Name",
		"Development",
		"<p>"+long+"</p>",
	))
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(body))
	}))
	defer ts.Close()

	arts, err := newTestClient(ts).Latest(context.Background(), 0)
	if err != nil {
		t.Fatal(err)
	}
	runes := []rune(arts[0].Summary)
	if len(runes) > 150 {
		t.Errorf("Summary too long: %d runes", len(runes))
	}
	if !strings.HasSuffix(arts[0].Summary, "…") {
		t.Errorf("Summary missing ellipsis: %q", arts[0].Summary)
	}
}

func TestLatestRankOrder(t *testing.T) {
	items := singleItem("A", "https://www.infoq.com/a", "Mon, 01 Jan 2024 00:00:00 GMT", "X", "Development", "") +
		singleItem("B", "https://www.infoq.com/b", "Tue, 02 Jan 2024 00:00:00 GMT", "Y", "Development", "") +
		singleItem("C", "https://www.infoq.com/c", "Wed, 03 Jan 2024 00:00:00 GMT", "Z", "Development", "")
	body := rssXML(items)
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(body))
	}))
	defer ts.Close()

	arts, err := newTestClient(ts).Latest(context.Background(), 0)
	if err != nil {
		t.Fatal(err)
	}
	if len(arts) != 3 {
		t.Fatalf("got %d articles, want 3", len(arts))
	}
	for i, a := range arts {
		if a.Rank != i+1 {
			t.Errorf("arts[%d].Rank = %d, want %d", i, a.Rank, i+1)
		}
	}
}

func TestLatestLimit(t *testing.T) {
	items := ""
	for i := 0; i < 5; i++ {
		items += singleItem("T", "https://www.infoq.com/t", "Mon, 01 Jan 2024 00:00:00 GMT", "A", "Development", "")
	}
	body := rssXML(items)
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(body))
	}))
	defer ts.Close()

	arts, err := newTestClient(ts).Latest(context.Background(), 2)
	if err != nil {
		t.Fatal(err)
	}
	if len(arts) != 2 {
		t.Errorf("got %d articles with limit=2, want 2", len(arts))
	}
}

func TestFeedUnknownSection(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	_, err := newTestClient(ts).Feed(context.Background(), "nonexistent", 0)
	if !errors.Is(err, infoq.ErrUnknownSection) {
		t.Errorf("got %v, want ErrUnknownSection", err)
	}
}

func TestFeedKnownSection(t *testing.T) {
	body := rssXML(singleItem(
		"Architecting for Resilience",
		"https://www.infoq.com/news/2024/01/architecting-resilience",
		"Fri, 05 Jan 2024 08:00:00 GMT",
		"Arch Author",
		"Architecture & Design",
		"<p>Architecture summary.</p>",
	))
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(body))
	}))
	defer ts.Close()

	arts, err := newTestClient(ts).Feed(context.Background(), "architecture-design", 0)
	if err != nil {
		t.Fatal(err)
	}
	if len(arts) != 1 {
		t.Fatalf("got %d articles, want 1", len(arts))
	}
	if arts[0].Title != "Architecting for Resilience" {
		t.Errorf("Title = %q", arts[0].Title)
	}
}

func TestSearchFiltersResults(t *testing.T) {
	items := singleItem("Kubernetes Networking", "https://www.infoq.com/k8s", "Mon, 01 Jan 2024 00:00:00 GMT", "A", "DevOps", "<p>k8s stuff</p>") +
		singleItem("Java Spring Boot", "https://www.infoq.com/spring", "Tue, 02 Jan 2024 00:00:00 GMT", "B", "Development", "<p>spring stuff</p>") +
		singleItem("Kubernetes Security", "https://www.infoq.com/k8s-sec", "Wed, 03 Jan 2024 00:00:00 GMT", "C", "Security", "<p>k8s security</p>")
	body := rssXML(items)
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(body))
	}))
	defer ts.Close()

	arts, err := newTestClient(ts).Search(context.Background(), "kubernetes", 0)
	if err != nil {
		t.Fatal(err)
	}
	if len(arts) != 2 {
		t.Errorf("got %d articles for 'kubernetes', want 2", len(arts))
	}
}

func TestSearchLimit(t *testing.T) {
	items := ""
	for i := 0; i < 5; i++ {
		items += singleItem("AI Article", "https://www.infoq.com/ai", "Mon, 01 Jan 2024 00:00:00 GMT", "A", "AI, ML & Data Engineering", "<p>ai content</p>")
	}
	body := rssXML(items)
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(body))
	}))
	defer ts.Close()

	arts, err := newTestClient(ts).Search(context.Background(), "ai", 2)
	if err != nil {
		t.Fatal(err)
	}
	if len(arts) != 2 {
		t.Errorf("got %d articles with limit=2, want 2", len(arts))
	}
}

func TestSectionsReturnsAll(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(rssXML("")))
	}))
	defer ts.Close()

	secs := newTestClient(ts).Sections()
	if len(secs) == 0 {
		t.Fatal("Sections returned empty list")
	}
	for i, s := range secs {
		if s.Rank != i+1 {
			t.Errorf("secs[%d].Rank = %d, want %d", i, s.Rank, i+1)
		}
		if s.Name == "" {
			t.Errorf("secs[%d].Name is empty", i)
		}
		if !strings.HasPrefix(s.URL, ts.URL) {
			t.Errorf("secs[%d].URL = %q, should start with test server URL", i, s.URL)
		}
	}
}

func TestGetRetriesOn503(t *testing.T) {
	var hits int
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits++
		if hits < 3 {
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}
		_, _ = w.Write([]byte(rssXML("")))
	}))
	defer ts.Close()

	cfg := infoq.DefaultConfig()
	cfg.BaseURL = ts.URL
	cfg.Rate = 0
	cfg.Retries = 5
	c := infoq.NewClient(cfg)

	start := time.Now()
	_, err := c.Latest(context.Background(), 0)
	if err != nil {
		t.Fatal(err)
	}
	if hits != 3 {
		t.Errorf("server saw %d hits, want 3", hits)
	}
	if time.Since(start) < 500*time.Millisecond {
		t.Error("retries did not back off")
	}
}

func TestGetUserAgent(t *testing.T) {
	var gotUA string
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotUA = r.Header.Get("User-Agent")
		_, _ = w.Write([]byte(rssXML("")))
	}))
	defer ts.Close()

	cfg := infoq.DefaultConfig()
	cfg.BaseURL = ts.URL
	cfg.Rate = 0
	c := infoq.NewClient(cfg)
	_, _ = c.Latest(context.Background(), 0)

	if gotUA == "" {
		t.Error("request carried no User-Agent")
	}
}
