package scheduler_test

import (
	"fmt"
	"log"
	"time"

	"github.com/MichaelS11/go-scheduler"
)

func Example_basic() {

	// This is for testing, to know when myFunction has been called by the scheduler job.
	chanStart := make(chan struct{}, 1)

	// Create a function for the job to call
	myFunction := func(dataInterface interface{}) {
		chanStart <- struct{}{}
		data := dataInterface.(string)
		fmt.Println(data)
	}

	// Create new scheduler
	s := scheduler.NewScheduler()

	// cron is in the form: Seconds, Minutes, Hours, Day of month, Month, Day of week, Year
	// A cron of * * * * * * * will run every second.
	// Parsing cron using:
	// https://github.com/gorhill/cronexpr

	// Make a new job that runs myFunction passing it "myData"
	err := s.Make("jobName", "* * * * * * *", myFunction, "myData")
	if err != nil {
		log.Fatalln("Make error:", err)
	}

	// The follow is normally not needed unless you want the job to run right away.
	// Updating next run time so don't have to wait till the next on the second for the job to run.
	err = s.UpdateNextRun("jobName", time.Now())
	if err != nil {
		log.Fatalln("UpdateNextRun error:", err)
	}

	// Starts the job schedule. Job will run at it's next run time.
	s.Start("jobName")

	// For testing, wait until the job has run before stopping all the jobs
	<-chanStart

	// Stop all the jobs and waits for then to all stop. In the below case it waits for at least a second.
	// Note this does not kill any running jobs, only stops them from running again.
	s.StopAllWait(time.Second)

	// Output:
	// myData
}
