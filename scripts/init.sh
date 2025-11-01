#!/usr/bin/env bash
set -euo pipefail

# Interactive initializer to be run AFTER cloning the template
# Usage:
#   git clone https://github.com/wiidz/gin_template.git my_service
#   cd my_service
#   bash scripts/init.sh

ROOT_DIR="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT_DIR"

SOURCE_MODULE="github.com/wiidz/gin_template"

PROJECT_NAME_DEFAULT="$(basename "$ROOT_DIR")"

echo "=== Gin Template -> New Project (in-place) ==="
read -r -p "项目名 [${PROJECT_NAME_DEFAULT}]: " PROJECT_NAME
PROJECT_NAME=${PROJECT_NAME:-$PROJECT_NAME_DEFAULT}

MODULE_DEFAULT="github.com/you/${PROJECT_NAME}"
read -r -p "Go module (module path) [${MODULE_DEFAULT}]: " MODULE
MODULE=${MODULE:-$MODULE_DEFAULT}

echo
echo "即将进行初始化："
echo "  目录: $ROOT_DIR"
echo "  项目名: $PROJECT_NAME"
echo "  Module: $MODULE"
read -r -p "确认继续? [y/N]: " CONFIRM
if [[ ! "$CONFIRM" =~ ^[Yy]$ ]]; then
  echo "已取消。"
  exit 1
fi

detect_sed_inplace() {
  if command -v gsed >/dev/null 2>&1; then
    echo "gsed -i"
  else
    echo "sed -i ''"
  fi
}

SED_INPLACE_CMD=( $(detect_sed_inplace) )

replace_module_line() {
  if [[ ! -f go.mod ]]; then
    echo "Error: go.mod 不存在，请确认当前目录为模板根目录" >&2
    exit 1
  fi
  "${SED_INPLACE_CMD[@]}" "s|^module ${SOURCE_MODULE}$|module ${MODULE}|" go.mod
}

replace_import_paths() {
  # shellcheck disable=SC2016
  find . -type f \
    \( -name '*.go' -o -name 'go.mod' -o -name '*.sum' -o -name '*.yaml' -o -name '*.toml' -o -name 'Makefile' \) \
    -not -path './.git/*' -not -path './cmd/gt/*' -not -path './tmp/*' \
    -print0 | xargs -0 ${SED_INPLACE_CMD[*]} -e "s|${SOURCE_MODULE}|${MODULE}|g"
}

cleanup_template_assets() {
  rm -rf .git tmp cmd/gt scripts/new_project.sh
}

run_git_init() {
  if command -v git >/dev/null 2>&1; then
    git init >/dev/null 2>&1 || true
    git add . >/dev/null 2>&1 || true
    git commit -m "bootstrap from template" >/dev/null 2>&1 || true
  else
    echo "git 未安装，跳过 git init" >&2
  fi
}

run_go_mod_tidy() {
  if command -v go >/dev/null 2>&1; then
    if ! go mod tidy; then
      echo "Warning: go mod tidy 执行失败" >&2
    fi
  else
    echo "Go 工具链未安装，跳过 go mod tidy" >&2
  fi
}

cleanup_template_assets
replace_module_line
replace_import_paths
run_git_init
run_go_mod_tidy

# 最后清理自身脚本与 scripts 目录（脚本已加载至内存，移除无影响）
rm -f scripts/init.sh 2>/dev/null || true
rmdir scripts 2>/dev/null || true

echo
echo "✅ 初始化完成"
echo "下一步："
echo "  make run   # 或 make tidy && make run"



