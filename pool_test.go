package workerpool_test

import (
	"reflect"
	"testing"
	"time"

	"github.com/mdouchement/workerpool"
)

// All tests are chained in this order.
// Any change can impact the other tests.

func TestSend(t *testing.T) {
	job := &workerpool.Job{
		ActionFunc: func(j *workerpool.Job) error {
			return nil
		},
	}
	jobID := workerpool.Send(job)

	if len(jobID) != 36 { // UUID's length
		t.Errorf("Expected '%v' but got '%v'", 36, len(jobID))
	}
}

func TestGetJob(t *testing.T) {
	job := &workerpool.Job{
		ActionFunc: func(j *workerpool.Job) error {
			time.Sleep(2 * time.Second)
			return nil
		},
	}
	jobID := workerpool.Send(job)

	job2 := workerpool.GetJob(jobID)

	if !reflect.DeepEqual(job, job2) {
		t.Errorf("Expected '%v' but got '%v'", job, job2)
	}
}

func TestGetJobStatus(t *testing.T) {
	job := &workerpool.Job{
		ActionFunc: func(j *workerpool.Job) error {
			time.Sleep(2 * time.Second)
			return nil
		},
	}
	jobID := workerpool.Send(job)

	status := workerpool.GetJobStatus(jobID)

	if status != workerpool.PENDING {
		t.Errorf("Expected '%v' but got '%v'", workerpool.RUNNING, status)
	}
}

func TestCancelJob(t *testing.T) {
	stopped := make(chan bool)

	job := &workerpool.Job{
		ActionFunc: func(j *workerpool.Job) error {
			defer close(stopped)

			ctx := j.Context()

			for {
				select {
				case <-ctx.Done(): // Block until done
					return nil
				}
			}
		},
	}
	jobID := workerpool.Send(job)

	workerpool.CancelJob(jobID)
	<-stopped

	job2 := workerpool.GetJob(jobID)
	if job2 != nil {
		t.Errorf("Expected '%v' but got '%v'", nil, job2)
	}
}

func TestGetPoolSize(t *testing.T) {
	size := workerpool.GetPoolSize()

	// Default value is 1
	if size != 1 {
		t.Errorf("Expected '%v' but got '%v'", 1, size)
	}
}

func TestSetPoolSize(t *testing.T) {
	size := 42
	workerpool.SetPoolSize(size)

	size2 := workerpool.GetPoolSize()

	if size2 != size {
		t.Errorf("Expected '%v' but got '%v'", size, size2)
	}
}

func TestGetJobsMetrics(t *testing.T) {
	time.Sleep(4 * time.Second) // Wait running jobs
	m := map[string]interface{}{
		workerpool.PENDING:   0,
		workerpool.RUNNING:   0,
		workerpool.COMPLETED: 3,
		workerpool.FAILED:    0,
		workerpool.CANCELLED: 1,
	}
	metrics := workerpool.GetJobsMetrics()

	if !reflect.DeepEqual(metrics, m) {
		t.Errorf("Expected '%v' but got '%v'", m, metrics)
	}
}
