package memlimit

import (
	"errors"
	"fmt"
	"log/slog"
	"math"
	"os"
	"runtime/debug"
	"strconv"
)

const (
	envGOMEMLIMIT   = "GOMEMLIMIT"
	envAUTOMEMLIMIT = "AUTOMEMLIMIT"
	// Deprecated: use memlimit.WithLogger instead
	envAUTOMEMLIMIT_DEBUG = "AUTOMEMLIMIT_DEBUG"

	defaultAUTOMEMLIMIT = 0.9
)

var (
	// ErrNoLimit is returned when the memory limit is not set.
	ErrNoLimit = errors.New("memory is not limited")
)

type config struct {
	logger   *slog.Logger
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

// WithProvider configures the provider.
//
// Default: FromCgroup
func WithProvider(provider Provider) Option {
	return func(cfg *config) {
		cfg.provider = provider
	}
}

// WithLogger configures the logger.
// It automatically attaches the "package" attribute to the logs.
//
// Default: slog.New(noopLogger{})
func WithLogger(logger *slog.Logger) Option {
	return func(cfg *config) {
		cfg.logger = memlimitLogger(logger)
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

func memlimitLogger(logger *slog.Logger) *slog.Logger {
	if logger == nil {
		return slog.New(noopLogger{})
	}
	return logger.With(slog.String("package", "github.com/KimMachineGun/automemlimit/memlimit"))
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
//   - WithLogger
func SetGoMemLimitWithOpts(opts ...Option) (_ int64, _err error) {
	// init config
	cfg := &config{
		logger:   slog.New(noopLogger{}),
		ratio:    defaultAUTOMEMLIMIT,
		provider: FromCgroup,
	}
	// TODO: remove this
	if debug, ok := os.LookupEnv(envAUTOMEMLIMIT_DEBUG); ok {
		defaultLogger := memlimitLogger(slog.Default())
		defaultLogger.Warn("AUTOMEMLIMIT_DEBUG is deprecated, use memlimit.WithLogger instead")
		if debug == "true" {
			cfg.logger = defaultLogger
		}
	}
	for _, opt := range opts {
		opt(cfg)
	}

	// log error if any on return
	defer func() {
		if _err != nil {
			cfg.logger.Error("failed to set GOMEMLIMIT", slog.Any("error", _err))
		}
	}()

	// parse experiments
	exps, err := parseExperiments()
	if err != nil {
		return 0, fmt.Errorf("failed to parse experiments: %w", err)
	}
	if exps.System {
		cfg.logger.Info("system experiment is enabled: using system memory limit as a fallback")
		cfg.provider = ApplyFallback(cfg.provider, FromSystem)
	}

	// capture the current GOMEMLIMIT for rollback in case of panic
	snapshot := debug.SetMemoryLimit(-1)
	defer func() {
		panicErr := recover()
		if panicErr != nil {
			if _err != nil {
				cfg.logger.Error("failed to set GOMEMLIMIT", slog.Any("error", _err))
			}
			_err = fmt.Errorf("panic during setting the Go's memory limit, rolling back to previous limit %d: %v",
				snapshot, panicErr,
			)
			debug.SetMemoryLimit(snapshot)
		}
	}()

	// check if GOMEMLIMIT is already set
	if val, ok := os.LookupEnv(envGOMEMLIMIT); ok {
		cfg.logger.Info("GOMEMLIMIT is already set, skipping", slog.String(envGOMEMLIMIT, val))
		return 0, nil
	}

	// parse AUTOMEMLIMIT
	ratio := cfg.ratio
	if val, ok := os.LookupEnv(envAUTOMEMLIMIT); ok {
		if val == "off" {
			cfg.logger.Info("AUTOMEMLIMIT is set to off, skipping")
			return 0, nil
		}
		_ratio, err := strconv.ParseFloat(val, 64)
		if err != nil {
			return 0, fmt.Errorf("cannot parse AUTOMEMLIMIT: %s", val)
		}
		ratio = _ratio
	}

	// set GOMEMLIMIT
	limit, err := setGoMemLimit(ApplyRatio(cfg.provider, ratio))
	if err != nil {
		if errors.Is(err, ErrNoLimit) {
			cfg.logger.Info("memory is not limited, skipping")
			return 0, nil
		}
		return 0, fmt.Errorf("failed to set GOMEMLIMIT: %w", err)
	}

	cfg.logger.Info("GOMEMLIMIT is updated", slog.Int64(envGOMEMLIMIT, limit))

	return limit, nil
}

// SetGoMemLimitWithEnv sets GOMEMLIMIT with the value from the environment variables.
// Since WithEnv is deprecated, this function is equivalent to SetGoMemLimitWithOpts().
// Deprecated: use SetGoMemLimitWithOpts instead.
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
