package downloader

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"

	"github.com/canhlinh/log4go"
	"github.com/canhlinh/pluto"
)

const (
	MaxParts = 20
)

type DirectDownloader struct {
	*Base
	pluto *pluto.Pluto
}

func NewDirectDownloader(fileID string, source *DownloadSource) *DirectDownloader {
	d := &DirectDownloader{}
	d.Base = NewBase(fileID, source)

	return d
}

func (d *DirectDownloader) init() error {
	cookies := []*http.Cookie{}
	if len(d.DlSource.Cookies) > 0 {
		for _, cookie := range d.DlSource.Cookies {
			cookies = append(cookies, &http.Cookie{
				Name:  cookie.Name,
				Value: cookie.Value,
			})
		}
	}

	headers := []string{}
	for key, value := range d.DlSource.Header {
		headers = append(headers, fmt.Sprintf("%s:%s", key, value))
	}
	headers = append(headers, fmt.Sprintf("%s:%s", "Cookie", CookiesToHeader(cookies)))

	fileURL, err := url.Parse(d.DlSource.Value)
	if err != nil {
		return err
	}

	log4go.Info("Max parts %v", d.DlSource.MaxParts)
	d.pluto, err = pluto.New(fileURL, headers, d.DlSource.MaxParts, false)
	if err != nil {
		return err
	}
	return nil
}

func (d *DirectDownloader) Do() (*DownloadResult, error) {

	if err := d.init(); err != nil {
		return nil, err
	}

	log4go.Info("Start download direct url %s", d.DlSource.Value)
	f, err := ioutil.TempFile(TempFolder, d.Base.FileID)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	if r, err := d.pluto.Download(context.Background(), f); err != nil {
		return nil, err
	} else {
		log4go.Info("Pluto download result file: %s size: %v", r.FileName, r.Size)
	}

	dlFile := &DownloadResult{
		FileID: d.FileID,
		Path:   Rename(f.Name()),
	}
	return dlFile, nil
}
