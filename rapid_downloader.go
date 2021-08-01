package downloader

import (
	"errors"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/canhlinh/log4go"
	"github.com/headzoo/surf"
	"github.com/headzoo/surf/browser"
)

type Rapid struct {
	*Base
	DirectURL string
	Referer   string
	Browser   *browser.Browser
	FileName  string
}

func NewRapid(fileID string, source *DownloadSource) *Rapid {
	rapid := &Rapid{}
	rapid.Base = NewBase(fileID, source)
	rapid.Browser = surf.NewBrowser()
	return rapid
}

func (r *Rapid) parse() error {
	if _, err := url.Parse(r.DlSource.Value); err != nil {
		return err
	}

	rapidLink, err := r.getRapidURLWithQuality(r.DlSource.Value)
	if err != nil {
		return err
	}

	r.Referer = rapidLink
	direct, err := r.getRapidDirectLink(rapidLink)
	if err != nil {
		return err
	}

	log4go.Info("Direct rapid url %s", direct)

	r.DirectURL = direct
	return nil
}

func (r *Rapid) getRapidDirectLink(rapidLink string) (string, error) {
	req, _ := http.NewRequest("GET", rapidLink, nil)
	req.Header.Add("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_13_3) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/64.0.3282.167 Safari/537.36")
	req.Header.Add("Referer", r.Referer)
	req.Header.Add("Accept", "*/*")
	req.Header.Add("Pragma", "no-cache")
	req.Header.Add("Cache-Control", "no-cache")
	req.Header.Add("Accept-Language", "en-US,en;q=0.9,vi;q=0.8")
	req.Header.Add("Accept-Encoding", "identity;q=1, *;q=0")
	res, _ := r.Client.Do(req)
	dom, _ := goquery.NewDocumentFromReader(res.Body)
	r.FileName = dom.Find("title").Text()

	if directLink, ok := dom.Find("#videojs > source").Attr("src"); ok {
		return directLink, nil
	}

	return "", errors.New("Không lấy được link rapid, ko biết nguyên nhân")
}

func (r *Rapid) getRapidURLWithQuality(rapidLink string) (string, error) {
	qualities, err := r.getQualities(rapidLink)
	if err != nil {
		return "", err
	}

	quality := r.getLargestQualities(qualities)
	log4go.Debug(quality)

	if quality == "none" {
		return "", errors.New("None quality")
	}

	return rapidLink + "&q=" + quality, nil
}

func (r *Rapid) getLargestQualities(qualities map[string]bool) string {
	if _, ok := qualities["720p"]; ok {
		return "720p"
	}

	if _, ok := qualities["480p"]; ok {
		return "480p"
	}

	if _, ok := qualities["360p"]; ok {
		return "360p"
	}

	if _, ok := qualities["240p"]; ok {
		return "240p"
	}
	return "none"
}

func (r *Rapid) getQualities(rapidURL string) (map[string]bool, error) {

	req, _ := http.NewRequest("GET", rapidURL, nil)
	req.Header.Add("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_13_3) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/64.0.3282.167 Safari/537.36")
	req.Header.Add("Referer", r.Referer)
	req.Header.Add("Accept", "*/*")
	req.Header.Add("Pragma", "no-cache")
	req.Header.Add("Cache-Control", "no-cache")
	req.Header.Add("Accept-Language", "en-US,en;q=0.9,vi;q=0.8")
	req.Header.Add("Accept-Encoding", "identity;q=1, *;q=0")

	res, err := r.Client.Do(req)
	if err != nil {
		return nil, err
	}

	if res.StatusCode != 200 {
		return nil, err
	}

	var quanties = map[string]bool{}

	doc, _ := goquery.NewDocumentFromReader(res.Body)
	doc.Find("a").Each(func(i int, s *goquery.Selection) {
		href, _ := s.Attr("href")
		if idx := strings.Index(href, "="); strings.HasPrefix(href, rapidURL) && idx > 0 {
			quanties[href[idx+1:]] = true
		}
	})

	if len(quanties) == 0 {
		return nil, errors.New("Không tìm được độ phân giải nào ")
	}

	return quanties, nil
}

func (r *Rapid) Do() (result *DownloadResult, err error) {
	if err := r.parse(); err != nil {
		return nil, err
	}

	log4go.Info("Start download rapid %s name %s", r.FileID, r.FileName)

	dir := makeDownloadDir()
	defer func() {
		if result == nil {
			if err := os.RemoveAll(dir); err != nil {
				log4go.Error(err)
			}
		}
	}()

	filePath := dir + string(os.PathSeparator) + r.FileName
	file, err := os.Create(filePath)
	if err != nil {
		log4go.Error(err)
		return nil, err
	}
	defer file.Close()

	req, _ := http.NewRequest("GET", r.DirectURL, nil)
	req.Header.Add("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_13_3) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/64.0.3282.167 Safari/537.36")
	req.Header.Add("Referer", r.Referer)
	req.Header.Add("Connection", "keep-alive")
	req.Header.Add("Pragma", "no-cache")
	req.Header.Add("Cache-Control", "no-cache")
	req.Header.Add("Accept-Language", "en-US,en;q=0.9,vi;q=0.8")
	req.Header.Add("Accept-Encoding", "identity;q=1, *;q=0")

	resp, err := r.Client.Do(req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != 200 {
		return nil, errors.New(resp.Status)
	}

	defer resp.Body.Close()

	fileSize, err := r.copy(file, resp.Body)
	if err != nil {
		log4go.Error(err)
		return nil, err
	}

	if fileSize < FiveMB {
		log4go.Error("File size nhỏ hơn 5MB")
		return nil, errors.New("File size nhỏ hơn 5MB")
	}

	result = &DownloadResult{
		FileID: r.FileID,
		Path:   filePath,
		Dir:    dir,
	}
	return result, nil
}
