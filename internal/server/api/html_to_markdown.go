package api

import (
	"bytes"
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"
	"golang.org/x/net/html"
)

// HTMLRequest is the structure for the incoming HTML request
type HTMLRequest struct {
	HTML string `json:"html"`
}

// MarkdownResponse is the structure for the outgoing Markdown response
type MarkdownResponse struct {
	Markdown string `json:"markdown"`
}

// ConvertHTMLToMarkdown handles the API endpoint for HTML to Markdown conversion
func (a *API) ConvertHTMLToMarkdown(c echo.Context) error {
	// Get the file from the request
	file, err := c.FormFile("html")
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "No HTML file provided").WithInternal(err)
	}

	// Open the file
	src, err := file.Open()
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to open HTML file").WithInternal(err)
	}
	defer src.Close()

	// Read the file content
	buf := new(bytes.Buffer)
	if _, err := buf.ReadFrom(src); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to read HTML file").WithInternal(err)
	}

	htmlContent := buf.String()
	if htmlContent == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "HTML content is required")
	}

	markdown, err := HTMLToMarkdown(htmlContent)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to convert HTML to Markdown").WithInternal(err)
	}

	return c.JSON(http.StatusOK, MarkdownResponse{
		Markdown: markdown,
	})
}

// HTMLToMarkdown converts HTML to Markdown
func HTMLToMarkdown(htmlCode string) (string, error) {
	doc, err := html.Parse(strings.NewReader(htmlCode))
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	processNode(&buf, doc, 0)
	return buf.String(), nil
}

func processNode(buf *bytes.Buffer, n *html.Node, depth int) {
	if n.Type == html.TextNode {
		text := strings.TrimSpace(n.Data)
		if text != "" {
			buf.WriteString(text)
		}
		return
	}

	if n.Type == html.ElementNode {
		switch n.Data {
		case "style", "script", "meta", "link":
		// Ignore these elements
		case "title", "h1":
			writeWrappedContent(buf, n, "\n\n# ", "\n", depth)
		case "h2":
			writeWrappedContent(buf, n, "\n\n## ", "\n", depth)
		case "h3":
			writeWrappedContent(buf, n, "\n\n### ", "\n", depth)
		case "h4":
			writeWrappedContent(buf, n, "\n\n#### ", "\n", depth)
		case "h5":
			writeWrappedContent(buf, n, "\n\n##### ", "\n", depth)
		case "h6":
			writeWrappedContent(buf, n, "\n\n###### ", "\n", depth)
		case "em", "i":
			writeWrappedContent(buf, n, " *", "* ", depth)
		case "strong", "b":
			writeWrappedContent(buf, n, " **", "** ", depth)
		case "code":
			writeWrappedContent(buf, n, " `", "` ", depth)
		case "pre":
			writeWrappedContent(buf, n, "\n```\n", "\n```\n", depth)
		case "a":
			buf.WriteString("[")
			processChildren(buf, n, depth+1)
			buf.WriteString("](")
			for _, a := range n.Attr {
				if a.Key == "href" {
					buf.WriteString(a.Val)
					break
				}
			}
			buf.WriteString(")")
		case "img":
			buf.WriteString("![")
			alt := ""
			src := ""
			for _, a := range n.Attr {
				if a.Key == "alt" {
					alt = a.Val
				} else if a.Key == "src" {
					src = a.Val
				}
			}
			buf.WriteString(alt)
			buf.WriteString("](")
			buf.WriteString(src)
			buf.WriteString(")")
		case "ul":
			if depth > 0 {
				buf.WriteString("\n")
			}
			processListItems(buf, n, depth, "*")
		case "ol":
			if depth > 0 {
				buf.WriteString("\n")
			}
			processListItems(buf, n, depth, "1.")
		case "blockquote":
			buf.WriteString("\n> ")
			processChildren(buf, n, depth+1)
			buf.WriteString("\n")
		case "p":
			processChildren(buf, n, depth+1)
			if n.NextSibling != nil && (n.NextSibling.Type == html.ElementNode && n.NextSibling.Data != "p") {
				buf.WriteString("\n")
			}
		case "br":
			buf.WriteString("\n")
		case "hr":
			buf.WriteString("\n---\n")
		case "div":
			processChildren(buf, n, depth+1)
			if n.NextSibling != nil {
				buf.WriteString("\n")
			}
		default:
			processChildren(buf, n, depth+1)
		}
	}

	// Process all other siblings in the document
	if n.Type == html.DocumentNode {
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			processNode(buf, c, depth+1)
		}
	}
}

func writeWrappedContent(buf *bytes.Buffer, n *html.Node, prefix, suffix string, depth int) {
	if buf.Len() == 0 {
		prefix = strings.TrimLeft(prefix, "\n")
	}
	buf.WriteString(prefix)
	processChildren(buf, n, depth+1)
	buf.WriteString(suffix)
}

func processChildren(buf *bytes.Buffer, n *html.Node, depth int) {
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		processNode(buf, c, depth)
	}
}

func processListItems(buf *bytes.Buffer, n *html.Node, depth int, marker string) {
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		if c.Type == html.ElementNode && c.Data == "li" {
			buf.WriteString("\n")
			buf.WriteString(marker)
			buf.WriteString(" ")
			processChildren(buf, c, depth+1)
		}
	}
	buf.WriteString("\n")
}
