# Go Cron Job Scheduler

Golang Cron Job Scheduler

[![GoDoc Reference](https://godoc.org/github.com/MichaelS11/go-scheduler?status.svg)](http://godoc.org/github.com/MichaelS11/go-scheduler)
[![Build Status](https://travis-ci.org/MichaelS11/go-scheduler.svg)](https://travis-ci.org/MichaelS11/go-scheduler)
[![Coverage](https://gocover.io/_badge/github.com/MichaelS11/go-scheduler)](https://gocover.io/github.com/MichaelS11/go-scheduler#)
[![Go Report Card](https://goreportcard.com/badge/github.com/MichaelS11/go-scheduler)](https://goreportcard.com/report/github.com/MichaelS11/go-scheduler)

## Get

go get github.com/MichaelS11/go-scheduler


## Simple Example

```go
package main

import (
	"fmt"
	"log"
	"time"

	"github.com/MichaelS11/go-scheduler"
)

func main() {
	myFunction := func(dataInterface interface{}) {
		data := dataInterface.(string)
		fmt.Println(data)
	}

	// Create new scheduler
	s := scheduler.NewScheduler()

	// Make a new job that runs myFunction passing it "myData"
	// A cron of * * * * * * * will run every second.
	err := s.Make("jobName", "* * * * * * *", myFunction, "myData")
	if err != nil {
		log.Fatalln("Make error:", err)
	}

	// Starts the job schedule. Job will run at it's next run time.
	s.Start("jobName")

	// Wait a second to make sure job was run
	time.Sleep(time.Second)

	// Stop all the jobs and waits for then to all stop. In the below case it waits for at least a second.
	// Note this does not kill any running jobs, only stops them from running again.
	s.StopAllWait(time.Second)

	// Output:
	// myData
}
```

## Important note about Cron format

The Cron format is in the form of:

Seconds, Minutes, Hours, Day of month, Month, Day of week, Year

    Field name     Mandatory?   Allowed values    Allowed special characters
    ----------     ----------   --------------    --------------------------
    Seconds        No           0-59              * / , -
    Minutes        Yes          0-59              * / , -
    Hours          Yes          0-23              * / , -
    Day of month   Yes          1-31              * / , - L W
    Month          Yes          1-12 or JAN-DEC   * / , -
    Day of week    Yes          0-6 or SUN-SAT    * / , - L #
    Year           No           1970–2099         * / , -

The Cron parser used is:

https://github.com/gorhill/cronexpr
