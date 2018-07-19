package scheduler

import (
	"fmt"
	"testing"
	"time"
)

var (
	chanDone     = make(chan struct{}, 2)
	chanStart    = make(chan struct{}, 2)
	testFunction = func(dataInterface interface{}) {
		switch data := (dataInterface).(type) {
		case nil:
		case *int:
			*(data)++
			chanDone <- struct{}{}
		case int:
			chanStart <- struct{}{}
			time.Sleep(time.Duration(data) * time.Millisecond)
			chanDone <- struct{}{}
		case uint:
			time.Sleep(time.Duration(data) * time.Millisecond)
		default:
			fmt.Printf("testFunction dataInterface type: %T\n", dataInterface)
		}
	}
)

func TestJobNotFound(t *testing.T) {
	s := NewScheduler()

	err := s.Start("a")
	if err != ErrJobNotFound {
		t.Fatalf("Start - expected: %v - received: %v", ErrJobNotFound, err)
	}

	err = s.Stop("a")
	if err != ErrJobNotFound {
		t.Fatalf("Stop - expected: %v - received: %v", ErrJobNotFound, err)
	}

	err = s.Delete("a")
	if err != ErrJobNotFound {
		t.Fatalf("Delete - expected: %v - received: %v", ErrJobNotFound, err)
	}

	err = s.UpdateCron("a", "1 0 0 1 1 * 2099")
	if err != ErrJobNotFound {
		t.Fatalf("UpdateCron - expected: %v - received: %v", ErrJobNotFound, err)
	}

	err = s.UpdateNextRun("a", time.Time{})
	if err != ErrJobNotFound {
		t.Fatalf("UpdateNextRun - expected: %v - received: %v", ErrJobNotFound, err)
	}

	err = s.UpdateFunction("a", testFunction, nil)
	if err != ErrJobNotFound {
		t.Fatalf("UpdateFunction - expected: %v - received: %v", ErrJobNotFound, err)
	}

	state, err := s.GetState("a")
	if err != ErrJobNotFound {
		t.Fatalf("GetState - expected: %v - received: %v", ErrJobNotFound, err)
	}
	if state != 0 {
		t.Fatalf("state - expected: %v - received: %v", 0, state)
	}

	data, err := s.GetData("a")
	if err != ErrJobNotFound {
		t.Fatalf("GetData - expected: %v - received: %v", ErrJobNotFound, err)
	}
	if data != nil {
		t.Fatalf("GetData - expected: %v - received: %v", nil, data)
	}
}

func TestJobBasic(t *testing.T) {
	s := NewScheduler()

	err := s.Make("a", "1 0 0 1 1 * 1", testFunction, nil)
	expected := "cron parse error: syntax error in year field: '1'"
	if err == nil || err.Error() != expected {
		t.Fatalf("Make - expected: %v - received: %v", expected, err)
	}

	err = s.Make("a", "1 0 0 1 1 * 2099", testFunction, nil)
	if err != nil {
		t.Fatal("Make error:", err)
	}

	err = s.Make("a", "1 0 0 1 1 * 2099", testFunction, nil)
	if err != ErrJobAlreadyExists {
		t.Fatalf("Make - expected: %v - received: %v", ErrJobAlreadyExists, err)
	}

	err = s.UpdateCron("a", "1 0 0 1 1 * 1")
	if err == nil || err.Error() != expected {
		t.Fatalf("UpdateCron - expected: %v - received: %v", expected, err)
	}

	err = s.UpdateCron("a", "1 0 0 1 1 * 2099")
	if err != nil {
		t.Fatal("UpdateCron error:", err)
	}

	err = s.UpdateNextRun("a", time.Now().Add(9999*time.Hour))
	if err != nil {
		t.Fatal("UpdateNextRun error:", err)
	}

	err = s.UpdateFunction("a", testFunction, nil)
	if err != nil {
		t.Fatal("UpdateFunction error:", err)
	}

	err = s.Start("a")
	if err != nil {
		t.Fatal("Start error:", err)
	}

	err = s.Start("a")
	if err != ErrJobMustBeStopped {
		t.Fatalf("Start - expected: %v - received: %v", ErrJobMustBeStopped, err)
	}

	err = s.UpdateNextRun("a", time.Now().Add(9999*time.Hour))
	if err != nil {
		t.Fatal("UpdateNextRun error:", err)
	}

	state, err := s.GetState("a")
	if err != nil {
		t.Fatal("GetState error:", err)
	}
	if state != StateScheduled {
		t.Fatalf("state - expected: %v - received: %v", StateScheduled, state)
	}

	err = s.Stop("a")
	if err != nil {
		t.Fatal("Stop error:", err)
	}

	state, err = s.GetState("a")
	if err != nil {
		t.Fatal("GetState error:", err)
	}
	if state != StateStopped {
		t.Fatalf("state - expected: %v - received: %v", StateStopped, state)
	}

	data, err := s.GetData("a")
	if err != nil {
		t.Fatal("GetData error:", err)
	}
	if data != nil {
		t.Fatalf("state - expected: %v - received: %v", nil, data)
	}

	err = s.Delete("a")
	if err != nil {
		t.Fatal("jobDelete error:", err)
	}

	err = s.Delete("a")
	if err != ErrJobNotFound {
		t.Fatalf("Delete - expected: %v - received: %v", ErrJobNotFound, err)
	}
}

