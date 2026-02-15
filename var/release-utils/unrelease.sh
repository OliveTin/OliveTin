#!/usr/bin/env bash
set -euo pipefail

RELEASE_NAME="${1:-}"
GHCR_IMAGE="ghcr.io/olivetin/olivetin"
DOCKERHUB_IMAGE="jamesread/olivetin"

log() {
  echo "[unrelease] $*"
}

prompt_confirm() {
  local prompt="$1"
  local default="${2:-n}"
  if [[ "$default" == "y" ]]; then
    read -r -p "$prompt [Y/n] " reply
  else
    read -r -p "$prompt [y/N] " reply
  fi
  reply="${reply:-$default}"
  case "$(echo "$reply" | tr '[:upper:]' '[:lower:]')" in
    y|yes) return 0 ;;
    *) return 1 ;;
  esac
}

if [[ -z "$RELEASE_NAME" ]]; then
  echo "Usage: $0 <release_name>" >&2
  echo "Example: $0 3000.10.0" >&2
  exit 1
fi

log "Release to remove: $RELEASE_NAME"
log "This will: 1) Delete GitHub release, 2) Delete GitHub tag, 3) Delete GHCR image tag, 4) Delete Docker Hub image tag"
echo

# --- GitHub release ---
log "Step 1: Delete GitHub release '$RELEASE_NAME'"
if prompt_confirm "Delete GitHub release?" "n"; then
  if err=$(gh release delete "$RELEASE_NAME" --yes 2>&1); then
    log "Deleted GitHub release."
  else
    log "Failed to delete GitHub release:" >&2
    echo "$err" | sed 's/^/[unrelease]   /' >&2
  fi
else
  log "Skipped GitHub release."
fi
echo

# --- GitHub tag ---
log "Step 2: Delete GitHub tag '$RELEASE_NAME'"
if prompt_confirm "Delete GitHub tag?" "n"; then
  repo=$(gh repo view --json nameWithOwner -q .nameWithOwner 2>/dev/null) || repo="olivetin/olivetin"
  if err=$(gh api -X DELETE "repos/$repo/git/refs/tags/$RELEASE_NAME" 2>&1); then
    log "Deleted GitHub tag."
  else
    log "Failed to delete GitHub tag:" >&2
    echo "$err" | sed 's/^/[unrelease]   /' >&2
  fi
else
  log "Skipped GitHub tag."
fi
echo

# --- GHCR ---
log "Step 3: Delete GHCR image tag $GHCR_IMAGE:$RELEASE_NAME"
if prompt_confirm "Delete GHCR container image version?" "n"; then
  list_err=$(gh api "orgs/olivetin/packages/container/olivetin/versions" --jq ".[] | select(.metadata.container.tags[]? == \"$RELEASE_NAME\") | .id" 2>&1) || true
  version_id=$(echo "$list_err" | head -1)
  if [[ -z "$version_id" || ! "$version_id" =~ ^[0-9]+$ ]]; then
    log "Could not resolve GHCR version for tag '$RELEASE_NAME' (need read:packages scope, or tag may not exist)." >&2
    if [[ "$list_err" == *"message"* ]]; then
      msg=$(echo "$list_err" | sed -n 's/.*"message":"\([^"]*\)".*/\1/p' | head -1)
      [[ -n "$msg" ]] && log "  $msg" >&2
    fi
  else
    if err=$(gh api -X DELETE "orgs/olivetin/packages/container/olivetin/versions/$version_id" 2>&1); then
      log "Deleted GHCR version (id: $version_id)."
    else
      log "Failed to delete GHCR version:" >&2
      echo "$err" | sed 's/^/[unrelease]   /' >&2
    fi
  fi
else
  log "Skipped GHCR."
fi
echo

# --- Docker Hub ---
log "Step 4: Delete Docker Hub image tag $DOCKERHUB_IMAGE:$RELEASE_NAME"
if prompt_confirm "Delete Docker Hub image tag? (requires DOCKERHUB_TOKEN)" "n"; then
  if [[ -z "${DOCKERHUB_TOKEN:-}" ]]; then
    log "DOCKERHUB_TOKEN is not set. Get a token from https://hub.docker.com/settings/security and run: DOCKERHUB_TOKEN=xxx $0 $RELEASE_NAME" >&2
    log "Skipped Docker Hub."
  else
    status=$(curl -s -o /dev/null -w "%{http_code}" -X DELETE \
      -H "Authorization: Bearer $DOCKERHUB_TOKEN" \
      "https://hub.docker.com/v2/repositories/$DOCKERHUB_IMAGE/tags/$RELEASE_NAME/")
    if [[ "$status" == "204" ]]; then
      log "Deleted Docker Hub tag."
    else
      log "Docker Hub delete returned HTTP $status (tag may not exist or token invalid)." >&2
    fi
  fi
else
  log "Skipped Docker Hub."
fi

log "Done."
