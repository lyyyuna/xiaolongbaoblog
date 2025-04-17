package site

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/renderer/html"
	"gopkg.in/yaml.v2"
)

type Blog struct {
	Meta BlogMeta

	filePath string
	Uri      string

	rawContent      []byte
	markdownContent []byte
	htmlContent     string
}

type BlogMeta struct {
	Title      string    `yaml:"title"`
	DateS      string    `yaml:"date"`
	Category   string    `yaml:"categories"`
	SeriesName string    `yaml:"series"`
	Date       time.Time `yaml:"-"`
	DateT      string    `yaml:"-"`
	IsSeries   bool      `yaml:"-"`
	Summary    string    `yaml:"summary"`
	MathJax    bool      `yaml:"mathjax"`
}

func NewBlog(path string) *Blog {
	rawContent, err := os.ReadFile(path)
	if err != nil {
		log.Fatalf("fail to read blog post file: %v", err)
	}

	uri := strings.Split(filepath.Base(path), ".")[0]
	b := &Blog{
		rawContent: rawContent,
		filePath:   path,
		Uri:        uri,
	}
	b.parse()
	return b
}

func (b *Blog) parse() {
	b.parseMetaData()
	b.parseMarkdown()
}

func (b *Blog) parseMetaData() {
	directives := bytes.Split(b.rawContent, []byte{'-', '-', '-'})
	if len(directives) < 2 {
		log.Fatalf("the blog: %v is invalid, should contain only 2 parts", b.filePath)
	}

	meta := BlogMeta{}
	err := yaml.Unmarshal(directives[0], &meta)
	if err != nil {
		log.Fatalf("fail to parse: %v blog meta", b.filePath)
	}

	if meta.DateS == "" || meta.Title == "" {
		log.Fatalf("some blog meta is empty: %v", b.filePath)
	}

	if meta.SeriesName == "" {
		meta.IsSeries = false
	} else {
		meta.IsSeries = true
	}

	meta.Date, err = time.Parse("2006-01-02 15:04:05", meta.DateS)
	if err != nil {
		log.Fatalf("fail to parse the date: %v", b.filePath)
	}

	year := meta.Date.Year()
	month := meta.Date.Month()
	day := meta.Date.Day()
	url := fmt.Sprintf("/%04d/%02d/%02d", year, month, day)
	b.Uri = path.Join(url, b.Uri) + "/"

	meta.DateT = fmt.Sprintf("%04d-%02d", year, month)

	b.Meta = meta
	b.markdownContent = bytes.Join(directives[1:], nil)
}

func (b *Blog) parseMarkdown() {
	// init markdown parser
	markdown := goldmark.New(
		goldmark.WithExtensions(extension.Table),
		goldmark.WithRendererOptions(
			html.WithUnsafe(),
		),
	)
	var buf bytes.Buffer
	err := markdown.Convert(b.markdownContent, &buf)
	if err != nil {
		log.Fatalf("fail to transform markdown to raw html: %v", err)
	}
	b.htmlContent = buf.String()
}

func (b *Blog) String() string {
	return fmt.Sprintf("meta: %v, filepath: %v, uri: %v, html: %v", b.Meta, b.filePath, b.Uri, b.htmlContent)
}
