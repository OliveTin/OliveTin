#!/usr/bin/env python3

# Find config option names in docs that do not match YAML keys from config.go.
#
# OliveTin config uses camelCase koanf tags (see service/internal/config/config.go).
# Docs sometimes use Go struct field names (PascalCase) or other wrong spellings.

from __future__ import annotations

from dataclasses import dataclass
from pathlib import Path
import re
import sys

ROOT = Path(__file__).resolve().parent
REPO_ROOT = ROOT.parents[2]
CONFIG_GO = REPO_ROOT / "service/internal/config/config.go"
DOC_DIRS = (ROOT / "pages", ROOT / "partials")

STRUCT_START_RE = re.compile(r"^type (\w+) struct\b")
FIELD_RE = re.compile(
    r'^\s+(\w+)\s+([^`]+?)`koanf:"([^"]+)"`',
)
BACKTICK_RE = re.compile(r"`([^`]+)`")
YAML_BLOCK_RE = re.compile(
    r"\[source,yaml\][^\n]*\n----\n(.*?)\n----",
    re.DOTALL,
)
YAML_KEY_RE = re.compile(r"^(\s*)([A-Za-z][\w]*)\s*:", re.MULTILINE)

SKIP_BACKTICK = frozenset({
    "Insecure*",
})

SKIP_YAML_PREFIXES = frozenset({
    "actions",
    "dashboards",
    "entities",
    "title",
    "shell",
    "icon",
    "arguments",
    "name",
    "type",
    "default",
    "description",
    "choices",
    "value",
    "permissions",
    "view",
    "exec",
    "logs",
    "kill",
    "matchUsergroups",
    "matchUsernames",
    "policy",
    "users",
    "username",
    "password",
    "usergroup",
    "enabled",
    "acls",
    "groups",
    "maxConcurrent",
    "timeout",
    "onclick",
    "execOnStartup",
    "maxRate",
    "limit",
    "duration",
    "id",
    "hidden",
    "category",
    "contents",
    "file",
    "properties",
    "inlineAction",
    "resultsDirectory",
    "outputDirectory",
    "directory",
    "showDiagnostics",
    "showLogList",
    "showVersionNumber",
    "headerSearch",
    "defaultGoMetrics",
    "contentSecurityPolicy",
    "xFrameOptions",
    "headerContentSecurityPolicy",
    "headerXContentTypeOptions",
    "headerXFrameOptions",
    "forceSecureCookies",
    "clientId",
    "clientSecret",
    "authUrl",
    "tokenUrl",
    "whoamiUrl",
    "scopes",
    "addToUsergroup",
    "userGroupField",
    "usernameField",
    "certBundlePath",
    "callbackTimeout",
    "insecureSkipVerify",
    "secret",
    "authType",
    "authHeader",
    "matchHeaders",
    "matchPath",
    "matchQuery",
    "extract",
    "template",
    "justification",
    "apiKey",
    "addToEveryAction",
    "execOnCron",
    "execOnCalendarFile",
    "shellAfterCompleted",
    "execOnWebhook",
    "triggers",
    "exec",
    "execOnFileCreatedInDir",
    "execOnFileChangedInDir",
    "entity",
    "popupOnStart",
    "saveLogs",
    "suggestions",
    "suggestionsBrowserKey",
    "rejectNull",
    "queueSize",
    "cssClass",
    "url",
    "target",
    "styleMods",
    "include",
    "bannerCss",
    "bannerMessage",
    "serviceHostMode",
    "themeCacheDisabled",
    "checkForUpdates",
    "logHistoryPageSize",
    "additionalNavigationLinks",
    "actionGroups",
    "authOAuth2Providers",
    "authOAuth2RedirectUrl",
    "authJwtHmacSecret",
})


@dataclass(frozen=True)
class Issue:
    path: str
    line: int
    found: str
    expected: str
    kind: str


def parse_structs(content: str) -> dict[str, list[tuple[str, str, str]]]:
    structs: dict[str, list[tuple[str, str, str]]] = {}
    current: str | None = None

    for line in content.splitlines():
        struct_match = STRUCT_START_RE.match(line)
        if struct_match:
            current = struct_match.group(1)
            structs[current] = []
            continue

        if current is None:
            continue

        if line.strip() == "}":
            current = None
            continue

        field_match = FIELD_RE.match(line)
        if not field_match:
            continue

        field_name, field_type, koanf_tag = field_match.groups()
        if koanf_tag == "-":
            continue

        structs[current].append((field_name, field_type.strip(), koanf_tag.strip()))

    return structs


def is_nested_struct(field_type: str, structs: dict[str, list[tuple[str, str, str]]]) -> str | None:
    inner = field_type.removeprefix("[]").removeprefix("*").strip()
    if inner in structs and inner not in {
        "Action",
        "EntityFile",
        "AccessControlList",
        "DashboardComponent",
        "NavigationLink",
        "OAuth2Provider",
        "LocalUser",
        "ActionArgument",
        "ActionArgumentChoice",
        "RateSpec",
        "WebhookConfig",
        "EntityProperty",
        "ActionGroup",
    }:
        return inner
    return None


