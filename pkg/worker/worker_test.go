package worker

import (
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"sync"
	"testing"
)

type MockJob struct {
	Err error
}

func (m *MockJob) Download(dst io.Writer) error {
	b := []byte("MOCK DOWNLOAD")
	dst.Write(b)
	return m.Err
}
func (m MockJob) FullName() string {
	return "Mockfile.txt"
}

func TestSuccessDownload(t *testing.T) {
	var wg sync.WaitGroup
	mock := MockJob{}

	jobs, results := NewPool(&wg, false, 1)
	jobs <- &mock
	res := <-results
	close(jobs)
	wg.Wait()

	content, err := ioutil.ReadFile(mock.FullName())
	if err != nil {
		t.Errorf("Not downloaded %w", err)
	}
	if string(content) != "MOCK DOWNLOAD" {
		t.Errorf("Downloaded content not saved properly %s", string(content))
	}
	if res.Error != nil {
		t.Errorf("Worker return job with error %w", res.Error)
	}
}

func TestFailedDownload(t *testing.T) {
	var wg sync.WaitGroup
	mock := MockJob{fmt.Errorf("Mocked Fail")}

	jobs, results := NewPool(&wg, false, 1)
	jobs <- &mock
	res := <-results
	close(jobs)
	wg.Wait()

	_, err := ioutil.ReadFile(mock.FullName())
	if !errors.Is(err, os.ErrNotExist) {
		t.Errorf("Failed to remove file %w", err)
	}
	if !errors.Is(res.Error, mock.Err) {
		t.Errorf("Worker response error (%s) do not match expected error (%s)", res.Error, mock.Err)
	}
}
