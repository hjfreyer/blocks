package main

import (
	"log"
	"os"
	"os/signal"
	"time"

	"github.com/hjfreyer/blocks"
	"github.com/hjfreyer/blocks/simplenn"
	"github.com/hjfreyer/blocks/snake"

	"net/http"
	_ "net/http/pprof"
)

const (
	metaMutateRate  = 0.1
	metaMutateWidth = 0.1
)

func main() {
	go func() {
		log.Println(http.ListenAndServe("localhost:6060", nil))
	}()
	s := make(chan blocks.Stat)

	var latest blocks.Stat
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		for _ = range c {
			snn := simplenn.New(16, 8, latest.Best)
			var hist []*snake.Game
			snake.PlayFullGame(11, snn.NewGame(), &hist)

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
		log.Printf("Gen %d: %0.2f %0.2f %v", i, ss.Max, ss.Avg, dur)
		latest = ss

	}
}
