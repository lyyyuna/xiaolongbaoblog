package site

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"net/url"
)

type indexNowRequest struct {
	Host        string   `json:"host"`
	Key         string   `json:"key"`
	KeyLocation string   `json:"keyLocation"`
	UrlList     []string `json:"urlList"`
}

func (s *Site) SubmitIndexNow() {

	host, _ := url.Parse(s.conf.Url)
	key := s.conf.IndexNow
	keyLocation := s.conf.Url + "/" + key + ".txt"

	updatedUrl := s.conf.Url + s.Blogs[0].Uri

	urlList := []string{updatedUrl, s.conf.Url}

	if s.Blogs[0].Meta.IsSeries {
		seriesUrl := s.conf.Url + "/" + s.conf.SeriesDir + "/" + s.Blogs[0].Meta.SeriesName + "/"
		urlList = append(urlList, seriesUrl)
	}

	requestData := indexNowRequest{
		Host:        host.Host,
		Key:         key,
		KeyLocation: keyLocation,
		UrlList:     urlList,
	}

	jsonData, _ := json.Marshal(requestData)

	resp, err := http.Post(
		"https://api.indexnow.org/IndexNow",
		"application/json; charset=utf-8",
		bytes.NewBuffer(jsonData),
	)
	if err != nil {
		log.Fatalf("发送 indexnow 失败: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		log.Fatalf("indexnow 提交失败, %v: %s", resp.StatusCode, string(body))
	}

	log.Printf("indexnow [%v] 提交成功", urlList)
}
