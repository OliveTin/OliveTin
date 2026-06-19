package logfilter

// Record exposes only log fields that may be used in filter expressions.
type Record struct {
	Status   string
	Action   string
	User     string
	Tags     []string
	Blocked  bool
	TimedOut bool
	Running  bool
	ExitCode int32
	Output   string
}

// StatusLabel matches the status text shown in the web UI.
func StatusLabel(executionFinished, blocked, timedOut, queued bool) string {
	if !executionFinished {
		if queued {
			return "Queued"
		}
		return "Running"
	}
	return finishedStatusLabel(blocked, timedOut)
}

func finishedStatusLabel(blocked, timedOut bool) string {
	if blocked {
		return "Blocked"
	}
	if timedOut {
		return "Timed out"
	}
	return "Completed"
}
