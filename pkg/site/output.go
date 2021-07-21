package site

import (
	"embed"
	"log"
	"os"
	"path/filepath"
	"sort"
	"text/template"
	"time"

	"github.com/gorilla/feeds"
	"github.com/otiai10/copy"
	"github.com/snabb/sitemap"
)

//go:embed tpl/index.tpl
var indexContent embed.FS

//go:embed tpl/post.tpl
var postContent embed.FS

//go:embed tpl/about.tpl
var aboutContent embed.FS

//go:embed tpl/series.tpl
var seriesContent embed.FS

func (s *Site) outputIndex(path string) {
	indexTmpl, err := template.ParseFS(indexContent, "tpl/index.tpl")
	if err != nil {
		log.Fatalf("fail to parse the index template: %v", err)
	}

	indexPath := filepath.Join(path, "index.html")
	indexF, err := os.Create(indexPath)
	if err != nil {
		log.Fatalf("fail to create the index file: %v", err)
	}
	defer indexF.Close()

	data := struct {
		Blogs     []*Blog
		Title     string
		Author    string
		Url       string
		SubTitle  string
		SeriesDir string
		Analytics string
	}{
		Blogs:     s.Blogs,
		Title:     s.conf.Title,
		Author:    s.conf.Author,
		Url:       s.conf.Url,
		SubTitle:  s.conf.SubTitle,
		SeriesDir: s.conf.SeriesDir,
		Analytics: s.conf.Analytics,
	}
	if err := indexTmpl.Execute(indexF, data); err != nil {
		log.Fatalf("fail to render the index template: %v", err)
	}
	log.Println("生成首页")
}

func (s *Site) outputPost(path string) {
	postTmpl, err := template.ParseFS(postContent, "tpl/post.tpl")
	if err != nil {
		log.Fatalf("fail to parse the post template: %v", err)
	}

	for _, blog := range s.Blogs {
		postPath := filepath.Join(path, blog.Uri)
		err := os.MkdirAll(postPath, os.ModePerm)
		if err != nil {
			log.Fatalf("fail to create post path in %v, err: %v", postPath, err)
		}
		postF, err := os.Create(filepath.Join(postPath, "index.html"))
		if err != nil {
			log.Fatalf("fail to create post in %v, err: %v", postPath, err)
		}
		defer postF.Close()

		var seriesIndex int
		if blog.Meta.IsSeries {
			for _, b := range s.Series[blog.Meta.Series] {
				if b == blog {
					break
				}
				seriesIndex++
			}
		}
		seriesIndex = len(s.Series[blog.Meta.Series]) - seriesIndex
		data := struct {
			SiteTitle    string
			SiteSubTitle string
			Author       string
			Title        string
			IsSeries     bool
			Series       string
			When         string
			Body         string
			SeriesIndex  int
			MathJax      bool
			SeriesDir    string
			Analytics    string
		}{
			SiteTitle:    s.conf.Title,
			SiteSubTitle: s.conf.SubTitle,
			Author:       s.conf.Author,
			Title:        blog.Meta.Title,
			IsSeries:     blog.Meta.IsSeries,
			Series:       blog.Meta.Series,
			When:         blog.Meta.DateT,
			Body:         blog.htmlContent,
			SeriesIndex:  seriesIndex,
			MathJax:      blog.Meta.MathJax,
			SeriesDir:    s.conf.SeriesDir,
			Analytics:    s.conf.Analytics,
		}

		if err := postTmpl.Execute(postF, data); err != nil {
			log.Fatalf("fail to render the post template: %v, err: %v", blog.filePath, err)
		}

		log.Printf("生成: %v, 链接: %v", blog.Meta.Title, blog.Uri)
	}
}

func (s *Site) outputStatic(path string) {
	excludes := make(map[string]int)
	excludes[s.conf.PostDir] = 1
	excludes[s.conf.AboutDir] = 1

	sourceDir := filepath.Join(".", s.conf.SourceDir)
	sourceEs, err := os.ReadDir(sourceDir)
	if err != nil {
		log.Fatalf("fail to read the posts from the dir: %v", err)
	}

	for _, entry := range sourceEs {
		if _, ok := excludes[entry.Name()]; ok {
			continue
		}

		src := filepath.Join(sourceDir, entry.Name())
		dst := filepath.Join(".", s.conf.OutputDir, entry.Name())

		copy.Copy(src, dst)

		log.Printf("拷贝静态资源：%v", entry.Name())
	}
}

