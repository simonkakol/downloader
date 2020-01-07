package protocols

import (
	"fmt"
	"io"
	"mime"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/simonkakol/downloader/pkg/worker"
)

//HTTPJob is single job for download one file from http/https
type HTTPJob struct {
	URL       string
	Name      string
	Extension string
	worker.WriteCounter
}

//FullName returns file name with extension
func (h HTTPJob) FullName() string {
	return h.Name + h.Extension
}

func (h *HTTPJob) setTotal(total uint64) {
	h.Bar.SetTotal(int64(total), false)
}

//Download write http remote S3Job to dst io.Writer
func (h *HTTPJob) Download(dst io.Writer) error {
	resp, err := http.Get(h.URL)
	if err != nil {
		h.Bar.SetTotal(0, true)
		return fmt.Errorf("Failed to download %s, Error: %v", h.URL, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		h.Bar.SetTotal(0, true)
		return fmt.Errorf("Error during download %s Error %s", h.URL, resp.Status)
	}

	cl, err := strconv.ParseUint(resp.Header.Get("Content-Length"), 10, 64)
	if err != nil {
		return fmt.Errorf("Invalid HTTP Content-Length header")
	}

	h.setTotal(cl)
	h.Name, h.Extension = determinateFileInfo(h.URL, resp.Header.Get("Content-Type"))

	if _, err = io.Copy(dst, io.TeeReader(resp.Body, h)); err != nil {
		return fmt.Errorf("Failed to write to a file. Error %w", err)
	}

	return nil
}

func determinateFileInfo(rawURL, contentType string) (name, extension string) {
	u, urlErr := url.Parse(rawURL)
	extensions, err := mime.ExtensionsByType(contentType)
	if err != nil && urlErr != nil {
		return "", ".bin"
	}

	path := strings.Split(u.Path, "/")
	p, err := url.PathUnescape(path[len(path)-1])
	if err != nil {
		return "", ".bin"
	}
	fInfo := strings.Split(p, ".")
	if len(fInfo) == 2 && func() bool {
		for _, ext := range extensions {
			if ext == "."+fInfo[1] {
				return true
			}
		}
		return false
	}() {
		return fInfo[0], "." + fInfo[1]
	}

	if len(extensions) > 0 {
		return fInfo[0], extensions[0]
	}

	return fInfo[0], ".bin"
}
