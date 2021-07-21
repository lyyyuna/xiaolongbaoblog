package site

import (
	"bytes"
	"log"
	"os"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/renderer/html"
)

type About struct {
	about       []byte
	htmlContent string
}

func NewAbout(path string) *About {
	rawContent, err := os.ReadFile(path)
	if err != nil {
		log.Fatalf("fail to read about me file: %v", err)
	}
	return &About{
		about: rawContent,
	}
}

func (a *About) Parse() {
	// init markdown parser
	markdown := goldmark.New(
		goldmark.WithExtensions(),
		goldmark.WithRendererOptions(
			html.WithUnsafe(),
		),
	)
	var buf bytes.Buffer
	err := markdown.Convert(a.about, &buf)
	if err != nil {
		log.Fatalf("fail to transform markdown to raw html: %v", err)
	}
	a.htmlContent = buf.String()
}
