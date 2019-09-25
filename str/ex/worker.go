package main

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/sspencer/str"
)

var rnd *rand.Rand

func init() {
	rnd = rand.New(rand.NewSource(time.Now().UnixNano()))
}

type myWorker int

func (m myWorker) StringWork(fn string) string {
	fmt.Println("processing", fn)
	time.Sleep(time.Duration(rnd.Float32()*1000+200) * time.Millisecond)
	if rnd.Float32() < 0.25 {
		fmt.Println("error processing", fn)
		return "" // return empty string on error
	}

	return fn + ".processed"
}

func main() {
	var input []string
	for i := 0; i < 23; i++ {
		input = append(input, fmt.Sprintf("/tmp/f%d.txt", i))
	}

	var w myWorker

	results := str.Worker(4, input, w)
	for _, f := range results {
		fmt.Println("==>", f)
	}
}
