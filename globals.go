package scheduler

import (
	"errors"
	"sync"
	"time"

	"github.com/gorhill/cronexpr"
)

// State what state the job is in
type State int

const (
	// StateScheduled when job pending schedule to run
	StateScheduled State = 1 << iota
	// StateRunning when job is running
	StateRunning
	// StateStopping when job is scheduled to stop after run
	StateStopping
	// StateStopped when job is stopped
	StateStopped
	// StateDeleting when job is scheduled to be deleted after run
	StateDeleting
)

var (
	// ErrJobNotFound is returned when job has not been found. Make sure to make a job first.
	ErrJobNotFound = errors.New("job not found")
	// ErrJobAlreadyExists is returned when a job name already exists. Make sure to delete the job with the same name first.
	ErrJobAlreadyExists = errors.New("job already exists")
	// ErrJobMustBeStopped is returned when a job is not stopped first.
	ErrJobMustBeStopped = errors.New("job must be stopped")
	// ErrJobIsRunning is returned when a job is running
	ErrJobIsRunning = errors.New("job is running")
)

// Scheduler is used to create and run jobs.
// Must use NewScheduler to create a new one.
type Scheduler struct {
	jobs               map[string]*jobStruct
	jobsRWMutex        *sync.RWMutex
	jobsNotStopped     int64
	chanJobsNotStopped chan struct{}
}

type jobStruct struct {
	name           string
	cronExpression *cronexpr.Expression
	function       func(interface{})
	data           interface{}
	mutex          *sync.Mutex
	state          State
	nextRun        time.Time
	timer          *time.Timer
}
