package downloader

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/canhlinh/log4go"
	"github.com/canhlinh/pluto"
	"gopkg.in/cheggaaa/pb.v1"
)

const (
	MaxParts = 20
)

var (
	DefaultSlowSpeed uint64 = 100000
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

	defer func() {
		bar.Finish()
		f.Close()
		close(quit)
	}()

	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		lowerCounter := 0
		for {
			select {
			case s := <-d.pluto.StatsChan:
				bar.SetTotal64(int64(s.Size))
				bar.Set64(int64(s.Downloaded))

				if s.Speed < DefaultSlowSpeed {
					lowerCounter++
					if lowerCounter >= 30 {
						log4go.Warn("Slow download speed detected. Cancelling download thread for file_id %s speed %d KB/s", d.FileID, s.Speed/1000)
						cancel()
						return
					}
				} else {
					lowerCounter = 0
				}

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
