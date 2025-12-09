package game

const (
	Rows = 6
	Cols = 7
)

type Board [Rows][Cols]int

func (b *Board) CheckWin(p int) bool {
	for r := 0; r < Rows; r++ {
		for c := 0; c < Cols-3; c++ {
			if b[r][c] == p && b[r][c+1] == p && b[r][c+2] == p && b[r][c+3] == p {
				return true
			}
		}
	}
	for r := 0; r < Rows-3; r++ {
		for c := 0; c < Cols; c++ {
			if b[r][c] == p && b[r+1][c] == p && b[r+2][c] == p && b[r+3][c] == p {
				return true
			}
		}
	}
	for r := 3; r < Rows; r++ {
		for c := 0; c < Cols-3; c++ {
			if b[r][c] == p && b[r-1][c+1] == p && b[r-2][c+2] == p && b[r-3][c+3] == p {
				return true
			}
		}
	}
	for r := 3; r < Rows; r++ {
		for c := 3; c < Cols; c++ {
			if b[r][c] == p && b[r-1][c-1] == p && b[r-2][c-2] == p && b[r-3][c-3] == p {
				return true
			}
		}
	}
	return false
}

func (b *Board) BotMove() int {
	for c := 0; c < Cols; c++ {
		if b.CanDrop(c) {
			temp := *b
			temp.Drop(c, 2)
			if temp.CheckWin(2) {
				return c
			}
		}
	}
	for c := 0; c < Cols; c++ {
		if b.CanDrop(c) {
			temp := *b
			temp.Drop(c, 1)
			if temp.CheckWin(1) {
				return c
			}
		}
	}
	if b.CanDrop(3) {
		return 3
	}

	for c := 0; c < Cols; c++ {
		if b.CanDrop(c) {
			return c
		}
	}
	return 0
}

func (b *Board) CanDrop(col int) bool {
	return b[0][col] == 0
}

func (b *Board) Drop(col int, p int) (int, bool) {
	if !b.CanDrop(col) {
		return -1, false
	}
	for r := Rows - 1; r >= 0; r-- {
		if b[r][col] == 0 {
			b[r][col] = p
			return r, true
		}
	}
	return -1, false
}
