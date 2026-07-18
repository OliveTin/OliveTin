#!/usr/bin/env python3
"""Scan project Makefiles for development tools and report PATH availability."""

from __future__ import annotations

import os
import re
import shutil
import sys
from collections import defaultdict
from pathlib import Path

ROOT = Path(__file__).resolve().parent.parent

FAIL_GROUPS = frozenset({
    "Core build and test",
    "Go codestyle (install via: make go-tools)",
    "Protocol buffers (install via: make -C service go-tools-all)",
})

SKIP_DIRS = frozenset({
    ".git",
    "node_modules",
    "vendor",
    "dist",
    "reports",
    "webui",
})

SKIP_COMMANDS = frozenset({
    ".", ":", "[", "bash", "case", "cat", "cd", "chmod", "chown", "cp", "curl",
    "do", "done", "echo", "elif", "else", "esac", "exit", "false", "fi", "for",
    "fuser", "grep", "head", "if", "kill", "killall", "lsof", "make", "mkdir",
    "mv", "objdump", "pwd", "rm", "sed", "set", "sh", "sleep", "tail", "test",
    "then", "touch", "trap", "true", "unzip", "while",
})

SKIP_PREFIXES = ("./", "../", "-", "$(")

GO_INSTALL_RE = re.compile(r"""go\s+install\s+(?:["'])([^"']+)(?:["'])""")

TOOL_GROUPS: list[tuple[str, frozenset[str]]] = [
    (
        "Core build and test",
        frozenset({"go", "npm", "npx", "node", "python3"}),
    ),
    (
        "Go codestyle (install via: make go-tools)",
        frozenset({"golangci-lint"}),
    ),
    (
        "Protocol buffers (install via: make -C service go-tools-all)",
        frozenset({"buf", "protoc-gen-go"}),
    ),
    (
        "Containers and packaging (optional)",
        frozenset({"buildah", "docker", "podman", "podman-compose"}),
    ),
]


class Color:
    RESET = "\033[0m"
    BOLD = "\033[1m"
    GREEN = "\033[32m"
    RED = "\033[31m"
    ORANGE = "\033[38;5;208m"
    DIM = "\033[2m"

    @classmethod
    def ok(cls, text: str) -> str:
        return f"{cls.BOLD}{cls.GREEN}{text}{cls.RESET}"

    @classmethod
    def fail(cls, text: str) -> str:
        return f"{cls.BOLD}{cls.RED}{text}{cls.RESET}"

    @classmethod
    def warn(cls, text: str) -> str:
        return f"{cls.BOLD}{cls.ORANGE}{text}{cls.RESET}"

    @classmethod
    def dim(cls, text: str) -> str:
        return f"{cls.DIM}{text}{cls.RESET}"


def colors_enabled() -> bool:
    if os.environ.get("NO_COLOR"):
        return False
    if "--no-color" in sys.argv:
        return False
    return sys.stdout.isatty()


def paint_ok(text: str) -> str:
    return Color.ok(text) if colors_enabled() else text


def paint_fail(text: str) -> str:
    return Color.fail(text) if colors_enabled() else text


def paint_warn(text: str) -> str:
    return Color.warn(text) if colors_enabled() else text


def paint_dim(text: str) -> str:
    return Color.dim(text) if colors_enabled() else text


def find_makefiles(root: Path) -> list[Path]:
    makefiles: list[Path] = []
    for path in root.rglob("Makefile"):
        rel_parts = path.relative_to(root).parts
        if any(part.startswith(".") or part in SKIP_DIRS for part in rel_parts):
            continue
        makefiles.append(path)
    return sorted(makefiles)


def tool_from_go_install(package: str) -> str:
    return package.rstrip("/").rsplit("/", 1)[-1]


def first_command_token(line: str) -> str | None:
    line = line.strip()
    if not line or line.startswith("#"):
        return None

    if line.startswith("$(call") or line.startswith("$(MAKE)"):
        return None

    for token in line.split():
        if "=" in token and not token.startswith("./"):
            continue
        if token.startswith(SKIP_PREFIXES):
            return None
        return token.strip("\"'")
    return None


