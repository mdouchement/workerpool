package workerpool_test

import (
	"testing"

	"github.com/mdouchement/workerpool"
	"github.com/stretchr/testify/assert"
)

func TestJob_Init(t *testing.T) {
	job := &workerpool.Job{}
	job.Init(&nullLogger{})

	// Check missing ActionFunc
	assert.Equal(t, workerpool.ErrActionNotDefined, job.Error())

	job = &workerpool.Job{
		ActionFunc: func(j *workerpool.Job) error {
			return nil
		},
	}
	job.Init(&nullLogger{})

	assert.Nil(t, job.Error())
}

func TestJob_ID(t *testing.T) {
	job := &workerpool.Job{
		ActionFunc: func(j *workerpool.Job) error {
			return nil
		},
	}
	job.Init(&nullLogger{})

	assert.Regexp(t, `^[a-fA-F0-9]{8}-[a-fA-F0-9]{4}-4[a-fA-F0-9]{3}-[8|9|aA|bB][a-fA-F0-9]{3}-[a-fA-F0-9]{12}$`, job.ID())
}

func TestJob_Context(t *testing.T) {
	job := &workerpool.Job{
		ActionFunc: func(j *workerpool.Job) error {
			return nil
		},
	}
	job.Init(&nullLogger{})

	ctx := job.Context()
	go job.Cancel()

	done := <-ctx.Done()

	assert.Equal(t, struct{}{}, done)
}

func TestJob_Run(t *testing.T) {
	hasRun := false
	job := &workerpool.Job{
		ActionFunc: func(j *workerpool.Job) error {
			hasRun = true
			return nil
		},
	}
	job.Init(&nullLogger{})

	job.Run()

	assert.True(t, hasRun)
	assert.Equal(t, workerpool.COMPLETED, job.Status())
}

func TestJob_Cancel(t *testing.T) {
	cancelled := false
	job := &workerpool.Job{
		ActionFunc: func(j *workerpool.Job) error {
			return nil
		},
		CancelFunc: func(j *workerpool.Job) error {
			cancelled = true
			return nil
		},
	}
	job.Init(&nullLogger{})

	ctx := job.Context()
	go job.Cancel()

	done := <-ctx.Done()

	assert.Equal(t, struct{}{}, done)
	assert.True(t, cancelled)
}

func TestJob_Status(t *testing.T) {
	job := &workerpool.Job{
		ActionFunc: func(j *workerpool.Job) error {
			return nil
		},
	}
	job.Init(&nullLogger{})

	assert.Equal(t, workerpool.PENDING, job.Status())
}

func TestJob_Error(t *testing.T) {
	job := &workerpool.Job{}
	job.Init(&nullLogger{})

	assert.Equal(t, workerpool.ErrActionNotDefined, job.Error())
}

func TestJob_ErrHandler(t *testing.T) {
	var errFromHandler error
	job := &workerpool.Job{
		ErrHandler: func(err error, panic bool) {
			errFromHandler = err
		},
	}
	job.Init(&nullLogger{})

	assert.Equal(t, workerpool.ErrActionNotDefined, errFromHandler)
}

func TestJob_OnStatusChangeFunc(t *testing.T) {
	status := workerpool.PENDING
	job := &workerpool.Job{
		ActionFunc: func(j *workerpool.Job) error {
			return nil
		},
		OnStatusChangeFunc: func(j *workerpool.Job) error {
			status = j.Status()
			return nil
		},
	}
	job.Init(&nullLogger{})

	go job.Run()
	<-job.Context().Done()

	assert.Equal(t, workerpool.COMPLETED, status)
}
