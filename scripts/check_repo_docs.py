#!/usr/bin/env python3
from __future__ import annotations

import argparse
import re
import sys
from pathlib import Path


LINK_RE = re.compile(r'(?<![\w`])!?\[[^\]]*\]\(([^)]+)\)|href="([^"]+)"')
IGNORED_PREFIXES = ("http://", "https://", "mailto:", "tel:", "data:", "javascript:")
DEFAULT_DOC_GLOBS = (
    "README.md",
    "CONTRIBUTING.md",
    "SUPPORT.md",
    ".github/PULL_REQUEST_TEMPLATE.md",
    "docs/**/*.md",
)


def iter_doc_files(repo_root: Path, explicit_paths: list[str]) -> list[Path]:
    if explicit_paths:
        files = []
        for raw in explicit_paths:
            path = (repo_root / raw).resolve()
            if path.is_file():
                files.append(path)
        return sorted(set(files))

    files = set()
    for pattern in DEFAULT_DOC_GLOBS:
        files.update(path for path in repo_root.glob(pattern) if path.is_file())
    return sorted(files)


def extract_targets(text: str) -> list[str]:
    targets = []
    for match in LINK_RE.finditer(text):
        target = match.group(1) or match.group(2)
        if target:
            targets.append(target.strip())
    return targets


def normalize_target(target: str) -> str | None:
    target = target.strip()
    if not target or target.startswith("#") or target.startswith("/"):
        return None
    if target.startswith(IGNORED_PREFIXES):
        return None
    if target.startswith("<") and target.endswith(">"):
        target = target[1:-1]
    target = target.split("#", 1)[0].split("?", 1)[0]
    return target or None


def check_files(repo_root: Path, files: list[Path]) -> list[str]:
    errors: list[str] = []
    for path in files:
        for target in extract_targets(path.read_text()):
            normalized = normalize_target(target)
            if normalized is None:
                continue
            resolved = (path.parent / normalized).resolve()
            if not resolved.exists():
                rel_source = path.relative_to(repo_root)
                errors.append(
                    f"{rel_source}: missing local docs target {normalized!r}"
                )
    return errors


def main(argv: list[str] | None = None) -> int:
    parser = argparse.ArgumentParser(description="Validate local links in repo docs.")
    parser.add_argument("paths", nargs="*", help="Optional subset of doc files to check.")
    args = parser.parse_args(argv)

    repo_root = Path(__file__).resolve().parents[1]
    files = iter_doc_files(repo_root, args.paths)
    errors = check_files(repo_root, files)
    if errors:
        print("Repository docs validation failed:", file=sys.stderr)
        for error in errors:
            print(f"  - {error}", file=sys.stderr)
        return 1

    print(f"Repository docs validation passed for {len(files)} file(s).")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
