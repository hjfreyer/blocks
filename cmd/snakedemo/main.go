package main

import (
	"log"
	"os"
	"os/signal"
	"time"

	"github.com/hjfreyer/blocks"
	"github.com/hjfreyer/blocks/simplenn"
	"github.com/hjfreyer/blocks/snake"
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
	s := snake.NewGame(11)
	snn := simplenn.New(12, 4, model)
	hist := []*snake.Game{s.Clone()}
	for i := 0; i < 100000 && s.State == snake.Live; i++ {
		s.Move(snn.Move(snake.Stimulus(s)))
		hist = append(hist, s.Clone())

	}
	f, err := os.Create("ui/data.json")
	defer f.Close()
	if err != nil {
		log.Fatal(err)
	}
	snake.WriteGame(hist, f)
}
