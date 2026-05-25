package main

import (
	"context"
	"fmt"
	"html"
	"io"
	"log"
	"net/http"
	"regexp"
	"strings"
	"time"

	htmltomarkdown "github.com/JohannesKaufmann/html-to-markdown/v2"
	"github.com/gocolly/colly/v2"
)

// markdownLinkRegexp matches [text](url) patterns (no newlines inside)
var markdownLinkRegexp = regexp.MustCompile(`\[([^\]]+)\]\((https?://[^)]+)\)`)

// expandLinksInMarkdown takes a [text](url) pattern and, when the link text
// does not already contain the URL itself, appends the bare URL after the
// reference so the actual address is visible in rendered/plain-text output.
func expandLinksInMarkdown(md string) string {
	return markdownLinkRegexp.ReplaceAllStringFunc(md, func(match string) string {
		textStart := strings.IndexByte(match, '[') + 1
		textEnd := strings.IndexByte(match[textStart:], ']')
		text := match[textStart : textStart+textEnd]
		urlStart := strings.IndexByte(match[textStart+textEnd:], '(') + textStart + textEnd
		urlEnd := strings.LastIndexByte(match, ')')
		url := match[urlStart:urlEnd]

		// Wenn der Link-Text schon die URL enthält (z.B. [https://example.com](https://example.com)),
		// belassen wir den Link unverändert.
		if strings.Contains(strings.ToLower(text), strings.ToLower(url)) {
			return match
		}

		// Sonst: [text](url) → [text](url) - `https://...`
		return fmt.Sprintf("%s - `%s`", match, url)
	})
}

// ─── Web-Crawling ────────────────────────────────────────────────────────────

// crawlPage crawlt eine Webseite mit Colly und konvertiert HTML → Markdown.
// Reine-Go-Lösung – kein externer Server erforderlich.
func (a *App) crawlPage(url string) (string, error) {
	host := parseHostname(url)
	c := colly.NewCollector(
		colly.AllowedDomains(host),
		colly.IgnoreRobotsTxt(),
		colly.AllowURLRevisit(),
	)

	var rawHTML string
	var pageTitle string
	var firstErr error

	// Seitentitel extrahieren
	c.OnHTML("title", func(e *colly.HTMLElement) {
		pageTitle = e.Text
	})

	// Inhaltsbereiche selektieren – Reihenfolge der Priorität
	contentSelectors := "main, article, .post-content, .entry-content, #content, body"
	c.OnHTML(contentSelectors, func(e *colly.HTMLElement) {
		if rawHTML == "" {
			rawHTML = e.ChildText("*") // Fallback: reiner Text
			// Versuch, das volle HTML über die goquery-DOM zu bekommen
			if h, err := e.DOM.Html(); err == nil && h != "" {
				rawHTML = h
			}
		}
	})

	c.OnError(func(r *colly.Response, err error) {
		if firstErr == nil {
			firstErr = fmt.Errorf("crawl error on %s: %w", r.Request.URL.String(), err)
		}
	})

	if err := c.Visit(url); err != nil {
		return "", fmt.Errorf("visit failed: %w", err)
	}

	if firstErr != nil {
		return "", firstErr
	}

	// Fallback: wenn Colly nichts extrahieren konnte, direkter HTTP-GET
	if rawHTML == "" {
		rawHTML = a.fetchRawPage(url)
	}

	if rawHTML == "" {
		return "", fmt.Errorf("empty content extracted from %s", url)
	}

	// HTML → Markdown konvertieren
	markdown, err := htmltomarkdown.ConvertString(rawHTML)
	if err != nil {
		return "", fmt.Errorf("html-to-markdown: %w", err)
	}

	// Links aufklappen: URL sichtbar machen, wenn Link-Text sie nicht enthält
	markdown = expandLinksInMarkdown(markdown)

	// Seitentitel als Überschrift voranstellen, falls nicht vorhanden
	if pageTitle != "" && !strings.Contains(markdown, "# "+pageTitle) {
		markdown = "# " + pageTitle + "\n\n" + markdown
	}

	log.Printf("[crawl] %s → %d bytes markdown", url, len(markdown))
	return strings.TrimSpace(markdown), nil
}

// fetchRawPage ist ein einfacher HTTP-GET-Fallback, wenn Colly keinen Inhalt liefert.
func (a *App) fetchRawPage(url string) string {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(a.cfg.CrawlTimeoutSec)*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return ""
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (compatible; FoundAItionBot/1.0)")

	client := &http.Client{
		Timeout: 0,
		CheckRedirect: func(r *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
	resp, err := client.Do(req)
	if err != nil {
		return ""
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil || len(body) == 0 {
		return ""
	}

	markdown, err := htmltomarkdown.ConvertString(string(body))
	if err != nil {
		return string(body)
	}
	return expandLinksInMarkdown(markdown)
}

// ─── URL-Hilfsfunktionen ─────────────────────────────────────────────────────

// parseHostname extrahiert den Hostnamen aus einer URL (ohne net/url-Dependency).
func parseHostname(rawURL string) string {
	s := rawURL
	if strings.HasPrefix(s, "http://") {
		s = s[7:]
	} else if strings.HasPrefix(s, "https://") {
		s = s[8:]
	}
	idx := strings.IndexByte(s, '/')
	if idx >= 0 {
		s = s[:idx]
	}
	// Port für AllowedDomains entfernen
	colon := strings.IndexByte(s, ':')
	if colon > 0 {
		s = s[:colon]
	}
	return s
}

// extractTitleFromMarkdown gibt die erste #-Überschrift aus Markdown-Text zurück.
func extractTitleFromMarkdown(md string) string {
	lines := strings.Split(md, "\n")
	for _, line := range lines {
		t := strings.TrimSpace(line)
		if strings.HasPrefix(t, "# ") && !strings.HasPrefix(t, "##") {
			return strings.TrimPrefix(t, "# ")
		}
	}
	return ""
}

// extractYouTubePageTitle holt den Originaltitel einer YouTube-Seite
// durch einfaches Parsen des HTML <title>-Tags. Als Fallback wenn yt-dlp versagt.
func extractYouTubePageTitle(url string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (compatible; FoundAItionBot/1.0)")

	resp, err := (&http.Client{Timeout: 15 * time.Second}).Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	// <title>…</title> im HTML extrahieren
	start := strings.Index(string(body), "<title>")
	end := strings.Index(string(body), "</title>")
	if start == -1 || end == -1 || end <= start {
		return "", fmt.Errorf("no <title> tag found")
	}

	title := string(body[start+7 : end])
	// HTML-Entitäten dekodieren (&quot; → ", &amp; → & etc.)
	title = html.UnescapeString(title)
	// YouTube hängt oft " - YouTube" an, das entfernen
	title = strings.TrimSuffix(title, " - YouTube")
	title = strings.TrimSpace(title)
	log.Printf("[yt-title] extracted from HTML: %q", title)
	return title, nil
}
