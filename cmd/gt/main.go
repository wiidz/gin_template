package main

import (
	"errors"
	"flag"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

const (
	sourceModule    = "gin_template"
	envTemplateRoot = "GIN_TEMPLATE_ROOT"
)

type newConfig struct {
	projectName string
	moduleName  string
	targetDir   string
	templateDir string
	skipGit     bool
	skipTidy    bool
}

func main() {
	if len(os.Args) < 2 {
		usage()
		os.Exit(1)
	}

	switch os.Args[1] {
	case "new":
		if err := handleNew(os.Args[2:]); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	case "help", "-h", "--help":
		usage()
	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n\n", os.Args[1])
		usage()
		os.Exit(1)
	}
}

func usage() {
	fmt.Println(`gt - Gin Template project generator

Usage:
  gt new <project-name> [flags]

Flags:
  --module <path>     Override module path (default: <project-name>)
  --dir <path>        Target directory to create project in (default: sibling to template root)
  --template <path>   Template directory (default: detected relative to binary or $GIN_TEMPLATE_ROOT)
  --skip-git          Do not run git init
  --skip-tidy         Do not run go mod tidy
  -h, --help          Show this help message`)
}

func handleNew(args []string) error {
	fs := flag.NewFlagSet("new", flag.ContinueOnError)
	module := fs.String("module", "", "module path for the new project")
	target := fs.String("dir", "", "target directory for the new project")
	template := fs.String("template", "", "template directory (defaults to detection)")
	skipGit := fs.Bool("skip-git", false, "skip git init")
	skipTidy := fs.Bool("skip-tidy", false, "skip go mod tidy")
	fs.Usage = func() {
		fmt.Fprintln(os.Stderr, "Usage: gt new <project-name> [flags]")
		fs.PrintDefaults()
	}

	if err := fs.Parse(args); err != nil {
		return err
	}

	if fs.NArg() != 1 {
		return errors.New("project name is required")
	}

	projectName := fs.Arg(0)
	moduleName := *module
	if moduleName == "" {
		moduleName = projectName
	}

	templateDir, err := determineTemplateDir(*template)
	if err != nil {
		return fmt.Errorf("determine template root: %w", err)
	}

	targetDir, err := determineTargetDir(*target, templateDir, projectName)
	if err != nil {
		return err
	}

	cfg := newConfig{
		projectName: projectName,
		moduleName:  moduleName,
		targetDir:   targetDir,
		templateDir: templateDir,
		skipGit:     *skipGit,
		skipTidy:    *skipTidy,
	}

	if err := createProject(cfg); err != nil {
		return err
	}

	fmt.Printf("\n✅ Project %s created at %s\n", projectName, targetDir)
	fmt.Println("Next steps:")
	fmt.Printf("  cd %s\n", targetDir)
	if !cfg.skipGit {
		fmt.Println("  git commit -m 'bootstrap from gin_template'")
	}
	fmt.Println("  Review configs and README for project-specific tweaks")
	return nil
}

func determineTemplateDir(flagValue string) (string, error) {
	if flagValue != "" {
		return absolutePath(flagValue)
	}

	if env := os.Getenv(envTemplateRoot); env != "" {
		return absolutePath(env)
	}

	if execPath, err := os.Executable(); err == nil {
		dir := filepath.Dir(execPath)
		candidate := filepath.Clean(filepath.Join(dir, "..", ".."))
		if looksLikeTemplateRoot(candidate) {
			return candidate, nil
		}
	}

	if cwd, err := os.Getwd(); err == nil {
		if looksLikeTemplateRoot(cwd) {
			return cwd, nil
		}
	}

	return "", errors.New("unable to detect template directory; specify with --template or set GIN_TEMPLATE_ROOT")
}

func determineTargetDir(flagValue, templateDir, projectName string) (string, error) {
	if flagValue != "" {
		path, err := absolutePath(flagValue)
		if err != nil {
			return "", err
		}
		if exists(path) {
			return "", fmt.Errorf("target directory %s already exists", path)
		}
		return path, nil
	}

	parent := filepath.Dir(templateDir)
	target := filepath.Join(parent, projectName)
	if exists(target) {
		return "", fmt.Errorf("target directory %s already exists", target)
	}
	return target, nil
}

func createProject(cfg newConfig) error {
	if err := os.MkdirAll(cfg.targetDir, 0o755); err != nil {
		return fmt.Errorf("create target dir: %w", err)
	}

	if err := copyTemplate(cfg); err != nil {
		return err
	}

	if !cfg.skipGit {
		if err := initGit(cfg.targetDir); err != nil {
			return err
		}
	}

	if !cfg.skipTidy {
		if err := runGoModTidy(cfg.targetDir); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: go mod tidy failed: %v\n", err)
		}
	}

	return nil
}

