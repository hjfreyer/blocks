package simplenn

import (
	"math"

	"github.com/hjfreyer/blocks/snake"
	"gonum.org/v1/gonum/mat"
)

const (
	senseInputs     = 9
	senseOutputs    = 3
	numMutateParams = 3
)

func layerSizes(numNeurons, numMemories int) (int, int) {
	inputs := senseInputs + numMemories
	outputs := senseOutputs + numMemories

	layer1Size := (inputs + 1) * numNeurons
	layer2Size := (numNeurons + 1) * outputs
	return layer1Size, layer2Size
}

func ModelSize(numNeurons, numMemories int) int {
	l1, l2 := layerSizes(numNeurons, numMemories)
	return numMutateParams + l1 + l2
}

type SimpleNN struct {
	NumNeurons  int
	NumMemories int
	Model       []float64

	layer1 *mat.Dense
	layer2 *mat.Dense
}

type SimpleNNGame struct {
	snn      *SimpleNN
	memories []float64
}

func New(numNeurons, numMemories int, model []float64) *SimpleNN {
	l1s, _ := layerSizes(numNeurons, numMemories)
	inputs := senseInputs + numMemories
	outputs := senseOutputs + numMemories

	layer1 := mat.NewDense(numNeurons, inputs+1,
		model[numMutateParams:numMutateParams+l1s])
	layer2 := mat.NewDense(outputs, numNeurons+1, model[numMutateParams+l1s:])

	return &SimpleNN{
		NumNeurons:  numNeurons,
		NumMemories: numMemories,
		Model:       model,

		layer1: layer1,
		layer2: layer2,
	}
}

func (s *SimpleNN) NewGame() *SimpleNNGame {
	return &SimpleNNGame{
		snn:      s,
		memories: make([]float64, s.NumMemories),
	}
}
func (s *SimpleNNGame) Move(input []float64) snake.Move {
	inputs := senseInputs + len(s.memories)
	outputs := senseOutputs + len(s.memories)

	ins := make([]float64, 0, inputs+1)
	ins = append(ins, input...)
	ins = append(ins, s.memories...)
	ins = append(ins, 1)

	v0 := mat.NewVecDense(len(ins), ins)

	v1 := mat.NewVecDense(s.snn.NumNeurons, nil)
	v1.MulVec(s.snn.layer1, v0)
	rv := v1.RawVector().Data

	for idx, w := range rv {
		rv[idx] = math.Tanh(w)
	}
	rv = append(rv, 1)
	v0 = mat.NewVecDense(s.snn.NumNeurons+1, rv)
	v1 = mat.NewVecDense(outputs, nil)
	v1.MulVec(s.snn.layer2, v0)

	maxIdx := 0

	for idx := 0; idx < senseOutputs; idx++ {
		if v1.AtVec(maxIdx) < v1.AtVec(idx) {
			maxIdx = idx
		}
	}

	for i := range s.memories {
		s.memories[i] = math.Tanh(v1.AtVec(senseOutputs + i))
	}

	return snake.Move(maxIdx)
}
