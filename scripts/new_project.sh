#!/usr/bin/env bash
set -euo pipefail

# new_project.sh — Bootstrap a new project from gin_template (clone or in-place)
#
# Usage examples:
#   1) Clone remote template into target dir and bootstrap:
#      scripts/new_project.sh --module github.com/you/my_service --dir /abs/path/to/my_service
#
#   2) Copy from local template directory (no network):
#      scripts/new_project.sh --module github.com/you/my_service \
#        --template /abs/path/to/gin_template --dir /abs/path/to/my_service
#
#   3) Run in an existing cloned template directory (in-place):
#      cd /abs/path/to/gin_template && scripts/new_project.sh --module github.com/you/my_service --in-place
#
# Options:
#   --module    Required. New module path, e.g. github.com/you/my_service
#   --dir       Target directory to create (required unless --in-place)
#   --template  Local template directory to copy from (optional; if omitted we git clone)
#   --skip-git  Skip git init and initial add/commit
#   --skip-tidy Skip running `go mod tidy`
#   --in-place  Operate in current directory instead of creating a new copy

SOURCE_MODULE="github.com/wiidz/gin_template"
DEFAULT_GIT_URL="https://github.com/wiidz/gin_template.git"

MODULE=""
TARGET_DIR=""
TEMPLATE_DIR=""
SKIP_GIT="false"
SKIP_TIDY="false"
IN_PLACE="false"

print_usage() {
  cat <<EOF
Usage:
  $(basename "$0") --module <github.com/you/my_service> [--dir <path> | --in-place] [--template <path>] [--skip-git] [--skip-tidy]

Examples:
  $(basename "$0") --module github.com/you/my_service --dir /abs/path/to/my_service
  $(basename "$0") --module github.com/you/my_service --template /abs/path/to/gin_template --dir /abs/path/to/my_service
  $(basename "$0") --module github.com/you/my_service --in-place
EOF
}

while [[ $# -gt 0 ]]; do
  case "$1" in
    --module)
      MODULE="${2:-}"; shift 2;;
    --dir)
      TARGET_DIR="${2:-}"; shift 2;;
    --template)
      TEMPLATE_DIR="${2:-}"; shift 2;;
    --skip-git)
      SKIP_GIT="true"; shift;;
    --skip-tidy)
      SKIP_TIDY="true"; shift;;
    --in-place)
      IN_PLACE="true"; shift;;
    -h|--help)
      print_usage; exit 0;;
    *)
      echo "Unknown argument: $1" >&2
      print_usage; exit 1;;
  esac
done

if [[ -z "$MODULE" ]]; then
  echo "Error: --module is required" >&2
  print_usage; exit 1
fi

if [[ "$IN_PLACE" == "true" ]]; then
  TARGET_DIR="$(pwd)"
else
  if [[ -z "$TARGET_DIR" ]]; then
    echo "Error: --dir is required unless using --in-place" >&2
    print_usage; exit 1
  fi
  if [[ -e "$TARGET_DIR" ]]; then
    echo "Error: target directory already exists: $TARGET_DIR" >&2
    exit 1
  fi
  mkdir -p "$TARGET_DIR"
fi

looks_like_template_root() {
  local path="$1"
  [[ -f "$path/go.mod" ]] && grep -q "^module ${SOURCE_MODULE}$" "$path/go.mod"
}

copy_from_local_template() {
  local src="$1" dest="$2"
  if command -v rsync >/dev/null 2>&1; then
    rsync -a --delete \
      --exclude ".git" \
      --exclude "tmp" \
      --exclude "cmd/gt" \
      --exclude "scripts/new_project.sh" \
      "$src/" "$dest/"
  else
    # Fallback using tar (BSD tar supports --exclude)
    (cd "$src" && tar --exclude .git --exclude tmp --exclude cmd/gt --exclude scripts/new_project.sh -cf - .) | (cd "$dest" && tar -xf -)
  fi
}

clone_remote_template() {
  local url="$1" dest="$2"
  if ! command -v git >/dev/null 2>&1; then
    echo "Error: git not found; install git or use --template for local copy" >&2
    exit 1
  fi
  git clone --depth=1 "$url" "$dest"
}

detect_sed_inplace() {
  if command -v gsed >/dev/null 2>&1; then
    echo "gsed -i"
  else
    # macOS/BSD sed requires an empty string after -i
    echo "sed -i ''"
  fi
}

SED_INPLACE_CMD=( $(detect_sed_inplace) )

bootstrap_in_dir() {
  local dir="$1"

  # Remove template-only content
  rm -rf "$dir/.git" "$dir/tmp" "$dir/cmd/gt" "$dir/scripts/new_project.sh" || true

  # Update go.mod module line (exact match)
  if [[ ! -f "$dir/go.mod" ]]; then
    echo "Error: go.mod not found in $dir" >&2
    exit 1
  fi
  (cd "$dir" && "${SED_INPLACE_CMD[@]}" "s|^module ${SOURCE_MODULE}$|module ${MODULE}|" go.mod)

  # Replace import paths across files
  # shellcheck disable=SC2016
  (cd "$dir" && find . -type f \
    \( -name '*.go' -o -name 'go.mod' -o -name '*.sum' -o -name '*.yaml' -o -name '*.toml' -o -name 'Makefile' \) \
    -not -path './.git/*' -not -path './cmd/gt/*' -not -path './tmp/*' -print0 \
    | xargs -0 ${SED_INPLACE_CMD[*]} -e "s|${SOURCE_MODULE}|${MODULE}|g")

  # Init git
  if [[ "$SKIP_GIT" != "true" ]]; then
    if command -v git >/dev/null 2>&1; then
      (cd "$dir" && git init && git add . && git commit -m "bootstrap from template") || true
    else
      echo "git not found; skipping git init" >&2
    fi
  fi

  # go mod tidy
  if [[ "$SKIP_TIDY" != "true" ]]; then
    if command -v go >/dev/null 2>&1; then
      (cd "$dir" && go mod tidy) || echo "Warning: go mod tidy failed" >&2
    else
      echo "Go toolchain not found; skipping go mod tidy" >&2
    fi
  fi
}

# 1) Prepare project directory content
if [[ "$IN_PLACE" == "true" ]]; then
  if ! looks_like_template_root "$TARGET_DIR"; then
    echo "Error: current directory doesn't look like the template root (go.mod module mismatch)" >&2
    echo "Hint: run without --in-place to create a copy, or specify --template to copy from local template." >&2
    exit 1
  fi
else
  if [[ -n "$TEMPLATE_DIR" ]]; then
    if ! looks_like_template_root "$TEMPLATE_DIR"; then
      echo "Error: --template path doesn't look like gin_template: $TEMPLATE_DIR" >&2
      exit 1
    fi
    copy_from_local_template "$TEMPLATE_DIR" "$TARGET_DIR"
  else
    clone_remote_template "$DEFAULT_GIT_URL" "$TARGET_DIR"
  fi
fi

# 2) Bootstrap (strip template + set module + replace imports + init)
bootstrap_in_dir "$TARGET_DIR"

cat <<EONEXT

✅ Project created at: $TARGET_DIR

Next steps:
  cd "$TARGET_DIR"
  make run        # or: make tidy && make run

Tips:
  If you rely on private repos under github.com/wiidz/*, you may need:
    go env -w 'GOPRIVATE=github.com/wiidz/*'

EONEXT


