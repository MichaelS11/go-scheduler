package scheduler

import (
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gorhill/cronexpr"
)

// Make creates a new job.
// Will error if job with same name is already created.
func (s *Scheduler) Make(name string, cron string, function func(interface{}), data interface{}) error {
	var err error

	job := jobStruct{
		name:     name,
		function: function,
		data:     data,
		mutex:    &sync.Mutex{},
		state:    StateStopped,
	}

	job.cronExpression, err = cronexpr.Parse(cron)
	if err != nil {
		return fmt.Errorf("cron parse error: %v", err)
	}
	job.nextRun = job.cronExpression.Next(time.Now().UTC())

	s.jobsRWMutex.Lock()
	defer s.jobsRWMutex.Unlock()

	_, ok := s.jobs[name]
	if ok {
		return ErrJobAlreadyExists
	}
	s.jobs[name] = &job

	return nil
}

// Start starts the job run schedule. Job will run at next run time.
// Job must be created and stopped to start the job run schedule.
func (s *Scheduler) Start(name string) error {
	s.jobsRWMutex.RLock()
	job, ok := s.jobs[name]
	s.jobsRWMutex.RUnlock()
	if !ok {
		return ErrJobNotFound
	}

	job.mutex.Lock()
	defer job.mutex.Unlock()

	if job.state != StateStopped {
		return ErrJobMustBeStopped
	}

	job.state = StateScheduled
	atomic.AddInt64(&s.jobsNotStopped, 1)
	job.timer = time.AfterFunc(job.nextRun.Sub(time.Now().UTC()), func() { s.run(job) })

	return nil
}

// Stop stops the job run schedule.
// If the job is not running, it will not run again untill job is started.
// If the job is running, the job will finish running then will not run again until the job is started.
// Will not error if job is stopped more than once.
func (s *Scheduler) Stop(name string) error {
	s.jobsRWMutex.RLock()
	job, ok := s.jobs[name]
	s.jobsRWMutex.RUnlock()
	if !ok {
		return ErrJobNotFound
	}

	job.mutex.Lock()
	s.stop(job)
	job.mutex.Unlock()

	return nil
}

// stop stops the job if can or sets flag to stop on next run
func (s *Scheduler) stop(job *jobStruct) {
	// assumes you already have the job mutex lock

	if job.state == StateStopped {
		return
	}

	if job.state&StateRunning > 0 {
		job.state |= StateStopping
		return
	}

	if job.timer.Stop() {
		job.timer = nil
		job.state = StateStopped
		atomic.AddInt64(&s.jobsNotStopped, -1)
		select {
		case s.chanJobsNotStopped <- struct{}{}:
		default:
		}
		return
	}

	//  timer has kicked off to run goroutine but run does not have job mutex lock
	job.state |= StateStopping
}

// Delete stops the job and deletes the job.
// If the job is not running, the job is deleted.
// If the job is running, the job will finish running then be deleted.
func (s *Scheduler) Delete(name string) error {
	s.jobsRWMutex.RLock()
	job, ok := s.jobs[name]
	s.jobsRWMutex.RUnlock()
	if !ok {
		return ErrJobNotFound
	}

	job.mutex.Lock()
	s.stop(job)
	s.jobDelete(job)
	job.mutex.Unlock()

	return nil
}

// jobDelete deletes the job if can or sets flag to delete on run
func (s *Scheduler) jobDelete(job *jobStruct) {
	// assumes you already have the job mutex lock

	if job.state != StateStopped {
		job.state |= StateDeleting
		return
	}

	s.jobsRWMutex.Lock()
	delete(s.jobs, job.name)
	s.jobsRWMutex.Unlock()
}

// UpdateCron updates the job's cron shedule
func (s *Scheduler) UpdateCron(name string, cron string) error {
	s.jobsRWMutex.RLock()
	job, ok := s.jobs[name]
	s.jobsRWMutex.RUnlock()
	if !ok {
		return ErrJobNotFound
	}

	cronExpression, err := cronexpr.Parse(cron)
	if err != nil {
		return fmt.Errorf("cron parse error: %v", err)
	}

	job.mutex.Lock()
	job.cronExpression = cronExpression
	job.mutex.Unlock()

	return nil
}

// UpdateNextRun updates the job's next run time.
// This is best used when the job is stopped, then it just updates the next run time.
// If the job is running or the next run cannot be stopped, this will return error ErrJobIsRunning so the job does not possibility run twice
func (s *Scheduler) UpdateNextRun(name string, nextRun time.Time) error {
	s.jobsRWMutex.RLock()
	job, ok := s.jobs[name]
	s.jobsRWMutex.RUnlock()
	if !ok {
		return ErrJobNotFound
	}

	job.mutex.Lock()
	defer job.mutex.Unlock()

	if job.state == StateStopped {
		job.nextRun = nextRun
		return nil
	}

	if job.state&StateRunning > 0 {
		return ErrJobIsRunning
	}

	if job.timer.Stop() {
		job.nextRun = nextRun
		job.timer = time.AfterFunc(job.nextRun.Sub(time.Now().UTC()), func() { s.run(job) })
		return nil
	}

	//  timer has kicked off to run goroutine but run does not have job mutex lock
	return ErrJobIsRunning
}

// UpdateFunction updates the job's function and data
func (s *Scheduler) UpdateFunction(name string, function func(interface{}), data interface{}) error {
	s.jobsRWMutex.RLock()
	job, ok := s.jobs[name]
	s.jobsRWMutex.RUnlock()
	if !ok {
		return ErrJobNotFound
	}

	job.mutex.Lock()
	job.function = function
	job.data = data
	job.mutex.Unlock()

	return nil
}

// GetState returns job state
func (s *Scheduler) GetState(name string) (State, error) {
	s.jobsRWMutex.RLock()
	job, ok := s.jobs[name]
	s.jobsRWMutex.RUnlock()
	if !ok {
		return 0, ErrJobNotFound
	}

	return job.state, nil
}

// GetData returns job data
func (s *Scheduler) GetData(name string) (interface{}, error) {
	s.jobsRWMutex.RLock()
	job, ok := s.jobs[name]
	s.jobsRWMutex.RUnlock()
	if !ok {
		return nil, ErrJobNotFound
	}

	return job.data, nil
}

// run runs the job
func (s *Scheduler) run(job *jobStruct) {
	job.mutex.Lock()
	job.timer = nil
	if s.doStoppingOrDeleting(job) {
		job.mutex.Unlock()
		return
	}
	job.state = StateRunning
	job.nextRun = job.cronExpression.Next(time.Now().UTC())
	job.mutex.Unlock()

	job.function(job.data)

	job.mutex.Lock()
	defer job.mutex.Unlock()
	if s.doStoppingOrDeleting(job) {
		return
	}

	job.state = StateScheduled
	job.timer = time.AfterFunc(job.nextRun.Sub(time.Now().UTC()), func() { s.run(job) })
}

// doStoppingOrDeleting return true if stopping or deleting
func (s *Scheduler) doStoppingOrDeleting(job *jobStruct) bool {
	// assumes you already have the job mutex lock

	if job.state&StateDeleting > 0 {
		job.state = StateStopped
		atomic.AddInt64(&s.jobsNotStopped, -1)
		select {
		case s.chanJobsNotStopped <- struct{}{}:
		default:
		}
		s.jobDelete(job)
		return true
	}
	if job.state&StateStopping > 0 {
		job.state = StateStopped
		atomic.AddInt64(&s.jobsNotStopped, -1)
		select {
		case s.chanJobsNotStopped <- struct{}{}:
		default:
		}
		return true
	}

	return false
}
