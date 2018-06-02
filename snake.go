package blocks

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

type SnakeGame struct {
	Size  int
	State GameState

	Fruit     image.Point
	Body      []image.Point
	Direction image.Point
}

func NewGame(size int) *SnakeGame {
	center := image.Pt(size/2, size/2)
	dir := image.Pt(0, -1)
	return &SnakeGame{
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

func (g *SnakeGame) Move(move Move) {
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

func (g *SnakeGame) Clone() *SnakeGame {
	var g2 SnakeGame
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

func Stimulus(g *SnakeGame) []float64 {
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

func eye(g *SnakeGame, dir image.Point) []int {
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

func WriteGame(g []*SnakeGame, o io.Writer) error {
	var out gameFmt
	out.Width = g[0].Size
	out.Height = g[0].Size
	out.Steps = make([]stepFmt, len(g))

	for idx, game := range g {
		out.Steps[idx].Comment = fmt.Sprintf("%v", Stimulus(game))
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

//def vect_div(p, q):/
//	codirectional = (np.sign(p) == np.sign(q)).all(axis=1)
//	return np.where(codirectional, np.abs(np.sum(p, axis=1)), 0)

/*
def ExportGame(stream, games):
	data = dict(width=games[0].size, height=games[0].size, steps=[])
	for g in games:
		step = dict(pts=[], comment='')

		fx, fy = g.fruit
		pts = [dict(x=fx, y=fy, type='fruit')]
		pts += [dict(x=x, y=y, type='snake')
				for (x, y) in g.body]
		data['steps'].append(dict(pts=pts, comment=str(get_stimuli(g))))

 	json.dump(data, stream, indent=2)


def get_stimuli(g):
	l_eye = get_eye(g, DirPlusMove(g.direction, 'L'))
	c_eye = get_eye(g, DirPlusMove(g.direction, 'C'))
	r_eye = get_eye(g, DirPlusMove(g.direction, 'R'))
	return l_eye + c_eye + r_eye

def vect_div(p, q):
	codirectional = (np.sign(p) == np.sign(q)).all(axis=1)
	return np.where(codirectional, np.abs(np.sum(p, axis=1)), 0)

def get_eye(g, dir):
	head = g.body[0]

	# Walls. Do it dumbly.
	if dir[0] == 1:  # Right.
		wall_idx = g.size - head[0]
	elif dir[0] == -1:  # Left.
		wall_idx = head[0] + 1
	elif dir[1] == 1:  # Down.
		wall_idx = g.size - head[1]
	elif dir[1] == -1:  # Up.
		wall_idx = head[1] + 1
	else:
		raise Exception("Ahhhh")

	fruit_div = vect_div([g.fruit-head], dir)
	fruit_idx = int(fruit_div[0]) if fruit_div[0] > 0 else -1

	body_div = vect_div(g.body[1:] - head, dir)
	body_div = body_div[0 < body_div]
	body_idx = int(body_div.min()) if len(body_div) else -1
	return [wall_idx, fruit_idx, body_idx]


if __name__ == '__main__':
	g = InitGame(11)
	games = [g]

	tries = 0
	while len(games[-1].body) < 5:
		tries += 1
		g = InitGame(11)
		games = [g]
		while g.state != 'DEAD':
			if len(games) > 100:
				break
			g = AdvanceGame(g, random.choice(MOVES))
			games.append(g)
	print 'tries = ', tries
	with open('ui/data.json', 'wb') as f:
		ExportGame(f, games)

*/
