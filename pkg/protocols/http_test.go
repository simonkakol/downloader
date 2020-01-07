package protocols

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	"regexp"
	"testing"

	"github.com/simonkakol/downloader/pkg/worker"
	"github.com/vbauerster/mpb"
)

func TestDownload(t *testing.T) {
	tables := []struct {
		servStatus  int
		servBody    string
		contentType string

		errorMatch *regexp.Regexp
	}{
		{http.StatusOK, "David Bowie", "text/plain", nil},
		{http.StatusNoContent, "", "text/html", regexp.MustCompile("^Error during download .+ Error 204 No Content$")},
		{http.StatusNotFound, "David ", "text/plain", regexp.MustCompile("^Error during download .+ Error 404 Not Found$")},
		{http.StatusBadRequest, " Bowie", "text/plain", regexp.MustCompile("^Error during download .+ Error 400 Bad Request$")},
		{http.StatusForbidden, "idie", "text/plain", regexp.MustCompile("^Error during download .+ Error 403 Forbidden$")},
	}

	for i, tt := range tables {
		t.Run(fmt.Sprintf("Download test %d", i), func(t *testing.T) {

			testServer := httptest.NewServer(http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
				res.WriteHeader(tt.servStatus)
				res.Header().Set("Content-Type", tt.contentType)
				res.Write([]byte(tt.servBody))
			}))

			job := HTTPJob{
				URL:          testServer.URL,
				Name:         "temp_0",
				Extension:    ".tmp",
				WriteCounter: mockWriteCounter,
			}

			mockFile := &bytes.Buffer{}
			err := job.Download(mockFile)
			testServer.Close()

			if tt.errorMatch != nil && !(tt.errorMatch.MatchString(err.Error())) {
				t.Errorf("Download error %s did not match %s", err, tt.errorMatch)
				t.FailNow()
			}
			if tt.errorMatch == nil && mockFile.String() != tt.servBody {
				t.Errorf("Download contend mismatch wanted %s got %s", tt.servBody, mockFile.String())
			}
			if tt.errorMatch == nil && job.Extension == ".tmp" {
				t.Errorf("Extension not changed")
			}
		})
	}
}

func TestFilenameParse(t *testing.T) {
	t.Parallel()
	tables := []struct {
		URL          string
		contentType  string
		expectedName string
		expectedExt  string
	}{
		{"http://example.org/path-to-file.txt", "text/plain", "path-to-file", ".txt"},
		{"http://example.org/path-to/file.jpg", "image/jpeg", "file", ".jpg"},
		{"http://example.org/path-to-file", "image/png", "path-to-file", ".png"},
		{"http://example.org/path-to-file", "", "path-to-file", ".bin"},
	}
	for _, tt := range tables {
		t.Run(fmt.Sprintf("Parse %s", tt.URL), func(t *testing.T) {
			tt := tt
			name, ext := determinateFileInfo(tt.URL, tt.contentType)

			if name != tt.expectedName {
				t.Errorf("File name expected %s got %s", name, tt.expectedName)
			}
			if ext != tt.expectedExt {
				t.Errorf("Extension expected %s got %s", ext, tt.expectedExt)
			}
		})
	}
}

var mockWriteCounter = worker.WriteCounter{Bar: mpb.New().AddBar(1)}
