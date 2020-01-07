package cmd

import (
	"fmt"
	"sync"

	"github.com/spf13/cobra"
	"github.com/vbauerster/mpb"
	"github.com/vbauerster/mpb/decor"

	"github.com/simonkakol/downloader/pkg/protocols"
	"github.com/simonkakol/downloader/pkg/worker"
)

func init() {
	httpCmd.Flags().BoolVarP(&concurrent, "concurrent", "c", false, "Download files concurrently")
	rootCmd.AddCommand(httpCmd)
}

var concurrent bool
var httpCmd = &cobra.Command{
	Use:   "http [url(s)]",
	Short: "Download http/https S3Job to local disk",
	Args:  cobra.MinimumNArgs(1),

	Run: func(cmd *cobra.Command, args []string) {
		var wg sync.WaitGroup
		prg := mpb.New(mpb.WithWaitGroup(&wg))
		jobs, results := worker.NewPool(&wg, concurrent, len(args))

		// Fetch results
		errors := []error{}
		go func() {
			for result := range results {
				errors = append(errors, result.Error)
			}
		}()

		for i, url := range args {
			HTTPJob := protocols.HTTPJob{
				URL:       url,
				Name:      fmt.Sprintf("tmp_%d", i),
				Extension: ".tmp",
				WriteCounter: worker.WriteCounter{
					Bar: prg.AddBar(
						initialFileSize,
						mpb.PrependDecorators(decor.Counters(decor.UnitKiB, "% .1f / % .1f")),
						mpb.AppendDecorators(decor.Percentage(decor.WC{W: 5})),
					),
				},
			}

			jobs <- &HTTPJob
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
