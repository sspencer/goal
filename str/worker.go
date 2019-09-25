package str

import (
	"sync"
)

const MaxThreads = 100

// StringWorker describes the interface to implement when
// your work can be paralleized.
type StringWorker interface {
	StringWork(string) string
}

// Worker concurrently calls the string worker up to 'numThreads' at a time.  The
// worker either returns a non-empty string to make it part of output, or an empty
// string if the work should be ignored.  NOTE: there may not be be a 1-1 mapping
// between input and output, and they may not be in the same order.  A good use
// for this is where the "strings" in question are file paths and a longer operation
// is transforming an input file to an output file.  If there was an error processing
// the file, an empty string is returned.
func Worker(numWorkers int, input []string, worker StringWorker) (output []string) {
	// sanity checks

	// short circuit on empty input
	if len(input) == 0 {
		return []string{}
	}

	numWorkers = boundWorkers(numWorkers, len(input))

	// synchronize writes into output
	var mutex sync.Mutex

	// create (n) workers
	sem := make(chan bool, numWorkers)

	for _, s := range input {
		sem <- true // blocks after (n)

		go func(inputFile string) {
			fn := worker.StringWork(inputFile)
			if fn != "" {
				mutex.Lock()
				output = append(output, fn)
				mutex.Unlock()
			}

			<-sem // release a slot
		}(s)
	}

	// wait until last (n) matches complete
	for i := 0; i < cap(sem); i++ {
		sem <- true
	}

	return output
}

// boundWorkers caps the number of threads the user requested, so it is
// no more than the amount of work, or MaxThreads.
func boundWorkers(numWorkers, numTasks int) int {
	// at least 1 "thread"
	if numWorkers < 1 {
		return 1
	}

	// no more theads than work
	if numWorkers > numTasks {
		numWorkers = numTasks
	}

	// yet, a reasonable cap on number of threads
	if numWorkers > MaxThreads {
		numWorkers = MaxThreads
	}

	return numWorkers
}
