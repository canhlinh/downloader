package downloader

import (
	"os"
	"path"

	"github.com/canhlinh/hlsdl"
	"github.com/canhlinh/log4go"
	"github.com/google/uuid"
)

type M3u8Downloader struct {
	*Base
}

func NewM3u8Downloader(fileID string, source *DownloadSource) *M3u8Downloader {
	d := &M3u8Downloader{}
	d.Base = NewBase(fileID, source)

	return d
}

func (d *M3u8Downloader) Do() (result *DownloadResult, err error) {
	dir := path.Join(TempFolder, uuid.New().String())
	defer func() {
		if result == nil {
			if err := os.RemoveAll(dir); err != nil {
				log4go.Error(err)
			}
		}
	}()

	filePath, err := hlsdl.New(d.DlSource.Value, d.DlSource.Header, dir, 10, false, "").Download()
	if err != nil {
		log4go.Error(err)
		os.RemoveAll(dir)
		return nil, err
	}

	result = &DownloadResult{
		FileID: d.FileID,
		Path:   filePath,
		Dir:    dir,
	}
	return result, nil
}
