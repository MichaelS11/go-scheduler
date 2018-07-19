package scheduler

import (
	"strconv"
	"sync/atomic"
	"testing"
	"time"
)

func TestJobs(t *testing.T) {
	var err error
	s := NewScheduler()

	for i := 0; i < 20; i++ {
		err = s.Make(strconv.FormatInt(int64(i), 10), "1 0 0 1 1 * 2099", testFunction, nil)
		if err != nil {
			t.Fatal("Make error:", err)
		}
	}

	names := s.Jobs()
	if len(names) != 20 {
		t.Fatalf("names - expected: %v - received: %v", 20, len(names))
	}

	for i := 0; i < 20; i++ {
		err = s.Delete(strconv.FormatInt(int64(i), 10))
		if err != nil {
			t.Fatal("Delete error:", err)
		}
	}
}

func TestStopAllWait(t *testing.T) {
	var err error
	var name string
	s := NewScheduler()

	for i := 0; i < 30; i++ {
		name = strconv.FormatInt(int64(i), 10)

		err = s.Make(name, "* * * * * * *", testFunction, nil)
		if err != nil {
			t.Fatal("Make error:", err)
		}

		err = s.UpdateNextRun(name, time.Now())
		if err != nil {
			t.Fatal("UpdateNextRun error:", err)
		}

		err = s.Start(name)
		if err != nil {
			t.Fatal("Start error:", err)
		}
	}

	s.StopAllWait(5 * time.Second)

	jobsNotStopped := atomic.LoadInt64(&s.jobsNotStopped)
	if jobsNotStopped != 0 {
		t.Fatalf("jobsNotStopped - expected: %v - received: %v", 0, jobsNotStopped)
	}

	for i := 0; i < 30; i++ {
		name = strconv.FormatInt(int64(i), 10)

		err = s.UpdateFunction(name, testFunction, uint(2))
		if err != nil {
			t.Fatal("UpdateFunction error:", err)
		}

		err = s.UpdateNextRun(name, time.Now())
		if err != nil {
			t.Fatal("UpdateNextRun error:", err)
		}

		err = s.Start(name)
		if err != nil {
			t.Fatal("Start error:", err)
		}
	}

	s.StopAllWait(5 * time.Second)

	jobsNotStopped = atomic.LoadInt64(&s.jobsNotStopped)
	if jobsNotStopped != 0 {
		t.Fatalf("jobsNotStopped - expected: %v - received: %v", 0, jobsNotStopped)
	}

	for i := 0; i < 30; i++ {
		name = strconv.FormatInt(int64(i), 10)

		err = s.UpdateNextRun(name, time.Now().Add(time.Millisecond))
		if err != nil {
			t.Fatal("UpdateNextRun error:", err)
		}

		err = s.Start(name)
		if err != nil {
			t.Fatal("Start error:", err)
		}

		time.Sleep(time.Millisecond)
	}

	s.StopAllWait(5 * time.Second)

	jobsNotStopped = atomic.LoadInt64(&s.jobsNotStopped)
	if jobsNotStopped != 0 {
		t.Fatalf("jobsNotStopped - expected: %v - received: %v", 0, jobsNotStopped)
	}

	for i := 0; i < 30; i++ {
		name = strconv.FormatInt(int64(i), 10)
		err = s.Delete(name)
		if err != nil {
			t.Fatal("Delete error:", err)
		}
	}

	err = s.Make("a", "* * * * * * *", testFunction, 100)
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

	s.StopAllWait(time.Nanosecond)

	jobsNotStopped = atomic.LoadInt64(&s.jobsNotStopped)
	if jobsNotStopped != 1 {
		t.Fatalf("jobsNotStopped - expected: %v - received: %v", 1, jobsNotStopped)
	}

	err = s.Delete("a")
	if err != nil {
		t.Fatal("Delete error:", err)
	}

	<-chanDone
}
