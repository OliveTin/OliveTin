package executor

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

const (
	executionResultSuccess = "success"
	executionResultFailed  = "failed"
	executionResultBlocked = "blocked"
	executionResultTimeout = "timeout"
	executionResultError   = "error"
)

var (
	metricActionsRequested = promauto.NewCounter(prometheus.CounterOpts{
		Name: "olivetin_actions_requested_count",
		Help: "The actions requested count",
	})

	metricActionExecutionsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "olivetin_action_executions_total",
		Help: "Total number of finished action executions grouped by result.",
	}, []string{"result"})

	metricActionExecutionDuration = promauto.NewHistogram(prometheus.HistogramOpts{
		Name:    "olivetin_action_execution_duration_seconds",
		Help:    "Action execution duration in seconds from start to finish.",
		Buckets: []float64{0.1, 0.5, 1, 2, 5, 10, 30, 60, 120, 300, 600},
	})

	executionResultLabels = []string{
		executionResultSuccess,
		executionResultFailed,
		executionResultBlocked,
		executionResultTimeout,
		executionResultError,
	}
)

func init() {
	for _, result := range executionResultLabels {
		metricActionExecutionsTotal.WithLabelValues(result)
	}
}

func executionResultLabel(entry *InternalLogEntry) string {
	if entry.Blocked {
		return executionResultBlocked
	}

	return finishedExecutionResultLabel(entry)
}

func finishedExecutionResultLabel(entry *InternalLogEntry) string {
	if entry.TimedOut {
		return executionResultTimeout
	}

	switch {
	case entry.ExitCode == 0:
		return executionResultSuccess
	case isPreExecutionError(entry):
		return executionResultError
	default:
		return executionResultFailed
	}
}

func isPreExecutionError(entry *InternalLogEntry) bool {
	return entry.ExitCode == DefaultExitCodeNotExecuted || !entry.ExecutionStarted
}

func recordExecutionMetrics(entry *InternalLogEntry) {
	if entry == nil || entry.Queued {
		return
	}

	metricActionExecutionsTotal.WithLabelValues(executionResultLabel(entry)).Inc()
	recordExecutionDuration(entry)
}

func recordExecutionDuration(entry *InternalLogEntry) {
	if entry.DatetimeFinished.IsZero() || entry.DatetimeStarted.IsZero() {
		return
	}

	duration := entry.DatetimeFinished.Sub(entry.DatetimeStarted).Seconds()
	if duration < 0 {
		return
	}

	metricActionExecutionDuration.Observe(duration)
}
