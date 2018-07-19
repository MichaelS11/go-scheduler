package scheduler

import (
	"sync"
	"sync/atomic"
	"time"
)

// NewScheduler creates a new Scheduler
func NewScheduler() *Scheduler {
	return &Scheduler{
		jobs:               make(map[string]*jobStruct, 1),
		jobsRWMutex:        &sync.RWMutex{},
		chanJobsNotStopped: make(chan struct{}, 2),
	}
}

// Jobs returns all job names
func (s *Scheduler) Jobs() []string {
	s.jobsRWMutex.RLock()
	names := make([]string, 0, len(s.jobs))
	for name := range s.jobs {
		names = append(names, name)
	}
	s.jobsRWMutex.RUnlock()
	return names
}

// StopAll stops all job from running again.
// Does not kill any running jobs.
func (s *Scheduler) StopAll() {
	s.jobsRWMutex.RLock()
	for _, job := range s.jobs {
		job.mutex.Lock()
		s.stop(job)
		job.mutex.Unlock()
	}
	s.jobsRWMutex.RUnlock()
}

// StopAllWait stops all job from running again and waits till they have all stopped or the timeout duration has passed.
// Does not kill any running jobs.
func (s *Scheduler) StopAllWait(timeout time.Duration) {
	s.StopAll()
	jobsNotStopped := atomic.LoadInt64(&s.jobsNotStopped)
	if jobsNotStopped < 1 {
		return
	}

	chanTimeout := time.After(timeout)
	for jobsNotStopped > 0 {
		select {
		case <-chanTimeout:
			return
		case <-s.chanJobsNotStopped:
			jobsNotStopped = atomic.LoadInt64(&s.jobsNotStopped)
		}
	}

	// send a struct in case there is another StopAllWait waiting
	select {
	case s.chanJobsNotStopped <- struct{}{}:
	default:
	}
}
