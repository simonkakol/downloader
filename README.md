# General Programming exercises #1

## Build

To build use at least Go 1.13

`go build -o downloader main.go`

## Usage

Usage:
  ./downloader [command]

Available Commands:
  help        Help about any command
  http        Download http/https S3Job to local disk
  s3          Download S3 file to local disk

For example:
  ./downloader http ​http://example.org/path-to-file.txt
or
  ./downloader s3 -b <bucketName> -i <accessKeyID> -s <secretAccesKey> ​path/to/file/100MB.bin

## Runing tests

`go test ./...`
