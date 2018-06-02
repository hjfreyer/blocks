package main

import (
	"fmt"
	"math/rand"
	"os"
	"runtime"

	"github.com/hjfreyer/blocks"
)

func main() {
	fmt.Println(runtime.GOMAXPROCS(0))
	return
	model := make([]float64, blocks.ModelSize)

	for idx := range model {
		model[idx] = rand.NormFloat64()
	}

	s := blocks.NewGame(11)
	hist := []*blocks.SnakeGame{s.Clone()}
	for i := 0; i < 100 && s.State == blocks.Live; i++ {

		for idx := range model {
			model[idx] = rand.NormFloat64()
		}

		stim := blocks.Stimulus(s)
		move := blocks.ApplyModel(model, stim)

		s.Move(move)
		hist = append(hist, s.Clone())

	}
	blocks.WriteGame(hist, os.Stdout)

}
