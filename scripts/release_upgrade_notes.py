#!/usr/bin/env python3
"""Move upgrade notes that shipped in a release under that release's heading.

Upgrade notes for changes that are merged but not yet released live under the
"Next release (unreleased)" section of upgrade_notes.adoc. When a release is
tagged, every note that was in the unreleased section *at the tagged commit*
moves into a new "== <version>" section. Notes added to the unreleased section
after the release was cut stay where they are.

Usage:
    release_upgrade_notes.py <version> <released-file> <target-file>

<released-file> is upgrade_notes.adoc as it was at the release tag, and
<target-file> is the copy to update in place (normally on the next branch).
"""

from __future__ import annotations

import sys
from pathlib import Path

UNRELEASED_HEADING = "== Next release (unreleased)"


def split_sections(lines: list[str], prefix: str) -> tuple[list[str], list[list[str]]]:
    """Split lines into a preamble and blocks that each start at a heading.

    An anchor line such as [#some-id] directly above a heading belongs to that
    heading's block, so that anchors move with their notes.
    """
    preamble: list[str] = []
    blocks: list[list[str]] = []

    for line in lines:
        if line.startswith(prefix) and not line.startswith(prefix + "="):
            anchor = _take_anchor(preamble if not blocks else blocks[-1])
            blocks.append(anchor + [line])
        elif blocks:
            blocks[-1].append(line)
        else:
            preamble.append(line)

    return preamble, blocks


def _take_anchor(block: list[str]) -> list[str]:
    if block and block[-1].startswith("[#"):
        return [block.pop()]
    return []


def note_key(note: list[str]) -> str:
    """Identify a note by its anchor, falling back to its heading."""
    for line in note:
        if line.startswith("[#") or line.startswith("=== "):
            return line.strip()
    return ""


def section_heading(section: list[str]) -> str:
    for line in section:
        if not line.startswith("[#"):
            return line.rstrip()
    return ""


def unreleased_notes(lines: list[str]) -> tuple[list[str], list[list[str]]] | None:
    _, sections = split_sections(lines, "== ")
    for section in sections:
        if section_heading(section) == UNRELEASED_HEADING:
            return split_sections(section, "=== ")
    return None


def released_keys(released_lines: list[str]) -> set[str]:
    found = unreleased_notes(released_lines)
    if found is None:
        return set()
    return {note_key(note) for note in found[1]}


def build_version_section(version: str, notes: list[list[str]]) -> list[str]:
    section = [f"== {version}\n", "\n"]
    for note in notes:
        section.extend(note)
    return section


def build_unreleased_section(heading: list[str], notes: list[list[str]]) -> list[str]:
    if not notes:
        return []
    section = list(heading)
    for note in notes:
        section.extend(note)
    return section


def move_released_notes(version: str, released_lines: list[str], target_lines: list[str]) -> list[str] | None:
    """Return the updated target lines, or None when there is nothing to move."""
    keys = released_keys(released_lines)
    if not keys:
        return None

    preamble, sections = split_sections(target_lines, "== ")
    if any(section_heading(section) == f"== {version}" for section in sections):
        return None

    result = list(preamble)
    for section in sections:
        if section_heading(section) != UNRELEASED_HEADING:
            result.extend(section)
            continue

        heading, notes = split_sections(section, "=== ")
        shipped = [note for note in notes if note_key(note) in keys]
        remaining = [note for note in notes if note_key(note) not in keys]

        result.extend(build_unreleased_section(heading, remaining))
        result.extend(build_version_section(version, shipped))

    return result


def main(argv: list[str]) -> int:
    if len(argv) != 4:
        print(__doc__, file=sys.stderr)
        return 2

    version, released_path, target_path = argv[1], Path(argv[2]), Path(argv[3])
    released_lines = released_path.read_text().splitlines(keepends=True)
    target_lines = target_path.read_text().splitlines(keepends=True)

    updated = move_released_notes(version, released_lines, target_lines)
    if updated is None:
        print(f"Nothing to move for {version}: no unreleased upgrade notes shipped, or the section already exists.")
        return 0

    target_path.write_text("".join(updated))
    print(f"Moved released upgrade notes under '== {version}' in {target_path}.")
    return 0


if __name__ == "__main__":
    sys.exit(main(sys.argv))
