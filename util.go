package ptt

import (
	"fmt"
	"net/http"

	"github.com/PuerkitoBio/goquery"
)

func parseURL(url string) (*goquery.Document, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.AddCookie(over18cookie)
	resp, err := defaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("ParseUrl: status code %d %s at %s url", resp.StatusCode, resp.Status, url)
	}
	return goquery.NewDocumentFromReader(resp.Body)
}