func copyTemplate(cfg newConfig) error {
	skipDirs := map[string]struct{}{
		".git":   {},
		"tmp":    {},
		"cmd/gt": {},
	}

	skipFiles := map[string]struct{}{
		"scripts/new_project.sh": {},
	}

	return filepath.WalkDir(cfg.templateDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		rel, err := filepath.Rel(cfg.templateDir, path)
		if err != nil {
			return err
		}

		if rel == "." {
			return nil
		}

		// Skip directories
		if d.IsDir() {
			if _, ok := skipDirs[normPath(rel)]; ok {
				return filepath.SkipDir
			}
			newDir := replaceInPath(filepath.Join(cfg.targetDir, rel), cfg.projectName)
			return os.MkdirAll(newDir, 0o755)
		}

		if _, ok := skipFiles[normPath(rel)]; ok {
			return nil
		}

		destPath := replaceInPath(filepath.Join(cfg.targetDir, rel), cfg.projectName)

		if err := os.MkdirAll(filepath.Dir(destPath), 0o755); err != nil {
			return err
		}

		data, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("read %s: %w", path, err)
		}

		content := strings.ReplaceAll(string(data), sourceModule, cfg.moduleName)

		if err := os.WriteFile(destPath, []byte(content), fileMode(d)); err != nil {
			return fmt.Errorf("write %s: %w", destPath, err)
		}

		if err := os.Chmod(destPath, fileMode(d)); err != nil {
			return err
		}

		return nil
	})
}

func replaceInPath(path, projectName string) string {
	return strings.ReplaceAll(path, sourceModule, projectName)
}

func fileMode(d fs.DirEntry) fs.FileMode {
	if info, err := d.Info(); err == nil {
		return info.Mode()
	}
	return 0o644
}

func initGit(targetDir string) error {
	if _, err := exec.LookPath("git"); err != nil {
		fmt.Fprintln(os.Stderr, "git not found in PATH; skipping git init")
		return nil
	}

	cmd := exec.Command("git", "init")
	cmd.Dir = targetDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("git init: %w", err)
	}

	add := exec.Command("git", "add", ".")
	add.Dir = targetDir
	add.Stdout = os.Stdout
	add.Stderr = os.Stderr
	if err := add.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: git add failed: %v\n", err)
	}
	return nil
}

func runGoModTidy(targetDir string) error {
	if _, err := exec.LookPath("go"); err != nil {
		fmt.Fprintln(os.Stderr, "Go toolchain not found; skipping go mod tidy")
		return nil
	}

	cmd := exec.Command("go", "mod", "tidy")
	cmd.Dir = targetDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func absolutePath(path string) (string, error) {
	if filepath.IsAbs(path) {
		return path, nil
	}
	cwd, err := os.Getwd()
	if err != nil {
		return "", err
	}
	return filepath.Join(cwd, path), nil
}

func looksLikeTemplateRoot(path string) bool {
	goModPath := filepath.Join(path, "go.mod")
	data, err := os.ReadFile(goModPath)
	if err != nil {
		return false
	}
	return strings.Contains(string(data), "module "+sourceModule)
}

func exists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func normPath(path string) string {
	return filepath.ToSlash(path)
}