func (s *Site) outputSeries(path string) {
	seriesTmpl, err := template.ParseFS(seriesContent, "tpl/series.tpl")
	if err != nil {
		log.Fatalf("fail to parse the series template: %v", err)
	}

	seriesDir := filepath.Join(".", s.conf.OutputDir, s.conf.SeriesDir)

	for k, seriesBlogs := range s.Series {
		thisSeriesDir := filepath.Join(seriesDir, k)
		err := os.MkdirAll(thisSeriesDir, os.ModePerm)
		if err != nil {
			log.Fatalf("fail to create series dir: %v, err: %v", thisSeriesDir, err)
		}

		thisSeriesF, err := os.Create(filepath.Join(thisSeriesDir, "index.html"))
		if err != nil {
			log.Fatalf("fail to create series index: %v, err: %v", thisSeriesDir, err)
		}
		defer thisSeriesF.Close()

		sort.Sort(byBlogDateAsc(seriesBlogs))
		data := struct {
			SiteTitle    string
			SiteSubTitle string
			Author       string
			SeriesTitle  string
			Blogs        []*Blog
			Analytics    string
		}{
			SiteTitle:    s.conf.Title,
			SiteSubTitle: s.conf.SubTitle,
			Author:       s.conf.Author,
			SeriesTitle:  k,
			Blogs:        seriesBlogs,
			Analytics:    s.conf.Analytics,
		}

		if err := seriesTmpl.Execute(thisSeriesF, data); err != nil {
			log.Fatalf("fail to render the series template: %v, err: %v", thisSeriesDir, err)
		}

		log.Printf("生成系列页面: %v", k)
	}
}

func (s *Site) outputAboutMe(path string) {
	aboutTmpl, err := template.ParseFS(aboutContent, "tpl/about.tpl")
	if err != nil {
		log.Fatalf("fail to parse the about template: %v", err)
	}

	aboutDir := filepath.Join(".", s.conf.OutputDir, s.conf.AboutDir)
	err = os.MkdirAll(aboutDir, os.ModePerm)
	if err != nil {
		log.Fatalf("fail to create the about dir: %v", err)
	}

	aboutF, err := os.Create(filepath.Join(aboutDir, "index.html"))
	if err != nil {
		log.Fatalf("fail to create the about html: %v", err)
	}
	defer aboutF.Close()

	data := struct {
		SiteTitle    string
		SiteSubTitle string
		Author       string
		Body         string
		Analytics    string
	}{
		SiteTitle:    s.conf.Title,
		SiteSubTitle: s.conf.SubTitle,
		Author:       s.conf.Author,
		Body:         s.AboutMe.htmlContent,
		Analytics:    s.conf.Analytics,
	}

	if err := aboutTmpl.Execute(aboutF, data); err != nil {
		log.Fatalf("fail to render the about template: %v", err)
	}

	log.Println("生成 about me")
}

func (s *Site) outputAtom(path string) {
	now := time.Now()

	feed := &feeds.Feed{
		Title:       s.conf.Title,
		Link:        &feeds.Link{Href: s.conf.Url},
		Description: s.conf.SubTitle,
		Author:      &feeds.Author{Name: s.conf.Author, Email: s.conf.Email},
		Created:     now,
	}

	items := make([]*feeds.Item, 0)
	for _, blog := range s.Blogs {
		item := &feeds.Item{
			Title:       blog.Meta.Title,
			Link:        &feeds.Link{Href: s.conf.Url + blog.Uri},
			Description: blog.Meta.Summary,
			Author:      &feeds.Author{Name: s.conf.Author, Email: s.conf.Email},
			Created:     blog.Meta.Date,
			Updated:     now,
			Content:     blog.htmlContent,
			Id:          s.conf.Url + blog.Uri,
		}
		items = append(items, item)
	}
	feed.Items = items

	atom, err := feed.ToAtom()
	if err != nil {
		log.Fatalf("fail to generate atom xml for this site: %v", err)
	}

	atomF, err := os.Create(filepath.Join(s.conf.OutputDir, "atom.xml"))
	if err != nil {
		log.Fatalf("fail to create atom xml file: %v", err)
	}
	defer atomF.Close()

	_, err = atomF.Write([]byte(atom))
	if err != nil {
		log.Fatalf("fail to write to atom xml file: %v", err)
	}

	log.Println("生成 atom xml")
}

func (s *Site) outputSitemap(path string) {
	sm := sitemap.New()

	for _, blog := range s.Blogs {
		sm.Add(&sitemap.URL{
			Loc:        s.conf.Url + blog.Uri,
			LastMod:    &blog.Meta.Date,
			ChangeFreq: sitemap.Monthly,
		})
	}

	sitemapF, err := os.Create(filepath.Join(path, "sitemap.xml"))
	if err != nil {
		log.Fatalf("fail to create sitemap xml: %v", err)
	}
	defer sitemapF.Close()

	sm.WriteTo(sitemapF)

	log.Println("生成 sitemap")
}
