package workerpool

import (
	"sync"
)

type registry struct {
	sync.RWMutex
	pool       *Workerpool
	jobs       map[string]*Job
	cancelChan chan string
	stsEnabled bool
	sts        map[string]map[string]struct{}
}

func newRegistry(p *Workerpool) *registry {
	r := &registry{
		pool:       p,
		jobs:       make(map[string]*Job),
		cancelChan: make(chan string, 500),
		sts: map[string]map[string]struct{}{
			PENDING:   make(map[string]struct{}),
			RUNNING:   make(map[string]struct{}),
			COMPLETED: make(map[string]struct{}),
			FAILED:    make(map[string]struct{}),
			CANCELLED: make(map[string]struct{}),
		},
	}
	r.init()
	return r
}

func (r *registry) init() {
	go func() {
		for id := range r.cancelChan {
			if job := r.get(id); job != nil {
				r.pool.log.Printf("workerpool: Canceling job %s", job.ID())
				job.Cancel()
				job.setStatus(CANCELLED)
			}
		}
	}()
}

func (r *registry) add(job *Job) string {
	change := job.OnStatusChangeFunc
	job.OnStatusChangeFunc = func(j *Job) error {
		r.pool.log.Printf("workerpool: Job %s is %s", job.ID(), j.Status())

		defer r.updateStatus(job.ID(), j.Status())
		return change(j)
	}

	r.Lock()
	defer r.Unlock()

	r.jobs[job.ID()] = job

	return job.ID()
}

func (r *registry) get(id string) *Job {
	r.Lock()
	defer r.Unlock()

	return r.jobs[id]
}

func (r *registry) cancel(id string) {
	r.cancelChan <- id
}

func (r *registry) statusRecording(enabled bool) {
	r.Lock()
	defer r.Unlock()

	r.stsEnabled = enabled
}

func (r *registry) statuses() map[string]interface{} {
	r.Lock()
	defer r.Unlock()

	sts := make(map[string]interface{})
	for k, v := range r.sts {
		sts[k] = len(v)
	}

	return sts
}

func (r *registry) updateStatus(id, status string) {
	r.Lock()
	defer r.Unlock()

	if !r.stsEnabled {
		if status == CANCELLED || status == COMPLETED || status == FAILED {
			delete(r.jobs, id)
		}
		return
	}

	for _, state := range []string{PENDING, RUNNING, COMPLETED, FAILED} {
		delete(r.sts[state], id)
	}

	switch status {
	case PENDING:
		fallthrough
	case RUNNING:
		r.sts[status][id] = struct{}{}
	case CANCELLED:
		fallthrough
	case COMPLETED:
		fallthrough
	case FAILED:
		delete(r.jobs, id)
		r.sts[status][id] = struct{}{} // FIXME Should take a lot of memory with a large amount of terminated jobs
	}
}
