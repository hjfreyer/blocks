package blocks

import (
	"math/rand"
	"runtime"
	"sort"
	"sync"

	"github.com/hjfreyer/blocks/simplenn"
	blocks "github.com/hjfreyer/blocks/snake"
)

const (
	neurons      = 16
	numMemories  = 8
	senseInputs  = 9
	senseOutputs = 3
	popSize      = 500
	popSave      = 20
)

type Mover interface {
	Move(input []float64) blocks.Move
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
		model := make([]float64, simplenn.ModelSize(neurons, numMemories))
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
	g := blocks.NewGame(11)
	snn := simplenn.New(neurons, numMemories, model)
	const gameLimit = 10000
	lastIncrease := 0
	lastLen := len(g.Body)
	for i := 0; i < gameLimit && g.State == blocks.Live; i++ {
		g.Move(snn.Move(blocks.Stimulus(g)))

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
