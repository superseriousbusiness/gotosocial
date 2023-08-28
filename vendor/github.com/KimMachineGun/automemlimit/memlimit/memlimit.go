package memlimit

import (
	"errors"
	"fmt"
	"io"
	"log"
	"math"
	"os"
	"runtime/debug"
	"strconv"
)

const (
	envGOMEMLIMIT         = "GOMEMLIMIT"
	envAUTOMEMLIMIT       = "AUTOMEMLIMIT"
	envAUTOMEMLIMIT_DEBUG = "AUTOMEMLIMIT_DEBUG"

	defaultAUTOMEMLIMIT = 0.9
)

var (
	// ErrNoLimit is returned when the memory limit is not set.
	ErrNoLimit = errors.New("memory is not limited")
	// ErrNoCgroup is returned when the process is not in cgroup.
	ErrNoCgroup = errors.New("process is not in cgroup")
	// ErrCgroupsNotSupported is returned when the system does not support cgroups.
	ErrCgroupsNotSupported = errors.New("cgroups is not supported on this system")
)

type config struct {
	logger   *log.Logger
	ratio    float64
	env      bool
	provider Provider
}

// Option is a function that configures the behavior of SetGoMemLimitWithOptions.
type Option func(cfg *config)

// WithRatio configures the ratio of the memory limit to set as GOMEMLIMIT.
//
// Default: 0.9
func WithRatio(ratio float64) Option {
	return func(cfg *config) {
		cfg.ratio = ratio
	}
}

// WithEnv configures whether to use environment variables.
//
// Default: false
func WithEnv() Option {
	return func(cfg *config) {
		cfg.env = true
	}
}

// WithProvider configures the provider.
//
// Default: FromCgroup
func WithProvider(provider Provider) Option {
	return func(cfg *config) {
		cfg.provider = provider
	}
}

// SetGoMemLimitWithOpts sets GOMEMLIMIT with options.
//
// Options:
//   - WithRatio
//   - WithEnv (see more SetGoMemLimitWithEnv)
//   - WithProvider
func SetGoMemLimitWithOpts(opts ...Option) (_ int64, _err error) {
	cfg := &config{
		logger:   log.New(io.Discard, "", log.LstdFlags),
		ratio:    defaultAUTOMEMLIMIT,
		env:      false,
		provider: FromCgroup,
	}
	if os.Getenv(envAUTOMEMLIMIT_DEBUG) == "true" {
		cfg.logger = log.Default()
	}
	for _, opt := range opts {
		opt(cfg)
	}
	defer func() {
		if _err != nil {
			cfg.logger.Println(_err)
		}
	}()

	snapshot := debug.SetMemoryLimit(-1)
	defer func() {
		err := recover()
		if err != nil {
			if _err != nil {
				cfg.logger.Println(_err)
			}
			_err = fmt.Errorf("panic during setting the Go's memory limit, rolling back to previous value %d: %v", snapshot, err)
			debug.SetMemoryLimit(snapshot)
		}
	}()

	if val, ok := os.LookupEnv(envGOMEMLIMIT); ok {
		cfg.logger.Printf("GOMEMLIMIT is set already, skipping: %s\n", val)
		return
	}

	ratio := cfg.ratio
	if val, ok := os.LookupEnv(envAUTOMEMLIMIT); ok {
		if val == "off" {
			cfg.logger.Printf("AUTOMEMLIMIT is set to off, skipping\n")
			return
		}
		_ratio, err := strconv.ParseFloat(val, 64)
		if err != nil {
			return 0, fmt.Errorf("cannot parse AUTOMEMLIMIT: %s", val)
		}
		ratio = _ratio
	}
	if ratio <= 0 || ratio > 1 {
		return 0, fmt.Errorf("invalid AUTOMEMLIMIT: %f", ratio)
	}

	limit, err := SetGoMemLimitWithProvider(cfg.provider, ratio)
	if err != nil {
		return 0, fmt.Errorf("failed to set GOMEMLIMIT: %w", err)
	}

	cfg.logger.Printf("GOMEMLIMIT=%d\n", limit)

	return limit, nil
}

// SetGoMemLimitWithEnv sets GOMEMLIMIT with the value from the environment variable.
// You can configure how much memory of the cgroup's memory limit to set as GOMEMLIMIT
// through AUTOMEMLIMIT in the half-open range (0.0,1.0].
//
// If AUTOMEMLIMIT is not set, it defaults to 0.9. (10% is the headroom for memory sources the Go runtime is unaware of.)
// If GOMEMLIMIT is already set or AUTOMEMLIMIT=off, this function does nothing.
func SetGoMemLimitWithEnv() {
	_, _ = SetGoMemLimitWithOpts(WithEnv())
}

// SetGoMemLimit sets GOMEMLIMIT with the value from the cgroup's memory limit and given ratio.
func SetGoMemLimit(ratio float64) (int64, error) {
	return SetGoMemLimitWithProvider(FromCgroup, ratio)
}

// Provider is a function that returns the memory limit.
type Provider func() (uint64, error)

// SetGoMemLimitWithProvider sets GOMEMLIMIT with the value from the given provider and ratio.
func SetGoMemLimitWithProvider(provider Provider, ratio float64) (int64, error) {
	limit, err := provider()
	if err != nil {
		return 0, err
	}
	goMemLimit := cappedFloat2Int(float64(limit) * ratio)
	debug.SetMemoryLimit(goMemLimit)
	return goMemLimit, nil
}

func cappedFloat2Int(f float64) int64 {
	if f > math.MaxInt64 {
		return math.MaxInt64
	}
	return int64(f)
}

// Limit is a helper Provider function that returns the given limit.
func Limit(limit uint64) func() (uint64, error) {
	return func() (uint64, error) {
		return limit, nil
	}
}
