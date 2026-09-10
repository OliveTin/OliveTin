#!/usr/bin/env python3

from pathlib import Path

ROOT = Path(__file__).resolve().parent.parent
GENERATED_DIRS = (
    ROOT / "service/gen",
    ROOT / "frontend/resources/scripts/gen",
)


def normalize_file(path: Path) -> None:
    content = path.read_bytes()
    normalized = content.rstrip(b"\r\n") + b"\n"

    if normalized != content:
        path.write_bytes(normalized)


def main() -> None:
    for directory in GENERATED_DIRS:
        for path in directory.rglob("*"):
            if path.is_file():
                normalize_file(path)


if __name__ == "__main__":
    main()
