package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	gocodegen "github.com/oboard/rune-lang/internal/codegen/go"
	"github.com/oboard/rune-lang/internal/compiler"
	runefmt "github.com/oboard/rune-lang/internal/format"
	"github.com/oboard/rune-lang/internal/lsp"
	"github.com/oboard/rune-lang/internal/parser"
	"github.com/oboard/rune-lang/internal/repl"
	"github.com/oboard/rune-lang/internal/tester"
)

var backend string

func main() {
	if err := rootCmd().Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func rootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "rune",
		Short: "Rune language toolchain",
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			return validateBackend(backend)
		},
	}
	cmd.PersistentFlags().StringVar(&backend, "backend", "go", "target backend: go or ts")
	cmd.AddCommand(runCmd(), buildCmd(), goCmd(), tsCmd(), checkCmd(), testCmd(), fmtCmd(), replCmd(), lspCmd())
	return cmd
}

func runCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "run <file.rn> [args...]",
		Short: "Compile and run a Rune program",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			switch backend {
			case "go":
				goFile, cleanup, err := compileGoToTemp(args[0])
				if err != nil {
					return err
				}
				defer cleanup()

				goArgs := append([]string{"run", goFile}, args[1:]...)
				run := exec.Command("go", goArgs...)
				run.Stdout = os.Stdout
				run.Stderr = os.Stderr
				run.Stdin = os.Stdin
				return run.Run()
			case "ts":
				tsFile, cleanup, err := compileTypeScriptToTemp(args[0])
				if err != nil {
					return err
				}
				defer cleanup()
				run, err := typeScriptRuntimeCommand(tsFile, args[1:])
				if err != nil {
					return err
				}
				run.Stdout = os.Stdout
				run.Stderr = os.Stderr
				run.Stdin = os.Stdin
				return run.Run()
			default:
				return validateBackend(backend)
			}
		},
	}
}

func buildCmd() *cobra.Command {
	var output string
	var target string
	cmd := &cobra.Command{
		Use:   "build <file.rn>",
		Short: "Compile a Rune program to an executable",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if backend != "go" {
				return fmt.Errorf("rune build only supports --backend go")
			}
			goFile, cleanup, err := compileGoToTemp(args[0])
			if err != nil {
				return err
			}
			defer cleanup()

			env, err := buildEnv(target)
			if err != nil {
				return err
			}

			out := output
			if out == "" {
				out = defaultBinaryName(args[0])
			}
			build := exec.Command("go", "build", "-o", out, goFile)
			build.Env = env
			build.Stdout = os.Stdout
			build.Stderr = os.Stderr
			build.Stdin = os.Stdin
			return build.Run()
		},
	}
	cmd.Flags().StringVarP(&output, "output", "o", "", "output executable path")
	cmd.Flags().StringVar(&target, "target", "", "build target as GOOS-GOARCH, for example linux-amd64")
	return cmd
}

func tsCmd() *cobra.Command {
	var output string
	cmd := &cobra.Command{
		Use:   "ts <file.rn>",
		Short: "Compile a Rune program to TypeScript",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			src, diags := compiler.GenerateTypeScriptFile(args[0])
			if len(diags) > 0 {
				printDiagnostics(args[0], diags)
				return fmt.Errorf("compile failed")
			}
			if output == "" {
				fmt.Fprint(cmd.OutOrStdout(), src)
				return nil
			}
			return os.WriteFile(output, []byte(src), 0o644)
		},
	}
	cmd.Flags().StringVarP(&output, "output", "o", "", "output TypeScript path")
	return cmd
}

func goCmd() *cobra.Command {
	var output string
	cmd := &cobra.Command{
		Use:   "go <file.rn>",
		Short: "Compile a Rune program to Go",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			src, diags := compiler.GenerateGoFile(args[0])
			if len(diags) > 0 {
				printDiagnostics(args[0], diags)
				return fmt.Errorf("compile failed")
			}
			if output == "" {
				fmt.Fprint(cmd.OutOrStdout(), src)
				return nil
			}
			return os.WriteFile(output, []byte(src), 0o644)
		},
	}
	cmd.Flags().StringVarP(&output, "output", "o", "", "output Go path")
	return cmd
}

func checkCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "check <file.rn>",
		Short: "Parse and type-check a Rune program",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			_, diags := compiler.AnalyzeFile(args[0])
			if len(diags) > 0 {
				printDiagnostics(args[0], diags)
				return fmt.Errorf("check failed")
			}
			fmt.Fprintf(cmd.OutOrStdout(), "ok %s\n", args[0])
			return nil
		},
	}
}

func testCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "test [path] [pattern]",
		Short: "Run Rune tests",
		Args:  cobra.MaximumNArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			path := "tests"
			pattern := ""
			if len(args) > 0 {
				path = args[0]
			}
			if len(args) > 1 {
				pattern = args[1]
			}
			_, err := tester.Run(path, pattern, cmd.OutOrStdout())
			return err
		},
	}
}

