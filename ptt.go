package ptt

import (
	"errors"
	"fmt"
	"net/http"

	mapset "github.com/deckarep/golang-set"
)

// crawlBoard
//

type PTT struct {
	baseURL   string
	bbsURL    string
	storePath string
	board     mapset.Set
}

var (
	over18cookie  *http.Cookie = &http.Cookie{Name: "over18", Value: "1"}
	defaultClient *http.Client = &http.Client{}
)

func NewPTT(storePathFolder string) *PTT {
	p := new(PTT)
	p.baseURL = "https://www.ptt.cc/"
	p.bbsURL = "https://www.ptt.cc/bbs/"
	p.storePath = storePathFolder
	p.board = mapset.NewSet()
	return p
}

func (p *PTT) SetBoard(board string) error {
	if isValidBoard(p.bbsURL, board) {
		p.board.Add(board)
	} else {
		return fmt.Errorf("board name %s not valid", board)
	}
	return nil
}

func (p *PTT) SetBoardWithSlice(board []string) error {
	errMsg := ""
	for id := range board {
		if isValidBoard(p.bbsURL, board[id]) {
			p.board.Add(board[id])
		} else {
			errMsg += board[id] + " "
		}
	}
	if len(errMsg) > 0 {
		return errors.New("board name " + errMsg + "not valid")
	}
	return nil
}

func isValidBoard(bbsUrl, board string) bool {
	if resp, err := http.Get(bbsUrl + board + "/index.html"); err != nil {
		fmt.Println(err)
		return false
	} else if resp.StatusCode != 200 {
		return false
	}
	return true
}
