package workerpool

import (
	"context"
	"errors"
	"fmt"
	"runtime"
	"sync"

	uuid "github.com/gofrs/uuid"
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
	isInitialized      bool
	once               sync.Once
	mu                 sync.Mutex
	id                 string
	status             string
	tomb               *tomb.Tomb
	err                error
	ctx                context.Context
	OnStatusChangeFunc JobAction
	BeforeFunc         JobAction
	ActionFunc         JobAction
	AfterFunc          JobAction
	CancelFunc         JobAction
	// ErrHandler is executed when an error or a panic occur.
	ErrHandler func(j *Job, err error, panic bool)
}

// Init initializes the job.
// It should be only called by the worker.
func (j *Job) Init(log Logger) {
	j.isInitialized = true
	j.mu = sync.Mutex{}

	if j.ErrHandler == nil {
		j.ErrHandler = func(j *Job, err error, panic bool) {
			if panic {
				stack := make([]byte, 4<<10)
				length := runtime.Stack(stack, true)
				log.Printf("[PANIC RECOVER] %s %s\n", err, stack[:length])
				return
			}

			log.Println(err)
		}
	}

	j.recover = func() {
		if r := recover(); r != nil {
			var err error
			switch r := r.(type) {
			case error:
				err = r
			default:
				err = fmt.Errorf("%v", r)
			}

			j.setError(err, true)
		}
	}

	if j.ctx == nil {
		j.ctx = context.Background()
	}
	t, _ := tomb.WithContext(j.ctx)
	j.tomb = t

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

	if j.ActionFunc == nil {
		j.setError(ErrActionNotDefined, false)
		return
	}

	j.setStatus(PENDING)
}

// ID returns the job's identifier.
func (j *Job) ID() string {
	j.once.Do(func() {
		j.id = uuid.Must(uuid.NewV4()).String()
	})

	return j.id
}

// Context returns a new context for the job.
func (j *Job) Context() context.Context {
	return j.tomb.Context(j.ctx)
}

// SetContext sets the given context to the job.
// It panics if the job is queued.
func (j *Job) SetContext(ctx context.Context) {
	if j.isInitialized {
		panic("Context cannot be updated for a queued job")
	}
	j.ctx = ctx
}

// Run starts the job.
func (j *Job) Run() {
	if j.Error() != nil {
		return
	}

	j.tomb.Go(j.run)
	_ = j.tomb.Wait() // err handled in `j.run`
}

// Cancel stops the job execution.
func (j *Job) Cancel() {
	defer j.recover()

	err := j.CancelFunc(j)
	if err != nil {
		j.setError(err, false)
	}

	j.tomb.Kill(nil)
	_ = j.tomb.Wait() // err handled in `j.run`
}

// Status returns the job's status.
func (j *Job) Status() string {
	j.mu.Lock()
	defer j.mu.Unlock()

	return j.status
}

func (j *Job) setStatus(status string) {
	defer func() {
		if err := j.OnStatusChangeFunc(j); err != nil {
			if j.Error() == nil {
				j.setError(err, false)
				return
			}
			j.ErrHandler(j, err, false) // Job already failed, just run the latest thing we can do.
		}
	}()

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

func (j *Job) setError(err error, panic bool) {
	defer j.ErrHandler(j, err, panic)
	defer j.setStatus(FAILED) // Executed before ErrHandler

	j.mu.Lock()
	defer j.mu.Unlock()

	j.err = err
}

func (j *Job) run() (err error) {
	defer j.recover()

	j.setStatus(RUNNING)
	if j.Error() != nil {
		return j.Error()
	}

	if err = j.BeforeFunc(j); err != nil {
		j.setError(err, false)
		return
	}

	err = j.ActionFunc(j)
	if err != nil {
		j.setError(err, false)
		return
	}

	err = j.AfterFunc(j)
	if err != nil {
		j.setError(err, false)
		return
	}

	j.setStatus(COMPLETED)
	return nil
}
