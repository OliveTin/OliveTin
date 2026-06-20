#!/bin/bash

echo "Starting backup script"

# Demo behaviour: 50% finish in 9 seconds (within the 10s action timeout),
# 50% run for 15 seconds (typically times out).
if (( RANDOM % 2 == 0 )); then
	maxFiles=9
	echo "Demo: backup will finish in 9 seconds"
else
	maxFiles=15
	echo "Demo: backup will run for 15 seconds (may exceed the action timeout)"
fi

for fileIndex in $(seq 1 "$maxFiles"); do
	echo "Backing up file: $fileIndex"
	sleep 1
done

echo "All files backed up"
