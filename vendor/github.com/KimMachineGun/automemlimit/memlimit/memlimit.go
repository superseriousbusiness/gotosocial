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
)

type config struct {
	logger   *log.Logger
	ratio    float64
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
//
// Deprecated: currently this does nothing.
func WithEnv() Option {
	return func(cfg *config) {}
}

// WithProvider configures the provider.
//
// Default: FromCgroup
func WithProvider(provider Provider) Option {
	return func(cfg *config) {
		cfg.provider = provider
	}
}

// SetGoMemLimitWithOpts sets GOMEMLIMIT with options and environment variables.
//
// You can configure how much memory of the cgroup's memory limit to set as GOMEMLIMIT
// through AUTOMEMLIMIT envrironment variable in the half-open range (0.0,1.0].
//
// If AUTOMEMLIMIT is not set, it defaults to 0.9. (10% is the headroom for memory sources the Go runtime is unaware of.)
// If GOMEMLIMIT is already set or AUTOMEMLIMIT=off, this function does nothing.
//
// If AUTOMEMLIMIT_EXPERIMENT is set, it enables experimental features.
// Please see the documentation of Experiments for more details.
//
// Options:
//   - WithRatio
//   - WithProvider
func SetGoMemLimitWithOpts(opts ...Option) (_ int64, _err error) {
	cfg := &config{
		logger:   log.New(io.Discard, "", log.LstdFlags),
		ratio:    defaultAUTOMEMLIMIT,
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

	exps, err := parseExperiments()
	if err != nil {
		return 0, fmt.Errorf("failed to parse experiments: %w", err)
	}
	if exps.System {
		cfg.logger.Println("system experiment is enabled: using system memory limit as a fallback")
		cfg.provider = ApplyFallback(cfg.provider, FromSystem)
	}

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
		return 0, nil
	}

	ratio := cfg.ratio
	if val, ok := os.LookupEnv(envAUTOMEMLIMIT); ok {
		if val == "off" {
			cfg.logger.Printf("AUTOMEMLIMIT is set to off, skipping\n")
			return 0, nil
		}
		_ratio, err := strconv.ParseFloat(val, 64)
		if err != nil {
			return 0, fmt.Errorf("cannot parse AUTOMEMLIMIT: %s", val)
		}
		ratio = _ratio
	}

	limit, err := setGoMemLimit(ApplyRatio(cfg.provider, ratio))
	if err != nil {
		if errors.Is(err, ErrNoLimit) {
			cfg.logger.Printf("memory is not limited, skipping: %v\n", err)
			return 0, nil
		}
		return 0, fmt.Errorf("failed to set GOMEMLIMIT: %w", err)
	}

	cfg.logger.Printf("GOMEMLIMIT=%d\n", limit)

	return limit, nil
}

func SetGoMemLimitWithEnv() {
	_, _ = SetGoMemLimitWithOpts()
}

// SetGoMemLimit sets GOMEMLIMIT with the value from the cgroup's memory limit and given ratio.
func SetGoMemLimit(ratio float64) (int64, error) {
	return SetGoMemLimitWithOpts(WithRatio(ratio))
}

// SetGoMemLimitWithProvider sets GOMEMLIMIT with the value from the given provider and ratio.
func SetGoMemLimitWithProvider(provider Provider, ratio float64) (int64, error) {
	return SetGoMemLimitWithOpts(WithProvider(provider), WithRatio(ratio))
}

func setGoMemLimit(provider Provider) (int64, error) {
	limit, err := provider()
	if err != nil {
		return 0, err
	}
	capped := cappedU64ToI64(limit)
	debug.SetMemoryLimit(capped)
	return capped, nil
}

func cappedU64ToI64(limit uint64) int64 {
	if limit > math.MaxInt64 {
		return math.MaxInt64
	}
	return int64(limit)
}
