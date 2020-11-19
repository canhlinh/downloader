package downloader

import (
	"fmt"
	"io"
	"mime"
	"net/http"
	"net/http/cookiejar"
	"net/textproto"
	"os"
	"path"
	"strings"
	"time"

	"github.com/canhlinh/log4go"
	"golang.org/x/net/publicsuffix"
)

const (
	SourceDrive    = "drive"
	SourceRapid    = "rapid"
	SourceDirect   = "direct"
	SourceRedirect = "redirect"
	DownloadFolder = "/tmp" + string(os.PathSeparator) + "download"
)

const (
	DL_SOURCE_DIRECT   = "direct_url"
	DL_SOURCE_DRIVE    = "drive"
	DL_SOURCE_720DRIVE = "drive_720p"
	DL_SOURCE_GIG      = "uploadgig"
)

var TempFolder = os.TempDir()

type Downloader interface {
	parse() error
	Do() (*DownloadResult, error)
	Delete(path string)
	copy(dst io.Writer, src io.Reader) (int64, error)
}

func NewDownloader(fileID string, source *DownloadSource) Downloader {
	if source.MaxParts == 0 {
		source.MaxParts = 1
	}

	switch source.Type {
	case SourceRapid:
		return NewRapid(fileID, source)
	case SourceDrive:
		return NewDrive(fileID, source)
	case SourceDirect:
		return NewDirectDownloader(fileID, source)
	default:
		return NewBase(fileID, source)
	}
}

type Base struct {
	FileID     string `json:"file_id"`
	FileLength int64
	DlSource   *DownloadSource `json:"dl_source"`
	Client     *http.Client
}

func GetFileName(header http.Header) (string, error) {

	if cd := header.Get("Content-Disposition"); cd != "" {
		if _, params, err := mime.ParseMediaType(cd); err == nil {
			filename := params["filename"]
			if filename == "" || strings.HasSuffix(filename, "/") || strings.Contains(filename, "\x00") {
				return "", fmt.Errorf("Invalid file name %s", filename)
			}
			return filename, nil
		}
	}
	return "", fmt.Errorf("Invalid file name")
}

func randomByte() []byte {
	return []byte(time.Now().Format(time.RFC3339))
}

func NewBase(fileID string, source *DownloadSource) *Base {
	c := &http.Client{}
	jar, _ := cookiejar.New(&cookiejar.Options{PublicSuffixList: publicsuffix.List})
	c.Jar = jar

	return &Base{
		Client:   c,
		FileID:   fileID,
		DlSource: source,
	}
}

func (b *Base) Do() (*DownloadResult, error) {
	return nil, fmt.Errorf("Can not exec download this type %s", b.DlSource.Type)
}

func (b *Base) Delete(path string) {
	os.RemoveAll(path)
}

func (b *Base) copy(dst io.Writer, src io.Reader) (written int64, err error) {

	var buf = make([]byte, 32*1024)

	for {
		nr, er := src.Read(buf)
		if nr > 0 {
			nw, ew := dst.Write(buf[0:nr])
			if nw > 0 {
				written += int64(nw)
			}
			if ew != nil {
				err = ew
				break
			}
			if nr != nw {
				err = io.ErrShortWrite
				break
			}
		}
		if er != nil {
			if er != io.EOF {
				err = er
			}
			break
		}
	}

	return written, err
}

func (b *Base) parse() error {
	return nil
}

func Rename(filePath string) string {
	if path.Ext(filePath) != "" {
		return filePath
	}

	newpath := filePath + ".mp4"

	if err := os.Rename(filePath, newpath); err != nil {
		log4go.Error(err)
		return filePath
	}

	log4go.Info("Rename file from %s to %s", filePath, newpath)

	return newpath
}

func CookiesToHeader(cookies []*http.Cookie) string {
	mimeheader := &textproto.MIMEHeader{}
	for _, cookie := range cookies {
		mimeheader.Add("Cookie", cookie.String())
	}
	return mimeheader.Get("Cookie")
}
