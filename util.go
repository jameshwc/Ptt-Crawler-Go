package ptt

import (
	"fmt"
	"net/http"

	"github.com/PuerkitoBio/goquery"
)

func parseURL(url string) (*goquery.Document, error) {
	// time.Sleep(100)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.AddCookie(over18cookie)
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 6.2; rv:20.0) Gecko/20121202 Firefox/26.0")
	resp, err := defaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		// resp.Body.Close()
		// resp, err = secondClient.Do(req)
		// if err != nil || resp.StatusCode != 200 {
		return nil, fmt.Errorf("ParseURL: status code %d at %s url", resp.StatusCode, url)
		// }
	}
	return goquery.NewDocumentFromReader(resp.Body)
}
