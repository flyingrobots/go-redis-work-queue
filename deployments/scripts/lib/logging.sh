#!/bin/bash
# Shared logging helpers for deployment scripts.

if [[ -n "${__GRWQ_LOGGING_SH:-}" ]]; then
  return 0 2>/dev/null || exit 0
fi
__GRWQ_LOGGING_SH=1

# ANSI colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

_timestamp() {
  date +'%Y-%m-%d %H:%M:%S'
}

log_info() {
  echo -e "${GREEN}[INFO $(_timestamp)]${NC} $*"
}

log_warn() {
  echo -e "${YELLOW}[WARN $(_timestamp)]${NC} $*"
}

log_error() {
  echo -e "${RED}[ERROR $(_timestamp)]${NC} $*" >&2
}

log_debug() {
  echo -e "${BLUE}[DEBUG $(_timestamp)]${NC} $*"
}

# Compatibility wrappers for legacy helper names.
log() { log_info "$*"; }
info() { log_info "$*"; }
warn() { log_warn "$*"; }
error() { log_error "$*"; }
