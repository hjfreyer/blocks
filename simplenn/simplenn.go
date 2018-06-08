package simplenn

import (
	"math"

	"github.com/hjfreyer/blocks/snake"
	"gonum.org/v1/gonum/mat"
)

const (
	senseInputs  = 9
	senseOutputs = 3
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
	return l1 + l2
}

type SimpleNN struct {
	NumNeurons  int
	NumMemories int
	Model       []float64

	memories []float64
}

func New(numNeurons, numMemories int, model []float64) *SimpleNN {
	return &SimpleNN{
		NumNeurons:  numNeurons,
		NumMemories: numMemories,
		Model:       model,
		memories:    make([]float64, numMemories),
	}
}

func (s *SimpleNN) Move(input []float64) snake.Move {
	l1s, _ := layerSizes(s.NumNeurons, s.NumMemories)
	inputs := senseInputs + s.NumMemories
	outputs := senseOutputs + s.NumMemories

	layer1 := mat.NewDense(s.NumNeurons, inputs+1, s.Model[:l1s])
	layer2 := mat.NewDense(outputs, s.NumNeurons+1, s.Model[l1s:])

	ins := make([]float64, 0, inputs+1)
	ins = append(ins, input...)
	ins = append(ins, s.memories...)
	ins = append(ins, 1)

	v0 := mat.NewVecDense(len(ins), ins)

	v1 := mat.NewVecDense(s.NumNeurons, nil)
	v1.MulVec(layer1, v0)
	rv := v1.RawVector().Data

	for idx, w := range rv {
		rv[idx] = math.Tanh(w)
	}
	rv = append(rv, 1)
	v0 = mat.NewVecDense(s.NumNeurons+1, rv)
	v1 = mat.NewVecDense(outputs, nil)
	v1.MulVec(layer2, v0)

	maxIdx := 0

	for idx := 0; idx < senseOutputs; idx++ {
		if v1.AtVec(maxIdx) < v1.AtVec(idx) {
			maxIdx = idx
		}
	}

	s.memories = make([]float64, s.NumMemories)
	for i := range s.memories {
		s.memories[i] = math.Tanh(v1.AtVec(senseOutputs + i))
	}

	return snake.Move(maxIdx)
}
