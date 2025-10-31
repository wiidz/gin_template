#!/usr/bin/env bash

set -euo pipefail

SOURCE_MODULE="gin_template"

usage() {
  cat <<EOF
Bootstrap a new Go project from this template.

Usage: $(basename "$0") <project-name> [--module MODULE_PATH] [--dir TARGET_DIR]
                    [--skip-git] [--skip-tidy]

Arguments:
  <project-name>        Name of the new project directory (required).

Options:
  --module MODULE_PATH  Module path to replace '${SOURCE_MODULE}'. Defaults to <project-name>.
  --dir TARGET_DIR      Target directory for the new project. Defaults to a sibling directory
                        next to this template, named after <project-name>.
  --skip-git            Do not run 'git init' in the new project directory.
  --skip-tidy           Do not run 'go mod tidy' after bootstrapping.
  -h, --help            Show this help message and exit.

Example:
  $(basename "$0") my_service --module github.com/you/my_service
EOF
}

PROJECT_NAME=""
TARGET_DIR=""
MODULE_PATH=""
INIT_GIT=1
RUN_TIDY=1

while [[ $# -gt 0 ]]; do
  case "$1" in
    --module)
      [[ $# -ge 2 ]] || { echo "Error: missing value for --module" >&2; exit 1; }
      MODULE_PATH="$2"
      shift 2
      ;;
    --dir)
      [[ $# -ge 2 ]] || { echo "Error: missing value for --dir" >&2; exit 1; }
      TARGET_DIR="$2"
      shift 2
      ;;
    --skip-git)
      INIT_GIT=0
      shift
      ;;
    --skip-tidy)
      RUN_TIDY=0
      shift
      ;;
    -h|--help)
      usage
      exit 0
      ;;
    --*)
      echo "Error: unknown option '$1'" >&2
      usage >&2
      exit 1
      ;;
    *)
      if [[ -z "$PROJECT_NAME" ]]; then
        PROJECT_NAME="$1"
        shift
      else
        echo "Error: unexpected argument '$1'" >&2
        usage >&2
        exit 1
      fi
      ;;
  esac
done

if [[ -z "$PROJECT_NAME" ]]; then
  echo "Error: <project-name> is required." >&2
  usage >&2
  exit 1
fi

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
TEMPLATE_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
DEFAULT_TARGET_BASE="$(dirname "$TEMPLATE_ROOT")"

if [[ -z "$MODULE_PATH" ]]; then
  MODULE_PATH="$PROJECT_NAME"
fi

abs_path() {
  local input_path="$1"
  if [[ "$input_path" == /* ]]; then
    printf '%s\n' "$input_path"
  else
    printf '%s/%s\n' "$(pwd)" "$input_path"
  fi
}

if [[ -z "$TARGET_DIR" ]]; then
  TARGET_DIR="$DEFAULT_TARGET_BASE/$PROJECT_NAME"
else
  TARGET_DIR="$(abs_path "$TARGET_DIR")"
fi

if [[ -e "$TARGET_DIR" ]]; then
  echo "Error: target directory '$TARGET_DIR' already exists." >&2
  exit 1
fi

if ! command -v rsync >/dev/null 2>&1; then
  echo "Error: 'rsync' is required but not found in PATH." >&2
  exit 1
fi

echo "Creating project at $TARGET_DIR"
mkdir -p "$TARGET_DIR"

RSYNC_EXCLUDES=(
  "--exclude" ".git/"
  "--exclude" ".idea/"
  "--exclude" ".vscode/"
  "--exclude" "tmp/server"
  "--exclude" "tmp/build-errors.log"
  "--exclude" "scripts/new_project.sh"
)

rsync -a "${RSYNC_EXCLUDES[@]}" "$TEMPLATE_ROOT/" "$TARGET_DIR/"

export SOURCE_MODULE
export MODULE_PATH

while IFS= read -r -d '' file; do
  perl -0pi -e 's/\Q$ENV{SOURCE_MODULE}\E/$ENV{MODULE_PATH}/g' "$file"
done < <(
  find "$TARGET_DIR" \
    \( -path '*/.git/*' -o -path '*/vendor/*' \) -prune -o \
    -type f \
    \( -name '*.go' -o -name '*.mod' -o -name '*.sum' -o -name '*.yaml' -o -name '*.yml' \
       -o -name '*.json' -o -name '*.env' -o -name '*.toml' -o -name '*.txt' \
       -o -name 'Makefile' -o -name '*.md' \) -print0
)

if [[ $INIT_GIT -eq 1 ]]; then
  if command -v git >/dev/null 2>&1; then
    (cd "$TARGET_DIR" && git init >/dev/null && git add . >/dev/null)
    echo "Initialized empty Git repository in $TARGET_DIR/.git"
  else
    echo "Warning: git not found; skipping Git initialization." >&2
  fi
else
  echo "Skipped git init (per --skip-git)."
fi

if [[ $RUN_TIDY -eq 1 ]]; then
  if command -v go >/dev/null 2>&1; then
    if (cd "$TARGET_DIR" && go mod tidy); then
      echo "Ran go mod tidy."
    else
      echo "Warning: go mod tidy failed. Please inspect the output above." >&2
    fi
  else
    echo "Warning: Go toolchain not found; skipping go mod tidy." >&2
  fi
else
  echo "Skipped go mod tidy (per --skip-tidy)."
fi

cat <<EOF

Done! Project '${PROJECT_NAME}' created at:
  $TARGET_DIR

Next steps:
  1. Update go.mod module path if needed: go mod edit -module ${MODULE_PATH}
  2. Review configuration files and README for project-specific updates.
EOF

