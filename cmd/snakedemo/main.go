package main

import (
	"log"
	"os"
	"os/signal"
	"time"

	"github.com/hjfreyer/blocks"
)

func main() {
	s := make(chan blocks.Stat)

	var latest blocks.Stat
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		for _ = range c {
			playAGame(latest.Best)
		}

	}()

	go blocks.Evolve(s)

	i := 0
	startTime := time.Now()
	for ss := range s {
		i++
		dur := (time.Now().Sub(startTime)) / time.Duration(i)
		log.Print(i, ss.Max, ss.Avg, dur)
		latest = ss

	}

}

func playAGame(model []float64) {
	s := blocks.NewGame(11)
	hist := []*blocks.SnakeGame{s.Clone()}
	for i := 0; i < 100000 && s.State == blocks.Live; i++ {
		stim := blocks.Stimulus(s)
		move := blocks.ApplyModel(model, stim)

		s.Move(move)
		hist = append(hist, s.Clone())

	}
	f, err := os.Create("ui/data.json")
	defer f.Close()
	if err != nil {
		log.Fatal(err)
	}
	blocks.WriteGame(hist, f)
}
