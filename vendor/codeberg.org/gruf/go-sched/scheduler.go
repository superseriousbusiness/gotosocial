package sched

import (
	"context"
	"sort"
	"time"

	"codeberg.org/gruf/go-atomics"
	"codeberg.org/gruf/go-runners"
)

var (
	// neverticks is a timer channel that never ticks (it's starved).
	neverticks = make(chan time.Time)

	// alwaysticks is a timer channel that always ticks (it's closed).
	alwaysticks = func() chan time.Time {
		ch := make(chan time.Time)
		close(ch)
		return ch
	}()
)

// Scheduler provides a means of running jobs at specific times and
// regular intervals, all while sharing a single underlying timer.
type Scheduler struct {
	jobs []*Job           // jobs is a list of tracked Jobs to be executed
	jch  chan interface{} // jch accepts either Jobs or job IDs to notify new/removed jobs
	svc  runners.Service  // svc manages the main scheduler routine
	jid  atomics.Uint64   // jid is used to iteratively generate unique IDs for jobs
}

// New returns a new Scheduler instance with given job change queue size.
func NewScheduler(queue int) Scheduler {
	if queue < 0 {
		queue = 10
	}
	return Scheduler{jch: make(chan interface{}, queue)}
}

// Start will attempt to start the Scheduler. Immediately returns false if the Service is already running, and true after completed run.
func (sch *Scheduler) Start() bool {
	return sch.svc.Run(sch.run)
}

// Stop will attempt to stop the Scheduler. Immediately returns false if not running, and true only after Scheduler is fully stopped.
func (sch *Scheduler) Stop() bool {
	return sch.svc.Stop()
}

// Running will return whether Scheduler is running.
func (sch *Scheduler) Running() bool {
	return sch.svc.Running()
}

// Schedule will add provided Job to the Scheduler, returning a cancel function.
func (sch *Scheduler) Schedule(job *Job) (cancel func()) {
	if job == nil {
		// Ensure there's a job!
		panic("nil job")
	}

	// Get last known job ID
	last := sch.jid.Load()

	// Give this job an ID and check overflow
	if job.id = sch.jid.Add(1); job.id < last {
		panic("scheduler job id overflow")
	}

	// Pass job to scheduler
	sch.jch <- job

	// Return cancel function for job ID
	return func() { sch.jch <- job.id }
}

// run is the main scheduler run routine, which runs for as long as ctx is valid.
func (sch *Scheduler) run(ctx context.Context) {
	var (
		// timerset represents whether timer was running
		// for a particular run of the loop. false means
		// that tch == neverticks || tch == alwaysticks
		timerset bool

		// timer tick channel (or a never-tick channel)
		tch <-chan time.Time

		// timer notifies this main routine to wake when
		// the job queued needs to be checked for executions
		timer *time.Timer

		// stopdrain will stop and drain the timer
		// if it has been running (i.e. timerset == true)
		stopdrain = func() {
			if timerset && !timer.Stop() {
				<-timer.C
			}
		}
	)

	for {
		select {
		// Handle received job/id
		case v := <-sch.jch:
			sch.handle(v)
			continue

		// No more
		default:
		}

		// Done
		break
	}

	// Create a stopped timer
	timer = time.NewTimer(1)
	<-timer.C

	for {
		// Reset timer state
		timerset = false

		if len(sch.jobs) > 0 {
			// Sort jobs by next occurring
			sort.Sort(byNext(sch.jobs))

			// Get execution time
			now := time.Now()

			// Get next job time
			next := sch.jobs[0].Next()

			if until := next.Sub(now); until <= 0 {
				// This job is behind schedule,
				// set timer to always tick
				tch = alwaysticks
			} else {
				// Reset timer to period
				timer.Reset(until)
				tch = timer.C
				timerset = true
			}
		} else {
			// Unset timer
			tch = neverticks
		}

		select {
		// Scheduler stopped
		case <-ctx.Done():
			stopdrain()
			return

		// Timer ticked, run scheduled
		case now := <-tch:
			sch.schedule(now)

		// Received update, handle job/id
		case v := <-sch.jch:
			sch.handle(v)
			stopdrain()
		}
	}
}

// handle takes an interfaces received from Scheduler.jch and handles either:
// - Job --> new job to add.
// - uint64 --> job ID to remove.
func (sch *Scheduler) handle(v interface{}) {
	switch v := v.(type) {
	// New job added
	case *Job:
		// Get current time
		now := time.Now()

		// Update the next call time
		next := v.timing.Next(now)
		v.next.Store(next)

		// Append this job to queued
		sch.jobs = append(sch.jobs, v)

	// Job removed
	case uint64:
		for i := 0; i < len(sch.jobs); i++ {
			if sch.jobs[i].id == v {
				// This is the job we're looking for! Drop this
				sch.jobs = append(sch.jobs[:i], sch.jobs[i+1:]...)
				return
			}
		}
	}
}

// schedule will iterate through the scheduler jobs and execute those necessary, updating their next call time.
func (sch *Scheduler) schedule(now time.Time) {
	for i := 0; i < len(sch.jobs); {
		// Scope our own var
		job := sch.jobs[i]

		// We know these jobs are ordered by .Next(), so as soon
		// as we reach one with .Next() after now, we can return
		if job.Next().After(now) {
			return
		}

		// Update the next call time
		next := job.timing.Next(now)
		job.next.Store(next)

		// Run this job async!
		go job.Run(now)

		if job.Next().IsZero() {
			// Zero time, this job is done and can be dropped
			sch.jobs = append(sch.jobs[:i], sch.jobs[i+1:]...)
			continue
		}

		// Iter
		i++
	}
}

// byNext is an implementation of sort.Interface to sort Jobs by their .Next() time.
type byNext []*Job

func (by byNext) Len() int {
	return len(by)
}

func (by byNext) Less(i int, j int) bool {
	return by[i].Next().Before(by[j].Next())
}

func (by byNext) Swap(i int, j int) {
	by[i], by[j] = by[j], by[i]
}
