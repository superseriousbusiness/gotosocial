package sched

import (
	"context"
	"sort"
	"sync"
	"sync/atomic"
	"time"

	"codeberg.org/gruf/go-runners"
)

// precision is the maximum time we can offer scheduler run-time precision down to.
const precision = time.Millisecond

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
	jid  atomic.Uint64    // jid is used to iteratively generate unique IDs for jobs
	rgo  func(func())     // goroutine runner, allows using goroutine pool to launch jobs
}

// Start will attempt to start the Scheduler. Immediately returns false if the Service is already running, and true after completed run.
func (sch *Scheduler) Start(gorun func(func())) bool {
	var block sync.Mutex

	// Use mutex to synchronize between started
	// goroutine and ourselves, to ensure that
	// we don't return before Scheduler init'd.
	block.Lock()
	defer block.Unlock()

	ok := sch.svc.GoRun(func(ctx context.Context) {
		// Create Scheduler job channel
		sch.jch = make(chan interface{})

		// Set goroutine runner function
		if sch.rgo = gorun; sch.rgo == nil {
			sch.rgo = func(f func()) { go f() }
		}

		// Unlock start routine
		block.Unlock()

		// Enter main loop
		sch.run(ctx)
	})

	if ok {
		// Wait on goroutine
		block.Lock()
	}

	return ok
}

// Stop will attempt to stop the Scheduler. Immediately returns false if not running, and true only after Scheduler is fully stopped.
func (sch *Scheduler) Stop() bool {
	return sch.svc.Stop()
}

// Running will return whether Scheduler is running (i.e. NOT stopped / stopping).
func (sch *Scheduler) Running() bool {
	return sch.svc.Running()
}

// Done returns a channel that's closed when Scheduler.Stop() is called.
func (sch *Scheduler) Done() <-chan struct{} {
	return sch.svc.Done()
}

// Schedule will add provided Job to the Scheduler, returning a cancel function.
func (sch *Scheduler) Schedule(job *Job) (cancel func()) {
	switch {
	// Check a job was passed
	case job == nil:
		panic("nil job")

	// Check we are running
	case !sch.Running():
		panic("scheduler not running")
	}

	// Calculate next job ID
	last := sch.jid.Load()
	next := sch.jid.Add(1)
	if next < last {
		panic("job id overflow")
	}

	// Pass job to scheduler
	job.id = next
	sch.jch <- job

	// Take ptrs to current state chs
	ctx := sch.svc.Done()
	jch := sch.jch

	// Return cancel function for job ID
	return func() {
		select {
		// Sched stopped
		case <-ctx:

		// Cancel this job
		case jch <- next:
		}
	}
}

// run is the main scheduler run routine, which runs for as long as ctx is valid.
func (sch *Scheduler) run(ctx context.Context) {
	var (
		// now stores the current time, and will only be
		// set when the timer channel is set to be the
		// 'alwaysticks' channel. this allows minimizing
		// the number of calls required to time.Now().
		now time.Time

		// timerset represents whether timer was running
		// for a particular run of the loop. false means
		// that tch == neverticks || tch == alwaysticks.
		timerset bool

		// timer tick channel (or always / never ticks).
		tch <-chan time.Time

		// timer notifies this main routine to wake when
		// the job queued needs to be checked for executions.
		timer *time.Timer

		// stopdrain will stop and drain the timer
		// if it has been running (i.e. timerset == true).
		stopdrain = func() {
			if timerset && !timer.Stop() {
				<-timer.C
			}
		}
	)

	// Create a stopped timer.
	timer = time.NewTimer(1)
	<-timer.C

	for {
		// Reset timer state.
		timerset = false

		if len(sch.jobs) > 0 {
			// Get now time.
			now = time.Now()

			// Sort jobs by next occurring.
			sort.Sort(byNext(sch.jobs))

			// Get next job time.
			next := sch.jobs[0].Next()

			// If this job is _just_ about to be ready, we don't bother
			// sleeping. It's wasted cycles only sleeping for some obscenely
			// tiny amount of time we can't guarantee precision for.
			if until := next.Sub(now); until <= precision/1e3 {
				// This job is behind,
				// set to always tick.
				tch = alwaysticks
			} else {
				// Reset timer to period.
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
		case t := <-tch:
			if !timerset {
				// 'alwaysticks' returns zero
				// times, BUT 'now' will have
				// been set during above sort.
				t = now
			}
			sch.schedule(t)

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

		// Pass to runner
		sch.rgo(func() {
			job.Run(now)
		})

		// Update the next call time
		next := job.timing.Next(now)
		job.next.Store(next)

		if next.IsZero() {
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
