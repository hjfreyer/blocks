package blocks

import (
	"math"

	"gonum.org/v1/gonum/mat"
)

const (
	neurons = 12
	inputs  = 9
	outputs = 3

	layer1Size = (inputs + 1) * neurons
	layer2Size = (neurons + 1) * outputs
	ModelSize  = layer1Size + layer2Size
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
