package worker

import (
	"fmt"
	"io"
	"os"
	"sync"

	"github.com/vbauerster/mpb"
)

const maxWorkers = 10

//Downloader is interface that wraps basic download methods
type Downloader interface {
	Download(dst io.Writer) error
	FullName() string
}

//Result is struct for worker job result
type Result struct {
	Error error
}

//NewPool create job input channels and worker for processing downloads
func NewPool(wg *sync.WaitGroup, isConcurrent bool, totalSize int) (chan<- Downloader, chan Result) {
	jobs := make(chan Downloader)
	results := make(chan Result)

	poolSize := workerPoolSize(isConcurrent, totalSize)
	wg.Add(poolSize)

	// Start workers
	for i := 0; i < poolSize; i++ {
		go worker(jobs, results, wg)
	}

	return jobs, results
}

func worker(jobs <-chan Downloader, results chan<- Result, wg *sync.WaitGroup) {
	for job := range jobs {
		tmpName := job.FullName()
		file, err := os.Create(tmpName)
		if err != nil {
			results <- Result{fmt.Errorf("Failed to create a file for %s, error: %w", tmpName, err)}
			continue
		}

		err = job.Download(file)
		file.Close()
		if err != nil {
			results <- Result{err}
			os.Remove(tmpName)
			continue
		}

		err = os.Rename(tmpName, job.FullName())
		if err != nil {
			results <- Result{
				fmt.Errorf("Failed to rename file %s to %s, error: %w", tmpName, job.FullName(), err),
			}
			continue
		}

		results <- Result{}
	}

	wg.Done()
}

func workerPoolSize(isConcurrent bool, jobSize int) int {
	switch {
	case isConcurrent && jobSize > maxWorkers:
		return maxWorkers
	case isConcurrent:
		return jobSize
	}

	return 1
}

//WriteCounter implements writing progress to progress bar
type WriteCounter struct {
	Bar *mpb.Bar
}

func (wc *WriteCounter) Write(p []byte) (int, error) {
	n := len(p)
	wc.Bar.IncrBy(n)
	return n, nil
}
