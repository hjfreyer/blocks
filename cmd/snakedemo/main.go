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

const (
	metaMutateRate  = 0.1
	metaMutateWidth = 0.1
)

func main() {
	s := make(chan blocks.Stat)

	var latest blocks.Stat
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		for _ = range c {
			snn := simplenn.New(16, 8, latest.Best)
			var hist []*snake.Game
			snake.PlayFullGame(11, snn, &hist)

			f, err := os.Create("ui/data.json")
			defer f.Close()
			if err != nil {
				log.Fatal(err)
			}
			snake.WriteGame(hist, f)
		}

	}()

	go blocks.Evolve(11, 16, 8, metaMutateRate, metaMutateWidth, s)

	i := 0
	startTime := time.Now()
	for ss := range s {
		i++
		dur := (time.Now().Sub(startTime)) / time.Duration(i)
		log.Print(i, ss.Max, ss.Avg, dur)
		latest = ss

	}
}
