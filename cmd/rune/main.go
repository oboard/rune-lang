package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/spf13/cobra"

	gocodegen "github.com/oboard/rune-lang/internal/codegen/go"
	"github.com/oboard/rune-lang/internal/compiler"
	runefmt "github.com/oboard/rune-lang/internal/format"
	"github.com/oboard/rune-lang/internal/lsp"
	"github.com/oboard/rune-lang/internal/repl"
)

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
	}
	cmd.AddCommand(runCmd(), buildCmd(), checkCmd(), fmtCmd(), replCmd(), lspCmd())
	return cmd
}

func runCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "run <file.rn> [args...]",
		Short: "Compile and run a Rune program",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			goFile, cleanup, err := compileToTemp(args[0])
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
		},
	}
}

func buildCmd() *cobra.Command {
	var output string
	cmd := &cobra.Command{
		Use:   "build <file.rn>",
		Short: "Compile a Rune program to an executable",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			goFile, cleanup, err := compileToTemp(args[0])
			if err != nil {
				return err
			}
			defer cleanup()

			out := output
			if out == "" {
				out = defaultBinaryName(args[0])
			}
			build := exec.Command("go", "build", "-o", out, goFile)
			build.Stdout = os.Stdout
			build.Stderr = os.Stderr
			build.Stdin = os.Stdin
			return build.Run()
		},
	}
	cmd.Flags().StringVarP(&output, "output", "o", "", "output executable path")
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

func fmtCmd() *cobra.Command {
	var checkOnly bool
	var stdout bool
	cmd := &cobra.Command{
		Use:   "fmt <file.rn>",
		Short: "Format a Rune source file",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			prog, diags := compiler.AnalyzeFile(args[0])
			if len(diags) > 0 {
				printDiagnostics(args[0], diags)
				return fmt.Errorf("format failed")
			}
			formatted := runefmt.File(prog.File)
			if stdout {
				fmt.Fprint(cmd.OutOrStdout(), formatted)
				return nil
			}
			original, err := os.ReadFile(args[0])
			if err != nil {
				return err
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
	return &cobra.Command{
		Use:   "lsp",
		Short: "Start the Rune language server",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return lsp.Serve(os.Stdin, os.Stdout)
		},
	}
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

func compileToTemp(path string) (string, func(), error) {
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

func printDiagnostics(path string, diags []compiler.Diagnostic) {
	for _, diag := range diags {
		if diag.Pos.Line > 0 {
			fmt.Fprintf(os.Stderr, "%s:%d:%d: %s\n", path, diag.Pos.Line, diag.Pos.Column, diag.Message)
		} else {
			fmt.Fprintf(os.Stderr, "%s: %s\n", path, diag.Message)
		}
	}
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
