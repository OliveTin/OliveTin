# Action group concurrency

Actions may belong to one or more named groups. Each group may define a maximum number of concurrent executions shared across all actions in that group.

When a user or trigger starts an action that belongs to a group, OliveTin counts how many executions for that group are currently active. Active means the execution has been requested but not yet finished, and is not waiting in a queue.

If every configured group for that action has spare capacity, the execution proceeds through the normal execution pipeline.

If any configured group is at capacity, the new execution is queued instead of rejected. The request receives a tracking identifier immediately. The log entry shows a queued status until the execution actually starts.

Queued executions run in first-in-first-out order per OliveTin instance. When an active execution in a group finishes, OliveTin attempts to start the oldest queued execution that belongs to that group, provided all groups for that queued action now have spare capacity.

An action may belong to multiple groups. In that case, all group limits must be satisfied before the action starts or leaves the queue.

Per-action concurrency limits apply only to executions of the same action binding. When a per-action limit is exceeded, the request is blocked immediately and is not queued.

Action group concurrency limits do not survive a process restart. Queued executions that have not started are discarded when OliveTin stops.

If an action references a group name that is not defined in configuration, OliveTin logs a warning and does not apply a group limit for that name.
