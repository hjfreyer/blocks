package blocks

import (
	"math"
	"math/rand"
	"runtime"
	"sort"
	"sync"

	"gonum.org/v1/gonum/mat"
)

const (
	neurons = 12
	inputs  = 9
	outputs = 3

	layer1Size = (inputs + 1) * neurons
	layer2Size = (neurons + 1) * outputs
	ModelSize  = layer1Size + layer2Size

	popSize = 100
	popSave = 10
)

func ApplyModel(model []float64, input []float64) Move {
	layer1 := mat.NewDense(neurons, inputs+1, model[:layer1Size])
	layer2 := mat.NewDense(outputs, neurons+1, model[layer1Size:])

	input = append(input, 1)

	v0 := mat.NewVecDense(inputs+1, input)

	v1 := mat.NewVecDense(neurons, nil)
	v1.MulVec(layer1, v0)
	rv := v1.RawVector().Data

	for idx, w := range rv {
		rv[idx] = math.Tanh(w)
	}
	rv = append(rv, 1)
	v0 = mat.NewVecDense(neurons+1, rv)
	v1 = mat.NewVecDense(outputs, nil)
	v1.MulVec(layer2, v0)

	maxIdx := 0

	for idx := 0; idx < v1.Len(); idx++ {
		if v1.AtVec(maxIdx) < v1.AtVec(idx) {
			maxIdx = idx
		}
	}

	return Move(maxIdx)
}

type Population struct {
	Members []*Organism
}

func (p *Population) Len() int      { return len(p.Members) }
func (p *Population) Swap(i, j int) { p.Members[i], p.Members[j] = p.Members[j], p.Members[i] }
func (p *Population) Less(i, j int) bool {
	return p.Members[j].Score() < p.Members[i].Score()
}

type Organism struct {
	Model      []float64
	NumGames   int
	TotalScore int
}

func (o Organism) Score() float64 {
	return float64(o.TotalScore) / float64(o.NumGames)
}

type Stat struct {
	Best []float64
	Max  float64
	Avg  float64
}

func Evolve(out chan<- Stat) {
	e := &Evolver{
		j: &Judger{},
	}
	e.j.start()

	pop := &Population{
		Members: make([]*Organism, popSize),
	}
	for idx := range pop.Members {
		model := make([]float64, ModelSize)
		for idx2 := range model {
			model[idx2] = rand.NormFloat64() * 10
		}
		pop.Members[idx] = &Organism{Model: model}
	}

	for {
		e.DoGeneration(pop)
		var s Stat
		s.Best = pop.Members[0].Model
		s.Max = pop.Members[0].Score()
		for _, o := range pop.Members {
			s.Avg += o.Score()
		}
		s.Avg /= float64(len(pop.Members))
		out <- s
	}
}

func runGame(model []float64) int {
	g := NewGame(11)
	const gameLimit = 10000
	lastIncrease := 0
	lastLen := len(g.Body)
	for i := 0; i < gameLimit && g.State == Live; i++ {
		g.Move(ApplyModel(model, Stimulus(g)))

		if lastLen < len(g.Body) {
			lastLen = len(g.Body)
			lastIncrease = i
		}

		if lastIncrease+1000 < i {
			return len(g.Body)
		}
	}
	return len(g.Body)
}

type Judger struct {
	wg    sync.WaitGroup
	input chan *Organism
}

func (e *Judger) start() {
	e.input = make(chan *Organism, 100)
	for i := 0; i < runtime.GOMAXPROCS(0); i++ {
		go e.run()
	}
}

func (j *Judger) run() {
	const gameCount = 5
	for org := range j.input {
		for i := 0; i < gameCount; i++ {
			org.NumGames++
			org.TotalScore += runGame(org.Model)
		}
		j.wg.Done()
	}
}

func (j *Judger) Judge(orgs []*Organism) {
	j.wg.Add(len(orgs))
	for _, o := range orgs {
		j.input <- o
	}
	j.wg.Wait()
}

type Evolver struct {
	j *Judger
}

func (e *Evolver) DoGeneration(p *Population) {
	reproduce(p.Members[popSave:], p.Members[:popSave])
	e.j.Judge(p.Members)
	sort.Sort(p)
}

func reproduce(dst, src []*Organism) {
	for didx := range dst {
		rents := rand.Perm(len(src))[:2]
		newModel := crossOver(src[rents[0]].Model, src[rents[1]].Model)
		dst[didx] = &Organism{Model: newModel}
	}
}

func crossOver(a, b []float64) []float64 {
	newModel := make([]float64, len(a))
	cutOff := rand.Intn(len(a) + 1)
	copy(newModel[:cutOff], a[:cutOff])
	copy(newModel[cutOff:], b[cutOff:])

	for idx := range newModel {
		if rand.Float64() < 0.05 {
			newModel[idx] += rand.NormFloat64() * 5
		}
	}

	return newModel
}