def collect_tools(makefiles: list[Path]) -> dict[str, set[str]]:
    tools: dict[str, set[str]] = defaultdict(set)

    for makefile in makefiles:
        rel = makefile.relative_to(ROOT).as_posix()
        try:
            lines = makefile.read_text(encoding="utf-8", errors="replace").splitlines()
        except OSError:
            continue

        for line in lines:
            if not line.startswith("\t"):
                continue

            recipe = line.lstrip("\t@")
            for package in GO_INSTALL_RE.findall(recipe):
                tools[tool_from_go_install(package)].add(rel)

            if "python -c" in recipe or recipe.startswith("python "):
                tools["python3"].add(rel)

            command = first_command_token(recipe)
            if command is None:
                continue

            command = command.lower()
            if command in SKIP_COMMANDS or command == "python":
                continue
            if "/" in command and command not in {"podman-compose"}:
                continue

            tools[command].add(rel)

    return dict(tools)


def resolve_python() -> tuple[bool, str | None]:
    for name in ("python3", "python"):
        path = shutil.which(name)
        if path is not None:
            return True, path
    return False, None


def resolve_tool(name: str) -> tuple[bool, str | None]:
    if name == "python3":
        return resolve_python()
    path = shutil.which(name)
    return path is not None, path


def group_tools(tools: dict[str, set[str]]) -> list[tuple[str, list[str]]]:
    grouped: list[tuple[str, list[str]]] = []
    assigned: set[str] = set()

    for title, members in TOOL_GROUPS:
        present = sorted(tool for tool in tools if tool in members)
        if present:
            grouped.append((title, present))
            assigned.update(present)

    remaining = sorted(tool for tool in tools if tool not in assigned)
    if remaining:
        grouped.append(("Other tools found in Makefiles", remaining))

    return grouped


def format_status(found: bool, path: str | None, *, required: bool) -> str:
    if found:
        return paint_ok(f"OK      {path}")

    if required:
        return paint_fail("MISSING")
    return paint_warn("MISSING")


def print_group(
    title: str,
    tool_names: list[str],
    tools: dict[str, set[str]],
    verbose: bool,
) -> tuple[int, int, list[str]]:
    required = title in FAIL_GROUPS
    print(title)
    available = 0
    missing: list[str] = []

    for name in tool_names:
        found, path = resolve_tool(name)
        available += int(found)
        if not found:
            missing.append(name)
        print(f"  {name:<16} {format_status(found, path, required=required)}")
        if verbose:
            for source in sorted(tools[name]):
                print(paint_dim(f"    referenced in {source}"))
    print()
    return available, len(tool_names), missing


def main() -> int:
    verbose = "--verbose" in sys.argv or "-v" in sys.argv
    makefiles = find_makefiles(ROOT)
    tools = collect_tools(makefiles)

    print("Development environment check")
    print(f"Scanned {len(makefiles)} Makefile(s) under {ROOT}")
    print()

    total_available = 0
    total_checked = 0
    required_missing: list[str] = []
    optional_missing: list[str] = []

    for title, tool_names in group_tools(tools):
        available, checked, missing = print_group(title, tool_names, tools, verbose)
        total_available += available
        total_checked += checked

        if title in FAIL_GROUPS:
            required_missing.extend(missing)
        else:
            optional_missing.extend(missing)

    summary = f"Summary: {total_available}/{total_checked} tools available on PATH"
    if required_missing:
        print(paint_fail(summary))
    elif optional_missing:
        print(paint_warn(summary))
    else:
        print(paint_ok(summary))

    if required_missing:
        print()
        print(paint_fail("Missing required tools: " + ", ".join(required_missing)))
        print(paint_dim("Install the missing tools for your platform, then re-run: make devcheck"))
        return 1

    if optional_missing:
        print()
        print(paint_warn("Missing optional tools: " + ", ".join(optional_missing)))

    return 0


if __name__ == "__main__":
    sys.exit(main())
