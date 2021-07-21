package site

import (
	"log"
	"os"
	"path/filepath"
	"sort"
	"sync"

	"github.com/lyyyuna/xiaolongbaoblog/pkg/config"
	"github.com/panjf2000/ants/v2"
)

type Site struct {
	Blogs      []*Blog
	Categories map[string][]*Blog
	Series     map[string][]*Blog

	conf    *config.Config
	AboutMe *About
}

func NewGenerate(conf *config.Config) *Site {
	postDir := filepath.Join(".", conf.SourceDir, conf.PostDir)
	posts, err := os.ReadDir(postDir)
	if err != nil {
		log.Fatalf("fail to read the posts from the dir: %v", err)
	}

	// parse markdown
	blogs := make([]*Blog, 0)
	defer ants.Release()
	var wg sync.WaitGroup
	var m sync.Mutex

	for _, post := range posts {
		if post.IsDir() {
			continue
		}

		wg.Add(1)
		fname := post.Name()
		ants.Submit(func() {
			blog := NewBlog(filepath.Join(postDir, fname))
			m.Lock()
			defer m.Unlock()
			blogs = append(blogs, blog)
			wg.Done()
		})
	}

	wg.Wait()

	sort.Sort(byBlogDate(blogs))

	s := &Site{
		Blogs:      blogs,
		Categories: make(map[string][]*Blog),
		Series:     make(map[string][]*Blog),
		conf:       conf,
		AboutMe:    NewAbout(filepath.Join(".", conf.SourceDir, conf.AboutDir, "index.md")),
	}

	// get categories
	for _, blog := range blogs {
		category := blog.Meta.Category
		_, ok := s.Categories[category]
		if !ok {
			s.Categories[category] = []*Blog{blog}
		} else {
			s.Categories[category] = append(s.Categories[category], blog)
		}
	}

	// get series
	for _, blog := range blogs {
		series := blog.Meta.Series
		if series == "" {
			continue
		}
		_, ok := s.Series[series]
		if !ok {
			s.Series[series] = []*Blog{blog}
		} else {
			s.Series[series] = append(s.Series[series], blog)
		}
	}

	// get about me
	s.AboutMe.Parse()

	return s
}

func (s *Site) Output() {
	outputDir := filepath.Join(".", s.conf.OutputDir)
	err := os.RemoveAll(outputDir)
	if err != nil {
		log.Fatalf("fail to remove the output dir: %v", err)
	}
	err = os.MkdirAll(outputDir, os.ModePerm)
	if err != nil {
		log.Fatalf("fail to create the output dir: %v", err)
	}
	// write index
	s.outputIndex(outputDir)
	// write each blog
	s.outputPost(outputDir)
	// write static files
	s.outputStatic(outputDir)
	// write series
	s.outputSeries(outputDir)
	// write about me
	s.outputAboutMe(outputDir)
	// write atom/rss feed
	s.outputAtom(outputDir)
	// write sitemap
	s.outputSitemap(outputDir)
}

type byBlogDate []*Blog

func (b byBlogDate) Len() int {
	return len(b)
}

func (b byBlogDate) Swap(i, j int) {
	b[i], b[j] = b[j], b[i]
}

func (b byBlogDate) Less(i, j int) bool {
	return b[i].Meta.Date.After(b[j].Meta.Date)
}

type byBlogDateAsc []*Blog

func (b byBlogDateAsc) Len() int {
	return len(b)
}

func (b byBlogDateAsc) Swap(i, j int) {
	b[i], b[j] = b[j], b[i]
}

func (b byBlogDateAsc) Less(i, j int) bool {
	return b[i].Meta.Date.Before(b[j].Meta.Date)
}
