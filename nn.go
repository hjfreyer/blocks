package blocks

import (
	"log"
	"math"
	"math/rand"
	"runtime"
	"sort"
	"sync"

	"github.com/hjfreyer/blocks/simplenn"
	"github.com/hjfreyer/blocks/snake"
)

const (
	senseInputs   = 9
	senseOutputs  = 3
	popSize       = 500
	numSpecies    = 10
	speciesFactor = 570
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
		model[0] = 0.1
		model[1] = 0.1
		pop.Members[idx] = &Organism{Model: model}
	}
	e.j.Judge(pop.Members)

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
		snn := simplenn.New(j.numNeurons, j.numMemories, org.Model)
		for i := 0; i < gameCount; i++ {
			org.NumGames++
			org.TotalScore += snake.PlayFullGame(j.size, snn.NewGame(), nil)
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
	p.Members = e.reproduce(p.Members)
	e.j.Judge(p.Members)
	sort.Sort(p)
}

/*
func species(model []float64) int {
	seed := make([]byte, 8)
	binary.LittleEndian.PutUint64(seed, math.Float64bits(model[2]))
	return int(crc32.ChecksumIEEE(seed) % numSpecies)
}
*/
func (e *Evolver) reproduce(src []*Organism) []*Organism {
	species := speciate(src)
	speciesByOrg := make([]int, len(src))
	for sIdx, s := range species {
		for _, oIdx := range s {
			speciesByOrg[oIdx] = sIdx
		}
	}

	log.Print(len(species))

	//	bestPerSpecies := make([]*Organism, numSpecies)

	weights := make([]float64, len(src))

	for idx, o := range src {
		weights[idx] = o.Score()

		//		if bestPerSpecies[spec] == nil || bestPerSpecies[spec].Score() < o.Score() {
		//			bestPerSpecies[spec] = o
		//		}
	}

	for idx := range weights {
		weights[idx] /= float64(len(species[speciesByOrg[idx]]))
		//		weights[idx] = math.Exp(preventOverflow(weights[idx]))
	}

	dst := make([]*Organism, 0, len(src))
	/*	for _, best := range bestPerSpecies {
		if best != nil {
			dst = append(dst, best)
		}
	}*/

	for len(dst) < len(src) {
		p1 := weightedRandomChoice(weights)
		p2 := weightedRandomChoice(weights)
		newModel := e.crossOver(src[p1].Model, src[p2].Model)
		dst = append(dst, &Organism{Model: newModel})
	}
	return dst
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

func speciate(orgs []*Organism) [][]int {
	var res [][]int
nextOrg:
	for oIdx, o := range orgs {
		for sIdx, s := range res {
			if dist(orgs[s[0]].Model, o.Model) < speciesFactor {
				res[sIdx] = append(res[sIdx], oIdx)
				continue nextOrg
			}
		}
		// No species found - make our own.
		res = append(res, []int{oIdx})
	}
	return res
}

func dist(a, b []float64) float64 {
	if len(a) != len(b) {
		panic("ahhhh")
	}
	var total float64
	for idx := range a {
		total += math.Abs(a[idx] - b[idx])
	}
	return total
}

func weightedRandomChoice(distribution []float64) int {
	var total float64
	for _, val := range distribution {
		total += val
	}
	r := rand.Float64() * total
	var cumlWeight float64
	for idx, weight := range distribution {
		cumlWeight += weight
		if r < cumlWeight {
			return idx
		}
	}
	panic("lolwut")
}

func preventOverflow(v float64) float64 {
	if v > 7.09782712893383973096e+02 {
		return 7.09782712893383973096e+02
	}
	if v < -7.45133219101941108420e+02 {
		return -7.45133219101941108420e+02
	}
	return v
}
