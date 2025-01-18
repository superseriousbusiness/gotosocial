package sched

import (
	"reflect"
	"strconv"
	"strings"
	"sync/atomic"
	"time"
	"unsafe"
)

// Job encapsulates logic for a scheduled job to be run according
// to a set Timing, executing the job with a set panic handler, and
// holding onto a next execution time safely in a concurrent environment.
type Job struct {
	id     uint64
	next   unsafe.Pointer // *time.Time
	timing Timing
	call   func(time.Time)
	panic  func(interface{})
}

// NewJob returns a new Job to run given function.
func NewJob(fn func(now time.Time)) *Job {
	if fn == nil {
		// Ensure a function
		panic("nil func")
	}

	j := &Job{ // set defaults
		timing: emptytiming, // i.e. fire immediately
		call:   fn,
		panic:  func(i interface{}) { panic(i) },
	}

	return j
}

// At sets this Job to execute at time, by passing (*sched.Once)(&at) to .With(). See .With() for details.
func (job *Job) At(at time.Time) *Job {
	return job.With((*Once)(&at))
}

// Every sets this Job to execute every period, by passing sched.Period(period) to .With(). See .With() for details.
func (job *Job) Every(period time.Duration) *Job {
	return job.With(Periodic(period))
}

// EveryAt sets this Job to execute every period starting at time, by passing &PeriodicAt{once: Once(at), period: Periodic(period)} to .With(). See .With() for details.
func (job *Job) EveryAt(at time.Time, period time.Duration) *Job {
	return job.With(&PeriodicAt{Once: Once(at), Period: Periodic(period)})
}

// With sets this Job's timing to given implementation, or if already set will wrap existing using sched.TimingWrap{}.
func (job *Job) With(t Timing) *Job {
	if t == nil {
		// Ensure a timing
		panic("nil Timing")
	}

	if job.id != 0 {
		// Cannot update scheduled job
		panic("job already scheduled")
	}

	if job.timing == emptytiming {
		// Set new timing
		job.timing = t
	} else {
		// Wrap old timing
		old := job.timing
		job.timing = &TimingWrap{
			Outer: t,
			Inner: old,
		}
	}

	return job
}

// OnPanic specifies how this job handles panics, default is an actual panic.
func (job *Job) OnPanic(fn func(interface{})) *Job {
	if fn == nil {
		// Ensure a function
		panic("nil func")
	}

	if job.id != 0 {
		// Cannot update scheduled job
		panic("job already scheduled")
	}

	job.panic = fn
	return job
}

// Next returns the next time this Job is expected to run.
func (job *Job) Next() time.Time {
	return loadTime(&job.next)
}

// Run will execute this Job and pass through given now time.
func (job *Job) Run(now time.Time) {
	defer func() {
		switch r := recover(); {
		case r == nil:
			// no panic
		case job != nil &&
			job.panic != nil:
			job.panic(r)
		default:
			panic(r)
		}
	}()
	job.call(now)
}

// String provides a debuggable string representation of Job including ID, next time and Timing type.
func (job *Job) String() string {
	var buf strings.Builder
	buf.WriteByte('{')
	buf.WriteString("id=")
	buf.WriteString(strconv.FormatUint(job.id, 10))
	buf.WriteByte(' ')
	buf.WriteString("next=")
	buf.WriteString(loadTime(&job.next).Format(time.StampMicro))
	buf.WriteByte(' ')
	buf.WriteString("timing=")
	buf.WriteString(reflect.TypeOf(job.timing).String())
	buf.WriteByte('}')
	return buf.String()
}

func loadTime(p *unsafe.Pointer) time.Time {
	if p := atomic.LoadPointer(p); p != nil {
		return *(*time.Time)(p)
	}
	return zerotime
}

func storeTime(p *unsafe.Pointer, t time.Time) {
	atomic.StorePointer(p, unsafe.Pointer(&t))
}
