# Workerpool

[![CircleCI](https://circleci.com/gh/mdouchement/workerpool.svg?style=shield)](https://circleci.com/gh/mdouchement/workerpool)
[![GoDoc](https://img.shields.io/badge/godoc-reference-blue.svg)](https://godoc.org/github.com/mdouchement/workerpool)
[![Go Report Card](https://goreportcard.com/badge/github.com/mdouchement/workerpool)](https://goreportcard.com/report/github.com/mdouchement/workerpool)
[![License](https://img.shields.io/github/license/mdouchement/workerpool.svg)](http://opensource.org/licenses/MIT)

Asynchronous in-memory workers for Go based projects.

It is based on [workerpool](https://github.com/dc0d/workerpool), a lightweight workers solution, and [tomb](https://github.com/go-tomb/tomb/tree/v2) that handles clean goroutine tracking and termination.

This project aimed to provide the following points:
- Dynamic concurrency (expand and shrink the number of workers)
- Trackable background jobs
  - Cancelable jobs
  - Jobs' status

## Usage

```sh
go get github.com/mdouchement/workerpool
```

```go
package main

import (
	"fmt"

	"github.com/mdouchement/workerpool"
)

func main() {
	// It is not mandatory to implement all *Func.
	job := &workerpool.Job{
		// Triggered when the job's status change
		OnStatusChangeFunc: func(j *workerpool.Job) error {
			fmt.Println("Status changed to:", j.Status())
			return nil
		},

		// Before starting job
		BeforeFunc: func(j *workerpool.Job) error {
			fmt.Println("Before starting job")
			return nil
		},

		// Task to perform
		ActionFunc: func(j *workerpool.Job) error {
			fmt.Println("Starting job")
			return nil
		},

		// When job is completed
		AfterFunc: func(j *workerpool.Job) error {
			fmt.Println("After job")
			return nil
		},

		// Triggered when workerpool.Cancel("job_id")
		// You can use j.Context() to forward cancellation signal.
		CancelFunc: func(j *workerpool.Job) error {
			fmt.Println("Execute cleanup function")
			return nil
		},

    // Triggered when an error or panic occurred.
    ErrHandler: func(err error, panic bool) {
      // Do something with the error.
    },
	}

	id := workerpool.Send(job)
	fmt.Println("Job id:", id)
}
```

## License

**MIT**


## Contributing

All PRs are welcome.

1. Fork it
2. Create your feature branch (git checkout -b my-new-feature)
3. Commit your changes (git commit -am 'Add some feature')
5. Push to the branch (git push origin my-new-feature)
6. Create new Pull Request

As possible, run the following commands to format and lint the code:

```sh
# Format
find . -name '*.go' -not -path './vendor*' -exec gofmt -s -w {} \;

# Lint
gometalinter --config=gometalinter.json ./...
```
