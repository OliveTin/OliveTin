# Configuration issues

This spec describes how OliveTin collects configuration warnings and errors and surfaces them in the web UI.

## Purpose

Operators should see configuration problems in Diagnostics instead of only in server logs. When any issues exist, the Diagnostics navigation link shows a count badge.

## When issues are rebuilt

The issue list is cleared and rebuilt when configuration is loaded or reloaded, and when entity data changes. Some findings that can only be detected while configuration is first being loaded (for example references to unset environment variables that are expanded away during that load) are kept across later rebuilds until the next configuration load begins.

## What is collected

Issues include:

- Unknown or unenforced action group references
- Checklist arguments with missing or invalid choice templates
- Arguments whose type was left unset (defaulted to a generic text type)
- Unset environment variables referenced from configuration
- Missing or invalid include directories
- Argument default or choice templates that fail to parse
- Literal argument defaults that fail type validation (templated defaults are not type-checked as raw text)
- Entity files that cannot be read or parsed, or are empty
- Entity-bound actions with no entity instances (after OliveTin has attempted to load that entity type, so startup does not report a false positive before entity files are read)
- Invalid cron schedules
- Entity-bound actions that also use scheduled cron execution (cron runs without an entity binding and will not execute)
- Filesystem watch paths that cannot be created (missing directories for file-in-dir triggers, calendar files, or entity files). Runtime watcher setup failures for action triggers include the related action so view permissions still apply; entity-file watchers without an action remain visible to anyone who may view Diagnostics.

Each issue has a severity of warning or error, a stable code, a human-readable message, and optional context such as action title, argument name, configuration source file, or detail value.

When the issue list is rebuilt, OliveTin logs only newly appeared issues so startup does not repeat the same warning for every rebuild.

When configuration is loaded from a base file and an include directory, OliveTin records which file defined each action and entity declaration. That path is shown as the configuration source file when available. Some issues (for example unset environment variables) may not have a specific file. Entity data file problems also show the entity data path in the detail column.

## Diagnostics page

Users who are allowed to view Diagnostics see a Configuration issues section listing the current issues in a table. When there are none, the section states that no configuration issues were detected.

Action-scoped issues are only included when the user is allowed to view that action. Issues that are not tied to an action (for example unset environment variables or missing include directories) remain visible to anyone who may view Diagnostics.

Users who are not allowed to view Diagnostics cannot retrieve the issue list.

## Navigation count

When Diagnostics is visible and at least one configuration issue exists that the user is allowed to see, the Diagnostics navigation link shows a count badge with the number of those issues. The badge clears when the visible issue count becomes zero after a configuration or entity refresh.

## Startup count

When the web UI starts, users who may view Diagnostics receive the same filtered configuration issue count used for the Diagnostics list and navigation badge. For other users the count is zero.
