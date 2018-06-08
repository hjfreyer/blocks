package snake

import (
	"encoding/json"
	"fmt"
	"image"
	"io"
	"math/rand"
)

type GameState int

const (
	Live GameState = iota
	Dead
)

type Move int

const (
	Left Move = iota
	Center
	Right
)

type Mover interface {
	Move(input []float64) Move
}

type Game struct {
	Size  int
	State GameState

	Fruit     image.Point
	Body      []image.Point
	Direction image.Point
}

func NewGame(size int) *Game {
	center := image.Pt(size/2, size/2)
	dir := image.Pt(0, -1)
	return &Game{
		Size:      size,
		State:     Live,
		Fruit:     randPt(size),
		Body:      []image.Point{center, center.Sub(dir)},
		Direction: dir,
	}
}

func randPt(size int) image.Point {
	return image.Pt(rand.Intn(size), rand.Intn(size))
}

func (g *Game) Move(move Move) {
	newDir := dirPlusMove(g.Direction, move)
	newHead := g.Body[0].Add(newDir)

	if !newHead.In(image.Rect(0, 0, g.Size, g.Size)) {
		g.State = Dead
		return
	}

	newBody := make([]image.Point, len(g.Body)+1)
	newBody[0] = newHead
	copy(newBody[1:], g.Body)

	if newHead == g.Fruit {
		g.Fruit = randPt(g.Size)
		for contains(newBody, g.Fruit) {
			g.Fruit = randPt(g.Size)
		}
	} else {
		newBody = newBody[:len(newBody)-1]
	}

	if contains(newBody[1:], newHead) {
		g.State = Dead
		return
	}
	g.Body = newBody
	g.Direction = newDir
}

func (g *Game) Clone() *Game {
	var g2 Game
	g2 = *g
	g2.Body = make([]image.Point, len(g.Body))
	copy(g2.Body, g.Body)
	return &g2
}

func dirPlusMove(dir image.Point, move Move) image.Point {
	switch move {
	case Center:
		return dir
	case Left:
		return image.Pt(dir.Y, -dir.X)
	case Right:
		return image.Pt(-dir.Y, dir.X)
	}
	panic("bad move")
}

func contains(pts []image.Point, pt image.Point) bool {
	for _, p := range pts {
		if pt == p {
			return true
		}
	}
	return false
}

func Stimulus(g *Game) []float64 {
	var ints []int
	ints = append(ints, eye(g, dirPlusMove(g.Direction, Left))...)
	ints = append(ints, eye(g, dirPlusMove(g.Direction, Center))...)
	ints = append(ints, eye(g, dirPlusMove(g.Direction, Right))...)

	floats := make([]float64, 9)
	for idx, i := range ints {
		floats[idx] = float64(i)
	}
	return floats
}

func eye(g *Game, dir image.Point) []int {
	head := g.Body[0]

	wallIdx := func() int {
		switch {
		case dir.X == 1: // Right.
			return g.Size - head.X
		case dir.X == -1: // Left
			return head.X + 1
		case dir.Y == 1: // Down
			return g.Size - head.Y
		case dir.Y == -1: // Up.
			return head.Y + 1
		default:
			panic("bad dir")
		}
	}()

	fruitIdx, fruitIsDiv := vectDiv(g.Fruit.Sub(head), dir)
	if !fruitIsDiv {
		fruitIdx = 100
	}

	bodyIdx := 100
	for _, b := range g.Body[1:] {
		candidateBody, bodyIsDiv := vectDiv(b.Sub(head), dir)
		if bodyIsDiv && candidateBody < bodyIdx {
			bodyIdx = candidateBody
		}
	}

	return []int{wallIdx, fruitIdx, bodyIdx}
}

func vectDiv(p, q image.Point) (int, bool) {
	if sgn(p.X) != sgn(q.X) || sgn(p.Y) != sgn(q.Y) {
		return 0, false
	}

	quot := p.X*sgn(q.X) + p.Y*sgn(q.Y)
	return quot, 0 < quot
}

func sgn(i int) int {
	switch {
	case i < 0:
		return -1
	case i == 0:
		return 0
	default:
		return 1
	}
}

type gameFmt struct {
	Width  int       `json:"width"`
	Height int       `json:"height"`
	Steps  []stepFmt `json:"steps"`
}

type stepFmt struct {
	Comment string  `json:"comment"`
	Points  []ptFmt `json:"pts"`
}

type ptFmt struct {
	X    int    `json:"x"`
	Y    int    `json:"y"`
	Type string `json:"type"`
}

func PlayFullGame(size int, mover Mover, history *[]*Game) int {
	s := NewGame(size)
	if history != nil {
		*history = []*Game{s.Clone()}
	}

	const gameLimit = 10000
	lastIncrease := 0
	lastLen := len(s.Body)
	for i := 0; i < gameLimit && s.State == Live; i++ {
		s.Move(mover.Move(Stimulus(s)))
		if history != nil {
			*history = append(*history, s.Clone())
		}

		if lastLen < len(s.Body) {
			lastLen = len(s.Body)
			lastIncrease = i
		}

		if lastIncrease+1000 < i {
			return len(s.Body)
		}
	}
	return len(s.Body)
}

func WriteGame(g []*Game, o io.Writer) error {
	var out gameFmt
	out.Width = g[0].Size
	out.Height = g[0].Size
	out.Steps = make([]stepFmt, len(g))

	for idx, game := range g {
		out.Steps[idx].Comment = fmt.Sprintf("%d: %v", idx, Stimulus(game))
		out.Steps[idx].Points = []ptFmt{
			{game.Fruit.X, game.Fruit.Y, "fruit"},
		}

		for _, seg := range game.Body {
			out.Steps[idx].Points = append(out.Steps[idx].Points, ptFmt{seg.X, seg.Y, "snake"})
		}
	}

	enc := json.NewEncoder(o)
	return enc.Encode(out)
}
