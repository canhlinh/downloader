package downloader

import (
	"net/http"
	"time"

	"github.com/canhlinh/pluto"
)

type Cookie struct {
	Domain         string  `json:"domain"`
	ExpirationDate float64 `json:"expirationDate"`
	HostOnly       bool    `json:"hostOnly"`
	HTTPOnly       bool    `json:"httpOnly"`
	Name           string  `json:"name"`
	Path           string  `json:"path"`
	SameSite       string  `json:"sameSite"`
	Secure         bool    `json:"secure"`
	Session        bool    `json:"session"`
	StoreID        string  `json:"storeId"`
	Value          string  `json:"value"`
}

func (c *Cookie) HttpCookie() *http.Cookie {
	return &http.Cookie{
		Name:     c.Name,
		Domain:   c.Domain,
		Value:    c.Value,
		Path:     c.Path,
		SameSite: http.SameSiteNoneMode,
		Expires:  time.Unix(int64(c.ExpirationDate), 0),
		Secure:   c.Secure,
	}
}

type Cookies []*Cookie

func (cookies Cookies) HttpCookies() []*http.Cookie {
	httpCookies := []*http.Cookie{}

	for _, c := range cookies {
		httpCookies = append(httpCookies, c.HttpCookie())
	}

	return httpCookies
}

type DownloadSource struct {
	Type     string            `json:"type"`
	Value    string            `json:"value"`
	MaxParts uint              `json:"max_parts"`
	Header   map[string]string `json:"header"`
	Cookies  Cookies           `json:"cookies"`
	Proxy    *pluto.Proxy      `json:"proxy"`
}

type DownloadResult struct {
	FileID string `json:"file_id"`
	Path   string `json:"path"`
	Dir    string `json:"dir"`
}
