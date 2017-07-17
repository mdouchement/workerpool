package workerpool

import (
	"context"
	"errors"
	"fmt"
	"runtime"
	"sync"

	uuid "github.com/satori/go.uuid"

	tomb "gopkg.in/tomb.v2"
)

const (
	// PENDING state when job is queued.
	PENDING = "pending"
	// RUNNING state when job is currently executed.
	RUNNING = "running"
	// COMPLETED state when job is terminated whithout error.
	COMPLETED = "completed"
	// FAILED state when job is terminated with error.
	FAILED = "failed"
	// CANCELLED state when job is cancelled by the user.
	CANCELLED = "cancelled"
)

type (
	// A JobAction defines the format of the function holded by a job.
	JobAction func(j *Job) error
)

var (
	// ErrActionNotDefined is returned when there is no ActionFunc declared in the job.
	ErrActionNotDefined = errors.New("Action function is not defined")
	// EmptyAction defines a no-op action for job's func.
	EmptyAction = func(j *Job) error {
		return nil
	}
)

// A Job performs actions.
type Job struct {
	recover            func()
	mu                 sync.Mutex
	id                 string
	status             string
	tomb               *tomb.Tomb
	err                error
	OnStatusChangeFunc JobAction
	BeforeFunc         JobAction
	ActionFunc         JobAction
	AfterFunc          JobAction
	CancelFunc         JobAction
}

// Init initializes the job.
// It should be only called by the worker.
func (j *Job) Init(log Logger) {
	j.id = uuid.NewV4().String()
	j.mu = sync.Mutex{}

	j.recover = func() {
		if r := recover(); r != nil {
			var err error
			switch r := r.(type) {
			case error:
				err = r
			default:
				err = fmt.Errorf("%v", r)
			}
			stack := make([]byte, 4<<10)
			length := runtime.Stack(stack, true)
			log.Printf("[PANIC RECOVER] %s %s\n", err, stack[:length])

			j.setError(err)
			j.setStatus(FAILED)
			j.AfterFunc(j)
		}
	}

	t, _ := tomb.WithContext(context.Background())
	j.tomb = t

	if j.ActionFunc == nil {
		j.setError(ErrActionNotDefined)
		return
	}

	if j.CancelFunc == nil {
		j.CancelFunc = EmptyAction
	}

	if j.OnStatusChangeFunc == nil {
		j.OnStatusChangeFunc = EmptyAction
	}

	if j.BeforeFunc == nil {
		j.BeforeFunc = EmptyAction
	}

	if j.AfterFunc == nil {
		j.AfterFunc = EmptyAction
	}

	j.setStatus(PENDING)
}

// ID returns the job's identifier.
func (j *Job) ID() string {
	return j.id
}

// Context returns a new context for the job.
func (j *Job) Context() context.Context {
	return j.tomb.Context(nil)
}

// Run starts the job.
func (j *Job) Run() {
	if j.Error() != nil {
		return
	}

	j.tomb.Go(j.run)
	j.tomb.Wait()
}

// Cancel stops the job execution.
func (j *Job) Cancel() {
	defer j.recover()

	j.CancelFunc(j)
	j.tomb.Kill(nil)
	j.err = j.tomb.Wait()
}

// Status returns the job's status.
func (j *Job) Status() string {
	j.mu.Lock()
	defer j.mu.Unlock()

	return j.status
}

func (j *Job) setStatus(status string) {
	defer j.OnStatusChangeFunc(j)

	j.mu.Lock()
	defer j.mu.Unlock()

	j.status = status
}

// Error returns the job's error if exists.
func (j *Job) Error() error {
	j.mu.Lock()
	defer j.mu.Unlock()

	return j.err
}

func (j *Job) setError(err error) {
	j.mu.Lock()
	defer j.mu.Unlock()

	j.err = err
}

func (j *Job) run() error {
	defer j.recover()

	defer j.AfterFunc(j)
	j.setStatus(RUNNING)
	j.BeforeFunc(j)

	var err error
	if err = j.ActionFunc(j); err != nil {
		j.setError(err)
		j.setStatus(FAILED)
	} else {
		j.setStatus(COMPLETED)
	}

	return err
}