def collect_config_keys(
    structs: dict[str, list[tuple[str, str, str]]],
) -> tuple[frozenset[str], dict[str, str]]:
    valid: set[str] = set()
    aliases: dict[str, str] = {}

    def walk(type_name: str, prefix: str = "") -> None:
        for field_name, _field_type, koanf_tag in structs.get(type_name, []):
            path = f"{prefix}.{koanf_tag}" if prefix else koanf_tag
            valid.add(path)

            if field_name != koanf_tag:
                aliases[field_name] = koanf_tag
                if prefix:
                    aliases[f"{prefix}.{field_name}"] = path

            nested = is_nested_struct(_field_type, structs)
            if nested:
                walk(nested, path)

    walk("Config")
    return frozenset(valid), aliases


def camelize_path(key: str) -> str:
    parts = []
    for part in key.split("."):
        if part and part[0].isupper():
            parts.append(part[0].lower() + part[1:])
        else:
            parts.append(part)
    return ".".join(parts)


def looks_like_config_key(key: str) -> bool:
    if not key or key in SKIP_BACKTICK:
        return False
    if "*" in key or " " in key or "/" in key or ":" in key:
        return False
    return bool(re.fullmatch(r"[A-Za-z][\w.]*", key))


def has_internal_uppercase(key: str) -> bool:
    if "." in key:
        return any(has_internal_uppercase(part) for part in key.split("."))
    return any(char.isupper() for char in key[1:])


def case_insensitive_match(key: str, valid: frozenset[str]) -> str | None:
    matches = [candidate for candidate in valid if candidate.lower() == key.lower()]
    if len(matches) == 1:
        return matches[0]
    return None


def resolve_key(
    key: str,
    valid: frozenset[str],
    aliases: dict[str, str],
    *,
    allow_case_insensitive: bool = False,
) -> str | None:
    if key in valid:
        return None

    if key in aliases and key != aliases[key]:
        return aliases[key]

    if allow_case_insensitive:
        matched = case_insensitive_match(key, valid)
        if matched is not None and matched != key:
            return matched

    if not has_internal_uppercase(key):
        return None

    camelized = camelize_path(key)
    if camelized in valid and key != camelized:
        return camelized

    return None


def scan_backticks(
    rel_path: str,
    content: str,
    valid: frozenset[str],
    aliases: dict[str, str],
) -> list[Issue]:
    issues: list[Issue] = []

    for line_number, line in enumerate(content.splitlines(), start=1):
        for match in BACKTICK_RE.finditer(line):
            key = match.group(1).strip()
            if not looks_like_config_key(key):
                continue

            expected = resolve_key(key, valid, aliases)
            if expected is None:
                continue

            # Backticks often label UI sections that share a name with config keys.
            if expected == "actions" and key == "Actions":
                continue

            issues.append(
                Issue(
                    path=rel_path,
                    line=line_number,
                    found=key,
                    expected=expected,
                    kind="backtick",
                )
            )

    return issues


def scan_yaml_blocks(
    rel_path: str,
    content: str,
    valid: frozenset[str],
    aliases: dict[str, str],
) -> list[Issue]:
    issues: list[Issue] = []

    for block in YAML_BLOCK_RE.finditer(content):
        block_text = block.group(1)
        block_start = block.start(1)

        for match in YAML_KEY_RE.finditer(block_text):
            indent = len(match.group(1).replace("\t", "    "))
            key = match.group(2)

            if indent != 0:
                continue
            if key in SKIP_YAML_PREFIXES:
                continue
            if key in valid:
                continue

            expected = resolve_key(
                key,
                valid,
                aliases,
                allow_case_insensitive=True,
            )
            if expected is None:
                continue

            issues.append(
                Issue(
                    path=rel_path,
                    line=content.count("\n", 0, block_start + match.start()) + 1,
                    found=key,
                    expected=expected,
                    kind="yaml",
                )
            )

    return issues


def iter_doc_files() -> list[Path]:
    files: list[Path] = []
    for doc_dir in DOC_DIRS:
        files.extend(sorted(doc_dir.rglob("*.adoc")))
    return files


def main() -> int:
    if not CONFIG_GO.is_file():
        print(f"config.go not found: {CONFIG_GO}", file=sys.stderr)
        return 2

    structs = parse_structs(CONFIG_GO.read_text())
    valid, aliases = collect_config_keys(structs)

    issues: list[Issue] = []
    for doc_path in iter_doc_files():
        content = doc_path.read_text()
        rel_path = str(doc_path.relative_to(ROOT))
        issues.extend(scan_backticks(rel_path, content, valid, aliases))
        issues.extend(scan_yaml_blocks(rel_path, content, valid, aliases))

    if not issues:
        print("No config key casing issues found.")
        return 0

    print(f"Issues: {len(issues)}")
    for issue in issues:
        print(
            f"{issue.path}:{issue.line}: [{issue.kind}] "
            f"`{issue.found}` should be `{issue.expected}`"
        )

    return 1


if __name__ == "__main__":
    sys.exit(main())
