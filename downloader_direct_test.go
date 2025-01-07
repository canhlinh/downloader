package downloader

import (
	"context"
	"net/url"
	"os"
	"testing"

	"github.com/canhlinh/pluto"
)

func TestDownloadDirect(t *testing.T) {
	t.Log("TestDownloadDirect")

	downloader := NewDirectDownloader("fileID", &DownloadSource{
		Value:    "http://localhost:8080/SampleVideo_1280x720_20mb.mp4",
		MaxParts: 1,
	})

	r, err := downloader.Do()
	if err != nil {
		t.Errorf("Do failed: %v", err)
		return
	}
	defer os.RemoveAll(r.Path)
}

func TestDownloadluto(t *testing.T) {
	t.Log("TestDownloadDirect")
	fileURL, err := url.Parse("https://filesamples.com/samples/video/mp4/sample_1280x720_surfing_with_audio.mp4")
	if err != nil {
		t.Errorf("url.Parse failed: %v", err)
	}

	p, err := pluto.New(fileURL, nil, 1, false, nil)
	if err != nil {
		t.Errorf("pluto.New failed: %v", err)
	}

	f, err := os.CreateTemp(os.TempDir(), "download")
	if err != nil {
		t.Errorf("os.CreateTemp failed: %v", err)
	}
	defer f.Close()

	_, err = p.Download(context.Background(), f)
	if err != nil {
		t.Errorf("Download failed: %v", err)
	}
}
