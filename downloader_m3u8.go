package downloader

import (
	"path"

	"github.com/canhlinh/hlsdl"
	"github.com/canhlinh/log4go"
	"github.com/canhlinh/pluto"
)

type M3u8Downloader struct {
	*Base
	pluto *pluto.Pluto
}

func NewM3u8Downloader(fileID string, source *DownloadSource) *M3u8Downloader {
	d := &M3u8Downloader{}
	d.Base = NewBase(fileID, source)

	return d
}

func (d *M3u8Downloader) Do() (*DownloadResult, error) {
	dir := path.Join(TempFolder, d.FileID)

	filePath, err := hlsdl.New(d.DlSource.Value, dir, 10, false).Download()
	if err != nil {
		log4go.Error(err)
		return nil, err
	}

	dlFile := &DownloadResult{
		FileID: d.FileID,
		Path:   filePath,
		Dir:    dir,
	}
	return dlFile, nil
}
