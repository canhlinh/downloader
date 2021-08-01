package downloader

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/canhlinh/log4go"
	"github.com/google/uuid"
)

const (
	FiveMB = 5 * 1024 * 1024
)

type Drive struct {
	*Base
	RespBody io.ReadCloser
	DriveID  string
}

func NewDrive(fileID string, source *DownloadSource) *Drive {
	drive := &Drive{}
	drive.Base = NewBase(fileID, source)
	drive.DriveID = source.Value

	return drive
}

func (d *Drive) getDownloadURL(url string) (io.ReadCloser, error) {
	req, _ := http.NewRequest(http.MethodGet, url, nil)
	res, err := d.Client.Do(req)
	if err != nil {
		return nil, err
	}

	if res.StatusCode != 200 {
		return nil, errors.New(res.Status)
	}

	if _, err := GetFileName(res.Header); err == nil {
		return res.Body, nil
	}

	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	downloadURI, exist := doc.Find("#uc-download-link").Attr("href")
	if !exist {
		if doc.Find(".uc-error-subcaption").Text() != "" {
			return nil, fmt.Errorf("%s", doc.Find(".uc-error-subcaption").Text())
		}
		return nil, errors.New("Không thể phân tích link download từ website")
	}

	resp, err := d.Client.Get(fmt.Sprintf("https://drive.google.com%s", downloadURI))
	if err != nil {
		return nil, err
	}

	return resp.Body, nil
}

func (d *Drive) parse() error {
	if err := d.GetDriveCookie(); err != nil {
		return err
	}

	initURL := fmt.Sprintf("https://drive.google.com/uc?id=%s&export=download", d.DriveID)
	body, err := d.getDownloadURL(initURL)
	if err != nil {
		return err
	}

	d.RespBody = body
	return nil
}

func (d *Drive) Do() (result *DownloadResult, err error) {
	log4go.Info("Start download drive_id %s", d.DriveID)
	if err := d.parse(); err != nil {
		log4go.Error(err)
		return nil, err
	}

	dir := makeDownloadDir()
	filePath := path.Join(dir, uuid.New().String())
	defer func() {
		if result == nil {
			if err := os.RemoveAll(dir); err != nil {
				log4go.Error(err)
			}
		}
	}()

	file, err := os.Create(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	defer d.RespBody.Close()
	if fileSize, err := d.copy(file, d.RespBody); err != nil {
		return nil, err
	} else if fileSize < FiveMB {
		return nil, errors.New("File size nhỏ hơn 5MB")
	}

	result = &DownloadResult{
		FileID: d.FileID,
		Path:   filePath,
		Dir:    dir,
	}
	return result, nil
}

func (d *Drive) Progress(current, total int) {
}

func (d *Drive) GetDriveCookie() error {
	webLink := fmt.Sprintf("https://drive.google.com/file/d/%s/view", d.DriveID)
	req, _ := http.NewRequest(http.MethodGet, webLink, nil)
	d.Client.Jar.SetCookies(req.URL, []*http.Cookie{})
	cookies := []*http.Cookie{}
	for _, cookie := range d.DlSource.Cookies {
		cookie := &http.Cookie{
			Name:    cookie.Name,
			Value:   cookie.Value,
			Domain:  ".google.com",
			Expires: time.Now().AddDate(0, 1, 0),
			Path:    "/",
		}
		switch cookie.Name {
		case "DRIVE_STREAM", "S":
			cookie.Domain = ".drive.google.com"
		}

		cookies = append(cookies, cookie)
	}
	d.Client.Jar.SetCookies(req.URL, cookies)
	req.Header.Add("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_12_6) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/60.0.3112.113 Safari/537.36")
	res, err := d.Client.Do(req)
	if err != nil {
		log4go.Error(err)
		return err
	}

	if res.StatusCode != 200 {
		return errors.New(res.Status)
	}
	res.Body.Close()

	return nil
}
