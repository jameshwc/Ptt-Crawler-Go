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
	baseUrl   string
	storePath string
	board     mapset.Set
}

func NewPTT(storePathFolder string) *PTT {
	p := new(PTT)
	p.baseUrl = "https://www.ptt.cc/bbs/"
	p.storePath = storePathFolder
	p.board = mapset.NewSet()
	return p
}

func (p *PTT) SetBoard(board string) error {
	if isValidBoard(p.baseUrl, board) {
		p.board.Add(board)
	} else {
		return fmt.Errorf("board name %s not valid", board)
	}
	return nil
}

func (p *PTT) SetBoardWithSlice(board []string) error {
	errMsg := ""
	for id := range board {
		if isValidBoard(p.baseUrl, board[id]) {
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

func isValidBoard(baseurl, board string) bool {
	if resp, err := http.Get(baseurl + board + "/index.html"); err != nil {
		fmt.Println(err)
		return false
	} else if resp.StatusCode != 200 {
		return false
	}
	return true
}
