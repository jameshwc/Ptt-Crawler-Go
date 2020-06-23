package ptt

type Option struct {
	StartPage          int
	EndPage            int
	Board              string
	URLlistFileName    string
	OutputJsonFileName string
}

func NewOption(StartPage, EndPage int, Board string) *Option {
	opt := new(Option)
	opt.Board = Board
	opt.StartPage = StartPage
	opt.EndPage = EndPage
	opt.URLlistFileName = ""
	return opt
}
