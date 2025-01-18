package doctree

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"maps"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
)

// Document represents a Markdown document.
type Document struct {
	URL      string
	Content  string
	Links    []string
	LinkedBy []*Document
}

// NewDocument creates a new Document from the given URL.
func NewDocument(url string) (*Document, error) {
	content, err := fetchContent(url)
	if err != nil {
		return nil, err
	}

	doc := &Document{
		URL:      url,
		Content:  content,
		Links:    nil,
		LinkedBy: nil,
	}

	doc.Links = doc.extractLinks()

	return doc, nil
}

// fetchContent reads the content from a URL (local file or HTTP/HTTPS).
func fetchContent(url string) (string, error) {
	if isFileURL(url) {
		return readFile(url)
	}

	return readHTTP(url)
}

// isFileURL checks if the given URL is a local file path.
func isFileURL(url string) bool {
	return strings.HasPrefix(url, "file://") || (!strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://"))
}

// readFile reads content from a local file.
func readFile(url string) (string, error) {
	url = strings.TrimPrefix(url, "file://")

	data, err := os.ReadFile(url)
	if err != nil {
		return "", fmt.Errorf("os read file %q: %w", url, err)
	}

	return string(data), nil
}

// readHTTP reads content from an HTTP/HTTPS URL.
func readHTTP(url string) (string, error) {
	client := http.DefaultClient

	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, url, nil)
	if err != nil {
		return "", fmt.Errorf("new get req: %w", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("get %q: %w", url, err)
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("read response: %w", err)
	}

	return string(data), nil
}

// extractLinks extracts all links from the Markdown content.
func (d *Document) extractLinks() []string {
	const matchGroupCount = 2

	re := regexp.MustCompile(`\[(.*)\]\((.*(\.md|\.txt|\.json))\)`)
	matches := re.FindAllStringSubmatch(d.Content, -1)

	var links []string

	for _, match := range matches {
		if len(match) > matchGroupCount {
			url := match[2]

			if !isFileURL(d.URL) && isFileURL(url) {
				// Processing relative links on remote documents is not trivial.
				// Then skip them.
				continue
			}

			if isFileURL(url) && !strings.HasPrefix(url, "file://") {
				directory := filepath.Dir(strings.TrimPrefix(d.URL, "file://"))
				url = filepath.Join(directory, url)
			}

			links = append(links, url)
		}
	}

	return links
}

// BuildGraph builds a graph of documents based on their links.
func BuildGraph(urls []string) (map[string]*Document, error) {
	documents := make(map[string]*Document)

	for _, url := range urls {
		if _, exists := documents[url]; !exists {
			doc, err := NewDocument(url)
			if err != nil {
				return nil, err
			}

			documents[url] = doc
		}
	}

	for hasNewDoc := true; hasNewDoc; {
		hasNewDoc = false

		for _, doc := range documents {
			for _, link := range doc.Links {
				linkedDoc, exists := documents[link]

				if exists {
					if !slices.ContainsFunc(linkedDoc.LinkedBy, func(d *Document) bool {
						return doc.URL == d.URL
					}) {
						linkedDoc.LinkedBy = append(linkedDoc.LinkedBy, doc)
					}
				} else {
					newDoc, err := NewDocument(link)
					if err != nil {
						slog.Warn("read linked document",
							slog.String("url", link),
							slog.String("error", err.Error()),
						)

						continue
					}

					documents[link] = newDoc
					newDoc.LinkedBy = append(newDoc.LinkedBy, doc)

					hasNewDoc = true
				}
			}
		}
	}

	return documents, nil
}

// FormatGraph reurns formatted string representation of the graph of documents.
func FormatGraph(documents map[string]*Document) string {
	var builder strings.Builder

	urls := slices.Collect(maps.Keys(documents))
	slices.Sort(urls)

	for _, url := range urls {
		doc := documents[url]

		builder.WriteString(fmt.Sprintf("Document: %s\n", url))

		builder.WriteString("Linked By:\n")

		for _, linkedByDoc := range doc.LinkedBy {
			builder.WriteString(fmt.Sprintf("  - %s\n", linkedByDoc.URL))
		}

		builder.WriteString("\n")
	}

	return builder.String()
}
