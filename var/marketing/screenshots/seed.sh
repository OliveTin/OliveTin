#!/usr/bin/env bash
# Runs a few fictional actions through OliveTin's public API so that the
# logs and execution pages have history to show. Fixed tracking ids keep
# the logs table and execution URLs the same on every run.
# Usage: seed.sh <base-url>
set -euo pipefail

base="${1:?base url}"
api="$base/api/olivetin.api.v1.OliveTinApiService"

rpc() {
	curl -fsS -X POST -H 'Content-Type: application/json' -d "$2" "$api/$1"
}

run() {
	local tracking_id="$1" binding_id="$2" arguments="${3:-[]}"

	rpc StartAction "{\"bindingId\": \"$binding_id\", \"uniqueTrackingId\": \"$tracking_id\", \"arguments\": $arguments}" >/dev/null

	for _ in $(seq 50); do
		rpc ExecutionStatus "{\"executionTrackingId\": \"$tracking_id\"}" |
			jq -e '.logEntry.executionFinished' >/dev/null && return 0
		sleep 0.1
	done

	echo "Action $binding_id did not finish" >&2
	return 1
}

run 5eed0000-0000-4000-8000-000000000001 run_playbook
run 5eed0000-0000-4000-8000-000000000002 nightly_backup
run 5eed0000-0000-4000-8000-000000000003 restart_container '[{"name": "container", "value": "demo-web"}]'
run 5eed0000-0000-4000-8000-000000000004 restart_media
run 5eed0000-0000-4000-8000-000000000005 porch_light
run 5eed0000-0000-4000-8000-000000000006 ping_host '[{"name": "host", "value": "nas.example.com"}, {"name": "count", "value": "4"}]'
run 5eed0000-0000-4000-8000-000000000007 check_disk
