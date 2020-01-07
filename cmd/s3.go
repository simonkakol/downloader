package cmd

import (
	"fmt"
	"sync"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/spf13/cobra"
	"github.com/vbauerster/mpb"
	"github.com/vbauerster/mpb/decor"

	"github.com/simonkakol/downloader/pkg/protocols"
	"github.com/simonkakol/downloader/pkg/worker"
)

func init() {
	s3Cmd.Flags().BoolVarP(&concurrent, "concurrent", "c", false, "Download up to %d files concurrently")
	s3Cmd.Flags().StringVarP(&region, "region", "r", "eu-central-1", "S3 region - defaults to eu-central-1")

	s3Cmd.Flags().StringVarP(&bucket, "bucket", "b", "", "S3 bucket")
	s3Cmd.MarkFlagRequired("bucket")

	s3Cmd.Flags().StringVarP(&id, "id", "i", "", "AWS_ACCESS_KEY_ID")
	s3Cmd.MarkFlagRequired("id")

	s3Cmd.Flags().StringVarP(&secret, "secret", "s", "", "AWS_SECRET_ACCESS_KEY")
	s3Cmd.MarkFlagRequired("secret")
	rootCmd.AddCommand(s3Cmd)
}

var (
	region string
	bucket string
	id     string
	secret string
)

var s3Cmd = &cobra.Command{
	Use:   "s3 [items]",
	Short: "Download S3 file to local disk",
	Args:  cobra.MinimumNArgs(1),

	Run: func(cmd *cobra.Command, args []string) {
		sess := session.Must(session.NewSession(&aws.Config{
			Credentials: credentials.NewStaticCredentials(id, secret, ""),
			Region:      aws.String(region),
		}))
		var wg sync.WaitGroup
		prg := mpb.New(mpb.WithWaitGroup(&wg))
		jobs, results := worker.NewPool(&wg, concurrent, len(args))

		// Fetch results
		var errors []error
		go func() {
			for result := range results {
				errors = append(errors, result.Error)
			}
		}()

		for _, item := range args {
			s3Job := protocols.S3Job{
				Item:      item,
				Bucket:    bucket,
				S3Session: sess.Copy(),
				WriteCounter: worker.WriteCounter{
					Bar: prg.AddBar(
						initialFileSize,
						mpb.PrependDecorators(decor.Counters(decor.UnitKiB, "% .1f / % .1f")),
						mpb.AppendDecorators(decor.Percentage(decor.WC{W: 5})),
					),
				},
			}

			jobs <- &s3Job
		}

		// Stop workers
		close(jobs)
		prg.Wait()
		// Safely close receiving channel after stopping workers
		close(results)

		// Print errors at the end
		for _, err := range errors {
			if err != nil {
				fmt.Println(err)
			}
		}
	},
}
