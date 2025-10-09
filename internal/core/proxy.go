package core

import (
	"log"
	"net/http"
	"net/url"
	"time"
)

type Proxy struct {
	URL      string
	Alive    bool
	LastTest time.Time
	CheckURL string
	Timeout  time.Duration
}

func (p *Proxy) Test(timeout time.Duration) bool {
	proxyURL, err := url.Parse(p.URL)
	if err != nil {
		p.Alive = false
		p.LastTest = time.Now()
		return false
	}

	client := &http.Client{
		Timeout: timeout,
		Transport: &http.Transport{
			Proxy: http.ProxyURL(proxyURL),
		},
	}

	resp, err := client.Get(p.CheckURL)
	if err != nil {
		log.Printf("Health check failed for %s: %v", p.URL, err)
		p.Alive = false
	} else {
		defer resp.Body.Close()
		p.Alive = resp.StatusCode >= 200 && resp.StatusCode < 300
	}

	p.LastTest = time.Now()
	return p.Alive
}
