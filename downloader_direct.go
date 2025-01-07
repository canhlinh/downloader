package downloader

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/canhlinh/log4go"
	"github.com/canhlinh/pluto"
	"gopkg.in/cheggaaa/pb.v1"
)

const (
	MaxParts = 20
)

var (
	DefaultSlowSpeed int64 = 100000
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
	d.pluto, err = pluto.New(fileURL, headers, d.DlSource.MaxParts, false, d.DlSource.Proxy)
	if err != nil {
		return err
	}
	return nil
}

func (d *DirectDownloader) Do() (result *DownloadResult, err error) {

	if err := d.init(); err != nil {
		return nil, err
	}
	log4go.Info("Start download direct url %s", d.DlSource.Value)

	quit := make(chan bool)
	dir := makeDownloadDir()
	defer func() {
		if result == nil {
			if err := os.RemoveAll(dir); err != nil {
				log4go.Error(err)
			}
		}
	}()

	f, err := os.CreateTemp(dir, d.Base.FileID)
	if err != nil {
		return nil, err
	}

	bar := pb.StartNew(0)
	bar.SetUnits(pb.U_BYTES)
	bar.ShowSpeed = true
	bar.ShowBar = false

	defer func() {
		bar.Finish()
		f.Close()
		close(quit)
	}()

	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		period := time.Duration(30)
		ticker := time.NewTicker(period * time.Second)
		downloaded := int64(0)
		for {
			select {
			case <-ticker.C:
				current := bar.Get()

				avgSpeed := (current - downloaded) / int64(period)
				if avgSpeed < DefaultSlowSpeed {
					log4go.Warn("No data downloaded in the last minute. Cancelling download thread for file_id %s. AvgSpeed %v(Kbs)", d.FileID, float64(avgSpeed)/1000)
					cancel()
					return
				}
				downloaded = current
			case s := <-d.pluto.StatsChan:
				bar.Set64(int64(s.Downloaded))
			case <-d.pluto.Finished:
				return
			case <-quit:
				return
			}
		}
	}()

	if r, err := d.pluto.Download(ctx, f); err != nil {
		if strings.Contains(err.Error(), "context cancel") {
			return nil, errors.New("cancelled due to slow download speed")
		}
		return nil, err
	} else {
		log4go.Info("Pluto download result file: %s size: %v", r.FileName, r.Size)
	}

	result = &DownloadResult{
		FileID: d.FileID,
		Path:   f.Name(),
		Dir:    dir,
	}
	return result, nil
}
