package workerpool_test

import (
	"errors"
	"testing"
	"time"

	"github.com/mdouchement/workerpool"
	"github.com/stretchr/testify/assert"
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

	assert.Regexp(t, `^[a-fA-F0-9]{8}-[a-fA-F0-9]{4}-4[a-fA-F0-9]{3}-[8|9|aA|bB][a-fA-F0-9]{3}-[a-fA-F0-9]{12}$`, jobID)
}

func TestGetJob(t *testing.T) {
	lock := make(chan bool)

	job := &workerpool.Job{
		ActionFunc: func(j *workerpool.Job) error {
			<-lock
			return nil
		},
	}
	jobID := workerpool.Send(job)

	job2 := workerpool.GetJob(jobID)
	assert.Equal(t, job, job2)

	close(lock)
}

func TestGetJobStatus(t *testing.T) {
	lock := make(chan bool)

	job := &workerpool.Job{
		ActionFunc: func(j *workerpool.Job) error {
			<-lock
			return nil
		},
	}
	jobID := workerpool.Send(job)

	status := workerpool.GetJobStatus(jobID)
	assert.Equal(t, status, workerpool.PENDING)

	close(lock)
}

func TestOnError(t *testing.T) {
	count := 0
	job := &workerpool.Job{
		AfterFunc: func(j *workerpool.Job) error {
			return errors.New("42")
		},

		ActionFunc: workerpool.EmptyAction,

		OnStatusChangeFunc: func(j *workerpool.Job) error {
			if j.Status() == workerpool.FAILED {
				count++
			}
			return nil
		},
	}
	workerpool.Send(job)

	<-job.Context().Done() // Block until done

	assert.Equal(t, 1, count)
}

func TestCancelJob(t *testing.T) {
	stopped := make(chan bool)

	job := &workerpool.Job{
		ActionFunc: func(j *workerpool.Job) error {
			defer close(stopped)

			ctx := j.Context()

			<-ctx.Done() // Block until done
			return nil
		},
	}
	jobID := workerpool.Send(job)

	workerpool.CancelJob(jobID)
	<-stopped

	job2 := workerpool.GetJob(jobID)
	assert.Nil(t, job2)
}

func TestGetPoolSize(t *testing.T) {
	size := workerpool.GetPoolSize()

	// Default value is 1
	assert.Equal(t, 1, size)
}

func TestSetPoolSize(t *testing.T) {
	size := 42
	workerpool.SetPoolSize(size)

	size2 := workerpool.GetPoolSize()

	assert.Equal(t, size, size2)
}

func TestShutdown(t *testing.T) {
	pool := workerpool.New(2, 10)
	pool.SetLogger(&nullLogger{})

	lock1 := make(chan bool)
	lock2 := make(chan bool)
	sleep := 250 * time.Millisecond
	var elapsed time.Duration

	job := &workerpool.Job{
		BeforeFunc: func(j *workerpool.Job) error {
			close(lock1)
			return nil
		},
		ActionFunc: func(j *workerpool.Job) error {
			start := time.Now()
			<-lock2
			time.Sleep(sleep)
			elapsed = time.Since(start)
			return nil
		},
	}
	pool.Send(job)

	<-lock1
	close(lock2)
	pool.Shutdown()

	assert.InEpsilon(t, sleep.Milliseconds(), elapsed.Milliseconds(), 10)
}

func TestGetJobsMetrics(t *testing.T) {
	workerpool.RecordJobsMetrics(true)
	defer workerpool.RecordJobsMetrics(false)

	m := map[string]interface{}{
		workerpool.PENDING:   0,
		workerpool.RUNNING:   0,
		workerpool.COMPLETED: 0,
		workerpool.FAILED:    0,
		workerpool.CANCELLED: 0,
	}

	metrics := workerpool.GetJobsMetrics()
	assert.Equal(t, m, metrics)

	//
	//

	lock1 := make(chan bool)
	lock2 := make(chan bool)

	workerpool.Send(&workerpool.Job{
		BeforeFunc: func(j *workerpool.Job) error {
			close(lock1)
			return nil
		},
		ActionFunc: func(j *workerpool.Job) error {
			<-lock2
			return nil
		},
	})
	m[workerpool.RUNNING] = 1

	<-lock1
	metrics = workerpool.GetJobsMetrics()
	assert.Equal(t, m, metrics)

	close(lock2)
}
