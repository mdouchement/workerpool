package workerpool_test

import (
	"reflect"
	"testing"

	"github.com/PredicSis/workerpool"
)

func TestInit(t *testing.T) {
	job := &workerpool.Job{}
	job.Init(&nullLogger{})

	// Check missing ActionFunc
	if job.Error() != workerpool.ErrActionNotDefined {
		t.Errorf("Expected '%v' but got '%v'", workerpool.ErrActionNotDefined, job.Error())
	}

	job = &workerpool.Job{
		ActionFunc: func(j *workerpool.Job) error {
			return nil
		},
	}
	job.Init(&nullLogger{})

	if job.Error() != nil {
		t.Errorf("Expected '%v' but got '%v'", nil, job.Error())
	}
}

func TestID(t *testing.T) {
	job := &workerpool.Job{
		ActionFunc: func(j *workerpool.Job) error {
			return nil
		},
	}
	job.Init(&nullLogger{})

	if len(job.ID()) != 36 {
		t.Errorf("Expected '%v' but got '%v'", 36, len(job.ID()))
	}
}

func TestContext(t *testing.T) {
	job := &workerpool.Job{
		ActionFunc: func(j *workerpool.Job) error {
			return nil
		},
	}
	job.Init(&nullLogger{})

	ctx := job.Context()
	go job.Cancel()

	done := <-ctx.Done()

	if !reflect.DeepEqual(done, struct{}{}) {
		t.Errorf("Expected '%v' but got '%v'", struct{}{}, done)
	}
}

func TestRun(t *testing.T) {
	hasRun := false
	job := &workerpool.Job{
		ActionFunc: func(j *workerpool.Job) error {
			hasRun = true
			return nil
		},
	}
	job.Init(&nullLogger{})

	job.Run()

	if !hasRun {
		t.Errorf("Expected '%v' but got '%v'", true, hasRun)
	}

	if job.Status() != workerpool.COMPLETED {
		t.Errorf("Expected '%v' but got '%v'", job.Status(), workerpool.COMPLETED)
	}
}

func TestCancel(t *testing.T) {
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

	if !reflect.DeepEqual(done, struct{}{}) {
		t.Errorf("Expected '%v' but got '%v'", struct{}{}, done)
	}

	if !cancelled {
		t.Errorf("Expected '%v' but got '%v'", true, cancelled)
	}
}

func TestStatus(t *testing.T) {
	job := &workerpool.Job{
		ActionFunc: func(j *workerpool.Job) error {
			return nil
		},
	}
	job.Init(&nullLogger{})

	if job.Status() != workerpool.PENDING {
		t.Errorf("Expected '%v' but got '%v'", workerpool.PENDING, job.Status())
	}
}

func TestError(t *testing.T) {
	job := &workerpool.Job{}
	job.Init(&nullLogger{})

	if job.Error() != workerpool.ErrActionNotDefined {
		t.Errorf("Expected '%v' but got '%v'", workerpool.ErrActionNotDefined, job.Error())
	}
}

func TestOnStatusChangeFunc(t *testing.T) {
	changed := make(chan bool)
	status := workerpool.PENDING
	job := &workerpool.Job{
		ActionFunc: func(j *workerpool.Job) error {
			return nil
		},
		AfterFunc: func(j *workerpool.Job) error {
			defer close(changed)
			return nil
		},
		OnStatusChangeFunc: func(j *workerpool.Job) error {
			status = j.Status()
			return nil
		},
	}
	job.Init(&nullLogger{})

	go job.Run()
	<-changed

	if status != workerpool.COMPLETED {
		t.Errorf("Expected '%v' but got '%v'", workerpool.COMPLETED, status)
	}
}
