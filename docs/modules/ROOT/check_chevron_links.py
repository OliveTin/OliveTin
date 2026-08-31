#!/usr/bin/env python3

# Find <<anchor>> links whose target is not defined on the assembled page.
# include::partial$... directives are expanded so chevrons inside partials are
# checked against every page that includes them.

from pathlib import Path
import re
import sys

ROOT = Path(__file__).resolve().parent
PAGES = ROOT / "pages"
PARTIALS = ROOT / "partials"
PARTIALS_ROOT = PARTIALS.resolve()

INCLUDE_RE = re.compile(r"include::partial\$([^\[\]]+)\[([^\]]*)\]")
CHEVRON_RE = re.compile(r"<<([^,>]+)(?:,[^>]*)?>>")


def expand_partials(content, stack):
    def replace(match):
        rel = match.group(1)
        partial_path = (PARTIALS / rel).resolve()

        if not partial_path.is_relative_to(PARTIALS_ROOT):
            return ""

        if not partial_path.is_file() or partial_path in stack:
            return ""

        included = partial_path.read_text()
        return expand_partials(included, stack | {partial_path})

    return INCLUDE_RE.sub(replace, content)


def count_anchors(content, anchor_id):
    escaped = re.escape(anchor_id)
    pattern = rf"\[(?:#|\[){escaped}(?=[,\]])"
    return len(re.findall(pattern, content))


def find_unresolved(content):
    unresolved = []

    for match in CHEVRON_RE.finditer(content):
        anchor_id = match.group(1).strip()

        if count_anchors(content, anchor_id) == 1:
            continue

        if anchor_id not in unresolved:
            unresolved.append(anchor_id)

    return unresolved


def main():
    filelist = {}

    for page in sorted(PAGES.rglob("*.adoc")):
        assembled = expand_partials(page.read_text(), set())
        missing = find_unresolved(assembled)

        if missing:
            filelist[str(page.relative_to(ROOT))] = missing

    print("Files:", len(filelist))

    for file in filelist:
        print(file)

        for match in filelist[file]:
            print("\t", match)

    sys.exit(1 if filelist else 0)


if __name__ == "__main__":
    main()
