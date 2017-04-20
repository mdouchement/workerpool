package workerpool

import (
	"sync"
)

type registry struct {
	sync.RWMutex
	pool       *Workerpool
	jobs       map[string]*Job
	cancelChan chan string
	sts        map[string]map[string]struct{}
}

func newRegistry(p *Workerpool) *registry {
	r := &registry{
		pool:       p,
		jobs:       make(map[string]*Job, 0),
		cancelChan: make(chan string, 500),
		sts: map[string]map[string]struct{}{
			PENDING:   make(map[string]struct{}, 0),
			RUNNING:   make(map[string]struct{}, 0),
			COMPLETED: make(map[string]struct{}, 0),
			FAILED:    make(map[string]struct{}, 0),
		},
	}
	r.init()
	return r
}

func (r *registry) init() {
	go func() {
		for {
			select {
			case id := <-r.cancelChan:
				if job := r.get(id); job != nil {
					r.pool.log.Printf("workerpool: Canceling job %s\n", job.ID())
					job.Cancel()
				}
			}
		}
	}()
}

func (r *registry) add(job *Job) string {
	change := job.OnStatusChangeFunc
	job.OnStatusChangeFunc = func(j *Job) error {
		r.pool.log.Printf("workerpool: Job %s is %s\n", job.ID(), j.Status())

		change(j)

		r.updateStatus(job.ID(), j.Status())
		return nil
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

func (r *registry) statuses() map[string]interface{} {
	r.Lock()
	defer r.Unlock()

	sts := make(map[string]interface{}, 0)
	for k, v := range r.sts {
		sts[k] = len(v)
	}

	return sts
}

func (r *registry) updateStatus(id, status string) {
	r.Lock()
	defer r.Unlock()

	for _, state := range []string{PENDING, RUNNING, COMPLETED, FAILED} {
		delete(r.sts[state], id)
	}

	switch status {
	case PENDING:
		fallthrough
	case RUNNING:
		r.sts[status][id] = struct{}{}
	case COMPLETED:
		fallthrough
	case FAILED:
		delete(r.jobs, id)
		r.sts[status][id] = struct{}{} // FIXME Should take a lot of memory with a large amount of terminated jobs
	}
}
