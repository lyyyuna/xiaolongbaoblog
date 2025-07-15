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

	"regexp"

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

	dirSeries      string
	rewriteImgPath map[string]string
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

func NewBlog(path string, dirSeries string) *Blog {
	rawContent, err := os.ReadFile(path)
	if err != nil {
		log.Fatalf("fail to read blog post file: %v", err)
	}

	uri := strings.Split(filepath.Base(path), ".")[0]
	b := &Blog{
		rawContent:     rawContent,
		filePath:       path,
		Uri:            uri,
		dirSeries:      dirSeries,
		rewriteImgPath: make(map[string]string),
	}
	b.parse()
	return b
}

func (b *Blog) parse() {
	b.parseMetaData()
	b.rewriteSeries()
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

	if b.dirSeries != "" {
		meta.SeriesName = b.dirSeries
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

func (b *Blog) rewriteSeries() {
	// 只替换目录 series 类型
	if b.dirSeries == "" {
		return
	}

	// 正则表达式匹配 Markdown 图片语法 ![alt](src)
	re := regexp.MustCompile(`!\[([^\]]*)\]\(([^)]+)\)`)

	b.markdownContent = re.ReplaceAllFunc(b.markdownContent, func(match []byte) []byte {
		parts := re.FindStringSubmatch(string(match))
		if len(parts) < 3 {
			return match
		}

		altText := parts[1]
		imgSrc := parts[2]

		oriImgSrc := b.dirSeries + "/" + imgSrc

		// 如果图片路径是相对路径（以 ./ 或 ../ 开头，或者不是 http:// 或 https:// 开头）
		if !strings.HasPrefix(imgSrc, "http://") &&
			!strings.HasPrefix(imgSrc, "https://") &&
			!strings.HasPrefix(imgSrc, "/") {

			// 确保路径不以 / 开头
			if after, ok := strings.CutPrefix(imgSrc, "./"); ok {
				imgSrc = after
			}

			// 添加基础 URL
			imgSrc = "/imgrrr/" + b.dirSeries + "/" + imgSrc
			b.rewriteImgPath[oriImgSrc] = imgSrc
		}

		return []byte(fmt.Sprintf("![%s](%s)", altText, imgSrc))
	})
}
