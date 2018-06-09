package blocks

import (
	"math"
	"math/rand"
	"runtime"
	"sort"
	"sync"

	"github.com/hjfreyer/blocks/simplenn"
	"github.com/hjfreyer/blocks/snake"
)

const (
	senseInputs  = 9
	senseOutputs = 3
	popSize      = 500
	popSave      = 20
)

func sigmoid(x float64) float64 {
	return 1.0 / (1.0 + math.Exp(-x))
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

func Evolve(size, numNeurons, numMemories int, metaMutateRate, metaMutateWidth float64, out chan<- Stat) {
	e := &Evolver{
		j: &Judger{size: size, numNeurons: numNeurons, numMemories: numMemories},
	}
	e.j.start()

	pop := &Population{
		Members: make([]*Organism, popSize),
	}
	for idx := range pop.Members {
		model := make([]float64, simplenn.ModelSize(numNeurons, numMemories))
		for idx2 := range model {
			model[idx2] = rand.NormFloat64()
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

type Judger struct {
	wg          sync.WaitGroup
	input       chan *Organism
	size        int
	numNeurons  int
	numMemories int
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
			snn := simplenn.New(j.numNeurons, j.numMemories, org.Model)
			org.TotalScore += snake.PlayFullGame(j.size, snn, nil)
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
	j               *Judger
	metaMutateRate  float64
	metaMutateWidth float64
}

func (e *Evolver) DoGeneration(p *Population) {
	e.reproduce(p.Members[popSave:], p.Members[:popSave])
	e.j.Judge(p.Members)
	sort.Sort(p)
}

func (e *Evolver) reproduce(dst, src []*Organism) {
	for didx := range dst {
		rents := rand.Perm(len(src))[:2]
		newModel := e.crossOver(src[rents[0]].Model, src[rents[1]].Model)
		dst[didx] = &Organism{Model: newModel}
	}
}

func (e *Evolver) crossOver(a, b []float64) []float64 {
	newModel := make([]float64, len(a))
	cutOff := rand.Intn(len(a) + 1)
	copy(newModel[:cutOff], a[:cutOff])
	copy(newModel[cutOff:], b[cutOff:])

	for idx := range newModel[:2] {
		if rand.Float64() < e.metaMutateRate {
			newModel[idx] *= 1 + e.metaMutateWidth*rand.NormFloat64()
		}
	}

	mutateRate := sigmoid(newModel[0])
	mutateWidth := newModel[1]

	for idx := range newModel[2:] {
		if rand.Float64() < sigmoid(mutateRate) {
			newModel[idx+2] *= 1 + rand.NormFloat64()*mutateWidth
		}
	}

	return newModel
}
