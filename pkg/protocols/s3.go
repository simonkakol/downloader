package protocols

import (
	"fmt"
	"io"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"

	"github.com/simonkakol/downloader/pkg/worker"
)

//S3Job is single job for download one item from Amazon S3
type S3Job struct {
	Item      string
	Bucket    string
	TotalSize uint64

	S3Session *session.Session
	worker.WriteCounter
}

//FullName returns file name with extension
func (h S3Job) FullName() string {
	path := strings.Split(h.Item, "/")
	return path[len(path)-1]
}

func (h *S3Job) setTotal(total uint64) {
	h.TotalSize = total
	h.Bar.SetTotal(int64(total), false)
}

//Download write s3 item to dst io.Writer
func (h *S3Job) Download(dst io.Writer) error {
	downloader := s3manager.NewDownloader(h.S3Session)

	head, err := downloader.S3.HeadObject(&s3.HeadObjectInput{
		Bucket: aws.String(h.Bucket),
		Key:    aws.String(h.Item),
	})
	if err != nil {
		return fmt.Errorf("Unable get item info %s, Error %w", h.Item, err)
	}
	h.setTotal(uint64(*head.ContentLength))

	// force single threaded download for single file download
	downloader.Concurrency = 1
	_, err = downloader.Download(
		&fakeWriterAt{io.MultiWriter(dst, h)},
		&s3.GetObjectInput{
			Bucket: aws.String(h.Bucket),
			Key:    aws.String(h.Item),
		})
	if err != nil {
		return fmt.Errorf("Unable to download item %s, %w", h.Item, err)
	}
	return nil
}

type fakeWriterAt struct {
	w io.Writer
}

func (fw fakeWriterAt) WriteAt(p []byte, offset int64) (n int, err error) {
	// ignore 'offset' because we forced sequential downloads (Concurrency = 1)
	return fw.w.Write(p)
}
