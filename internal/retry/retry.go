package retry

import (
	"context"
	"fmt"
	"math"
	"time"
)

// Config 重试配置
type Config struct {
	Enabled         bool
	MaxRetries      int
	InitialDelay    time.Duration
	MaxDelay        time.Duration
	ExponentialBase float64
}

// DefaultConfig 默认重试配置
func DefaultConfig() *Config {
	return &Config{
		Enabled:         true,
		MaxRetries:      3,
		InitialDelay:    time.Second,
		MaxDelay:        60 * time.Second,
		ExponentialBase: 2.0,
	}
}

// ExhaustedError 重试耗尽错误
type ExhaustedError struct {
	LastError error
	Attempts  int
}

func (e *ExhaustedError) Error() string {
	return fmt.Sprintf("retry failed after %d attempts: %v", e.Attempts, e.LastError)
}

// OnRetryFunc 重试回调函数类型
type OnRetryFunc func(err error, attempt int)

// CalculateDelay 计算延迟时间（指数退避）
func (c *Config) CalculateDelay(attempt int) time.Duration {
	delay := float64(c.InitialDelay) * math.Pow(c.ExponentialBase, float64(attempt))
	if delay > float64(c.MaxDelay) {
		delay = float64(c.MaxDelay)
	}
	return time.Duration(delay)
}

// Do 执行带重试的函数
func Do[T any](ctx context.Context, cfg *Config, fn func() (T, error), onRetry OnRetryFunc) (T, error) {
	var zero T
	var lastErr error

	if cfg == nil {
		cfg = DefaultConfig()
	}

	if !cfg.Enabled {
		return fn()
	}

	for attempt := 0; attempt <= cfg.MaxRetries; attempt++ {
		result, err := fn()
		if err == nil {
			return result, nil
		}

		lastErr = err

		if attempt >= cfg.MaxRetries {
			return zero, &ExhaustedError{LastError: lastErr, Attempts: attempt + 1}
		}

		delay := cfg.CalculateDelay(attempt)

		if onRetry != nil {
			onRetry(err, attempt+1)
		}

		select {
		case <-ctx.Done():
			return zero, ctx.Err()
		case <-time.After(delay):
		}
	}

	return zero, lastErr
}