func fmtCmd() *cobra.Command {
	var checkOnly bool
	var stdout bool
	cmd := &cobra.Command{
		Use:   "fmt <file.rn>",
		Short: "Format a Rune source file",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			original, err := os.ReadFile(args[0])
			if err != nil {
				return err
			}
			file, errs := parser.Parse(string(original))
			if len(errs) > 0 {
				printDiagnostics(args[0], parseDiagnostics(errs))
				return fmt.Errorf("format failed")
			}
			formatted := runefmt.Source(file, string(original))
			if stdout {
				fmt.Fprint(cmd.OutOrStdout(), formatted)
				return nil
			}
			if string(original) == formatted {
				return nil
			}
			if checkOnly {
				return fmt.Errorf("%s is not formatted", args[0])
			}
			return os.WriteFile(args[0], []byte(formatted), 0o644)
		},
	}
	cmd.Flags().BoolVar(&checkOnly, "check", false, "fail if the file is not formatted")
	cmd.Flags().BoolVar(&stdout, "stdout", false, "write formatted source to stdout")
	return cmd
}

func lspCmd() *cobra.Command {
	var stdio bool
	cmd := &cobra.Command{
		Use:   "lsp",
		Short: "Start the Rune language server",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return lsp.Serve(os.Stdin, os.Stdout)
		},
	}
	cmd.Flags().BoolVar(&stdio, "stdio", true, "serve LSP over stdin/stdout")
	return cmd
}

func replCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "repl",
		Short: "Start the Rune REPL",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return repl.Serve(os.Stdin, os.Stdout)
		},
	}
}

func compileGoToTemp(path string) (string, func(), error) {
	prog, diags := compiler.AnalyzeFile(path)
	if len(diags) > 0 {
		printDiagnostics(path, diags)
		return "", func() {}, fmt.Errorf("compile failed")
	}
	src, err := gocodegen.GenerateIR(prog.IR)
	if err != nil {
		return "", func() {}, err
	}
	dir, err := os.MkdirTemp("", "rune-*")
	if err != nil {
		return "", func() {}, err
	}
	cleanup := func() {
		_ = os.RemoveAll(dir)
	}
	goFile := filepath.Join(dir, "main.go")
	if err := os.WriteFile(goFile, []byte(src), 0o644); err != nil {
		cleanup()
		return "", func() {}, err
	}
	return goFile, cleanup, nil
}

func compileTypeScriptToTemp(path string) (string, func(), error) {
	src, diags := compiler.GenerateTypeScriptFile(path)
	if len(diags) > 0 {
		printDiagnostics(path, diags)
		return "", func() {}, fmt.Errorf("compile failed")
	}
	dir, err := os.MkdirTemp("", "rune-*")
	if err != nil {
		return "", func() {}, err
	}
	cleanup := func() {
		_ = os.RemoveAll(dir)
	}
	tsFile := filepath.Join(dir, "main.ts")
	src += "\nif (typeof __main === \"function\") {\n  __main();\n}\n"
	if err := os.WriteFile(tsFile, []byte(src), 0o644); err != nil {
		cleanup()
		return "", func() {}, err
	}
	return tsFile, cleanup, nil
}

func typeScriptRuntimeCommand(path string, args []string) (*exec.Cmd, error) {
	if runner, err := exec.LookPath("bun"); err == nil {
		return exec.Command(runner, append([]string{path}, args...)...), nil
	}
	if runner, err := exec.LookPath("deno"); err == nil {
		return exec.Command(runner, append([]string{"run", path}, args...)...), nil
	}
	if runner, err := exec.LookPath("node"); err == nil {
		return exec.Command(runner, append([]string{path}, args...)...), nil
	}
	return nil, fmt.Errorf("TypeScript backend requires bun, deno, or node")
}

func buildEnv(target string) ([]string, error) {
	env := os.Environ()
	if target == "" {
		return env, nil
	}
	goos, goarch, err := parseTarget(target)
	if err != nil {
		return nil, err
	}
	return append(env, "GOOS="+goos, "GOARCH="+goarch), nil
}

func parseTarget(target string) (string, string, error) {
	parts := strings.Split(target, "-")
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "", "", fmt.Errorf("invalid target %q, expected GOOS-GOARCH", target)
	}
	return parts[0], parts[1], nil
}

func validateBackend(value string) error {
	switch value {
	case "go", "ts":
		return nil
	default:
		return fmt.Errorf("invalid backend %q, expected go or ts", value)
	}
}

func printDiagnostics(path string, diags []compiler.Diagnostic) {
	for _, diag := range diags {
		if diag.Pos.Line > 0 {
			fmt.Fprintf(os.Stderr, "%s:%d:%d: %s\n", path, diag.Pos.Line, diag.Pos.Column, diag.Message)
		} else {
			fmt.Fprintf(os.Stderr, "%s: %s\n", path, diag.Message)
		}
	}
}

func parseDiagnostics(errs []parser.Error) []compiler.Diagnostic {
	diags := make([]compiler.Diagnostic, 0, len(errs))
	for _, err := range errs {
		diags = append(diags, compiler.Diagnostic{
			Message: err.Message,
			Pos:     err.Pos,
		})
	}
	return diags
}

func defaultBinaryName(path string) string {
	base := filepath.Base(path)
	ext := filepath.Ext(base)
	if ext != "" {
		base = base[:len(base)-len(ext)]
	}
	if base == "" {
		return "main"
	}
	return base
}
