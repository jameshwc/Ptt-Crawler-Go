package ptt

import (
	"net/http"
	"strconv"
	"testing"
)

const (
	testDir          = "test/"
	testPages        = 1
	testNumOfRoutine = 100
)

func Benchmark_GetArticlesURLThread(b *testing.B) {
	p := NewPTT(testDir, testPages, testNumOfRoutine)
	for i := 0; i < b.N; i++ {
		p.getArticlesURLThread("Seniorhigh", 500, -1)
	}
}

func Benchmark_GetArticlesURL(b *testing.B) {
	p := NewPTT(testDir, testPages, testNumOfRoutine)
	for i := 0; i < b.N; i++ {
		p.GetArticlesURL("Seniorhigh", 500, -1)
	}
}

func Test_getLastArticlePage(t *testing.T) {
	boards := []string{"Gossiping", "Womentalk", "b07902xxx"}
	for i := range boards {
		url := "https://www.ptt.cc/bbs/" + boards[i] + "/index"
		n := getLastArticlePage(url)
		sendRequest := func(url string, shouldFailed bool) {
			req, err := http.NewRequest("GET", url, nil)
			if err != nil {
				t.Fatal("NewRequest")
			}
			req.AddCookie(over18cookie)
			resp, err := defaultClient.Do(req)
			if err != nil {
				t.Fail()
			} else if resp.StatusCode == 200 && shouldFailed {
				t.Fail()
			} else if resp.StatusCode != 200 && !shouldFailed {
				t.Fail()
			}
		}
		sendRequest(url+strconv.Itoa(n)+".html", false)
		sendRequest(url+strconv.Itoa(n+2)+".html", true) // getLastArticlePage will not return exact last page; instead, it return the num of last two page to avoid overlapped articles

	}
}
