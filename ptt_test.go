package ptt

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func testOneBoard(t *testing.T) {

}
func Test_SetBoard(t *testing.T) {
	p := NewPTT("dat/")
	validBoard := []string{"gossiping", "Gossiping", "seniorhigh", "b07902xxx"}
	invalidBoard := []string{"ABCfjisiodjs", "fjiosw9dnjsc", "123", "ABCDEFG"}
	assert := assert.New(t)
	assert.Equal(len(validBoard), len(invalidBoard))
	for i := 0; i < len(validBoard); i++ {
		if err := p.SetBoard(validBoard[i]); err != nil {
			t.Fail()
		}
		if err := p.SetBoard(invalidBoard[i]); err == nil {
			t.Fail()
		} else if strings.Compare(err.Error(), "board name "+invalidBoard[i]+" not valid") != 0 {
			t.Fail()
		}
	}

	pWithValidSlice := NewPTT("dat/")
	pWithValidSlice.SetBoardWithSlice(validBoard)
	pWithInvalidSlice := NewPTT("dat/")
	pWithInvalidSlice.SetBoardWithSlice(invalidBoard)
	assert.True(p.board.Equal(pWithValidSlice.board))
	assert.False(p.board.Equal(pWithInvalidSlice.board))
}
