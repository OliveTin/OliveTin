#!/usr/bin/env python3
"""Set or list the [status:...] tag on OliveTin's repository security advisories.

Advisories in triage or draft carry a status tag at the end of their summary,
so maintainers and reporters can see where each one is. The tag is removed
when an advisory is published or closed, and before a CVE is requested.

Usage:
    ghsa_status.py --list
    ghsa_status.py <GHSA-ID> <status>
    ghsa_status.py <GHSA-ID> clear

Statuses: pending-review, awaiting-reporter, accepted, feedback-on-fix,
pending-release. Requires the GitHub CLI (gh), authenticated with access to
the repository's security advisories.
"""

from __future__ import annotations

import json
import re
import subprocess
import sys

REPO = "OliveTin/OliveTin"

STATUSES = (
    "pending-review",
    "awaiting-reporter",
    "accepted",
    "feedback-on-fix",
    "pending-release",
)

OPEN_STATES = ("triage", "draft")

TRAILING_TAGS_PATTERN = re.compile(r"(?:\s*\[status:[^\]]*\])+\s*$")
TAG_PATTERN = re.compile(r"\[status:([^\]]*)\]")
GHSA_PATTERN = re.compile(r"^GHSA(-[23456789cfghjmpqrvwx]{4}){3}$")


def gh_api(path: str, method: str = "GET", body: dict | None = None, paginate: bool = False) -> object:
    command = ["gh", "api", "-X", method, path]
    if paginate:
        command += ["--paginate", "--slurp"]
    if body is not None:
        command += ["--input", "-"]
    result = subprocess.run(
        command,
        input=json.dumps(body) if body is not None else None,
        capture_output=True,
        text=True,
        check=False,
    )
    if result.returncode != 0:
        raise SystemExit(f"gh api {method} {path} failed: {result.stderr.strip() or result.stdout.strip()}")
    return json.loads(result.stdout)


def strip_status(summary: str) -> str:
    return TRAILING_TAGS_PATTERN.sub("", summary).rstrip()


def status_tags(summary: str) -> list[str]:
    match = TRAILING_TAGS_PATTERN.search(summary)
    if not match:
        return []
    return TAG_PATTERN.findall(match.group(0))


def has_valid_status(tags: list[str]) -> bool:
    return len(tags) == 1 and tags[0] in STATUSES


def with_status(summary: str, status: str) -> str:
    base = strip_status(summary)
    if status == "clear":
        return base
    return f"{base} [status:{status}]"


def list_advisories() -> int:
    pages = gh_api(f"/repos/{REPO}/security-advisories?per_page=100", paginate=True)
    advisories = [advisory for page in pages for advisory in page]
    problems = 0

    for advisory in advisories:
        tags = status_tags(advisory["summary"])
        is_open = advisory["state"] in OPEN_STATES

        note = ""
        if is_open and not has_valid_status(tags):
            note = "  <- missing, unknown or multiple statuses"
            problems += 1
        elif not is_open and tags:
            note = "  <- should not have a status"
            problems += 1

        if is_open or tags:
            shown = ",".join(tag or "(empty)" for tag in tags) or "-"
            print(f"{advisory['ghsa_id']}  {advisory['state']:<9} {shown:<18} {strip_status(advisory['summary'])}{note}")

    return 1 if problems else 0


def set_status(ghsa_id: str, status: str) -> int:
    if not GHSA_PATTERN.match(ghsa_id):
        raise SystemExit(f"Not a GHSA ID: {ghsa_id}")
    if status != "clear" and status not in STATUSES:
        raise SystemExit(f"Unknown status '{status}'. Use one of: {', '.join(STATUSES)}, or clear")

    advisory = gh_api(f"/repos/{REPO}/security-advisories/{ghsa_id}")
    if status != "clear" and advisory["state"] not in OPEN_STATES:
        raise SystemExit(f"{ghsa_id} is {advisory['state']}; only triage and draft advisories carry a status")

    summary = with_status(advisory["summary"], status)
    if summary == advisory["summary"]:
        print(f"{ghsa_id} already has that status.")
        return 0

    updated = gh_api(f"/repos/{REPO}/security-advisories/{ghsa_id}", method="PATCH", body={"summary": summary})
    print(f"{ghsa_id}: {updated['summary']}")
    return 0


def main(argv: list[str]) -> int:
    if argv[1:] == ["--list"]:
        return list_advisories()
    if len(argv) == 3:
        return set_status(argv[1], argv[2])

    print(__doc__, file=sys.stderr)
    return 2


if __name__ == "__main__":
    sys.exit(main(sys.argv))