func TestJobRun(t *testing.T) {
	s := NewScheduler()

	jobData := 1
	err := s.Make("a", "* * * * * * *", testFunction, &jobData)
	if err != nil {
		t.Fatal("Make error:", err)
	}

	err = s.UpdateNextRun("a", time.Now())
	if err != nil {
		t.Fatal("UpdateNextRun error:", err)
	}

	err = s.Start("a")
	if err != nil {
		t.Fatal("Start error:", err)
	}

	<-chanDone

	err = s.Stop("a")
	if err != nil {
		t.Fatal("Stop error:", err)
	}

	dataInterface, err := s.GetData("a")
	if err != nil {
		t.Fatal("GetData error:", err)
	}
	if jobData != 2 {
		t.Fatalf("jobData - expected: %v - received: %v", 2, jobData)
	}
	jobData = *(dataInterface.(*int))
	if jobData != 2 {
		t.Fatalf("jobData - expected: %v - received: %v", 2, jobData)
	}

	jobData = 1

	err = s.UpdateNextRun("a", time.Now())
	if err != nil {
		t.Fatal("UpdateNextRun error:", err)
	}

	err = s.Start("a")
	if err != nil {
		t.Fatal("Start error:", err)
	}

	err = s.UpdateNextRun("a", time.Now())
	if err != nil && err != ErrJobIsRunning {
		t.Fatal("UpdateNextRun error:", err)
	}

	<-chanDone

	err = s.UpdateNextRun("a", time.Now())
	if err != nil && err != ErrJobIsRunning {
		t.Fatal("UpdateNextRun error:", err)
	}

	<-chanDone

	dataInterface, err = s.GetData("a")
	if err != nil {
		t.Fatal("GetData error:", err)
	}
	if jobData != 3 {
		t.Fatalf("jobData - expected: %v - received: %v", 3, jobData)
	}
	jobData = *(dataInterface.(*int))
	if jobData != 3 {
		t.Fatalf("jobData - expected: %v - received: %v", 3, jobData)
	}

	err = s.Delete("a")
	if err != nil {
		t.Fatal("jobDelete error:", err)
	}
}

func TestJobStop(t *testing.T) {
	s := NewScheduler()

	err := s.Make("a", "* * * * * * *", testFunction, 200)
	if err != nil {
		t.Fatal("Make error:", err)
	}

	err = s.UpdateNextRun("a", time.Now())
	if err != nil {
		t.Fatal("UpdateNextRun error:", err)
	}

	err = s.Start("a")
	if err != nil {
		t.Fatal("Start error:", err)
	}

	<-chanStart

	err = s.UpdateNextRun("a", time.Now())
	if err != nil && err != ErrJobIsRunning {
		t.Fatal("UpdateNextRun error:", err)
	}

	err = s.Stop("a")
	if err != nil {
		t.Fatal("Stop error:", err)
	}

	for i := 0; ; i++ {
		state, err := s.GetState("a")
		if err != nil {
			t.Fatal("GetState error:", err)
		}
		if state == StateStopped {
			break
		}
		if state != StateRunning|StateStopping {
			t.Fatalf("state - expected: %v - received: %v", StateRunning|StateStopping, state)
		}
		if i > 25 {
			t.Fatal("timeout")
		}
		time.Sleep(100 * time.Millisecond)
	}

	<-chanDone

	err = s.Delete("a")
	if err != nil {
		t.Fatal("jobDelete error:", err)
	}
}

func TestJobDelete(t *testing.T) {
	s := NewScheduler()

	err := s.Make("a", "* * * * * * *", testFunction, 200)
	if err != nil {
		t.Fatal("Make error:", err)
	}

	err = s.UpdateNextRun("a", time.Now())
	if err != nil {
		t.Fatal("UpdateNextRun error:", err)
	}

	err = s.Start("a")
	if err != nil {
		t.Fatal("Start error:", err)
	}

	err = s.UpdateNextRun("a", time.Now())
	if err != nil && err != ErrJobIsRunning {
		t.Fatal("UpdateNextRun error:", err)
	}

	<-chanStart

	err = s.UpdateNextRun("a", time.Now())
	if err != nil && err != ErrJobIsRunning {
		t.Fatal("UpdateNextRun error:", err)
	}

	err = s.Delete("a")
	if err != nil {
		t.Fatal("Delete error:", err)
	}

	for i := 0; ; i++ {
		state, err := s.GetState("a")
		if err == ErrJobNotFound {
			if state != 0 {
				t.Fatalf("state - expected: %v - received: %v", 0, state)
			}
			break
		} else if err != nil {
			t.Fatal("GetState error:", err)
		}
		if state != StateRunning|StateStopping|StateDeleting {
			t.Fatalf("state - expected: %v - received: %v", StateRunning|StateStopping|StateDeleting, state)
		}
		if i > 25 {
			t.Fatal("timeout")
		}
		time.Sleep(100 * time.Millisecond)
	}

	<-chanDone
}
