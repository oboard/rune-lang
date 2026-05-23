package tester

import (
	"bytes"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"time"

	"github.com/oboard/rune-lang/internal/compiler"
	"github.com/oboard/rune-lang/internal/interpreter"
)

type Summary struct {
	Passed  int
	Failed  int
	Skipped int
	Files   int
}

func Run(path string, pattern string, out io.Writer) (Summary, error) {
	if path == "" {
		path = "tests"
	}
	files, err := runeFiles(path)
	if err != nil {
		return Summary{}, err
	}
	var match *regexp.Regexp
	if pattern != "" {
		match, err = regexp.Compile(pattern)
		if err != nil {
			return Summary{}, fmt.Errorf("invalid test pattern: %w", err)
		}
	}

	start := time.Now()
	summary := Summary{Files: len(files)}
	for _, file := range files {
		prog, diags := compiler.AnalyzeFile(file)
		if len(diags) > 0 {
			summary.Failed++
			fmt.Fprintf(out, "=== RUN %s\n", file)
			fmt.Fprintf(out, "--- FAIL %s (compile)\n", file)
			for _, diag := range diags {
				if diag.Pos.Line > 0 {
					fmt.Fprintf(out, "    %s:%d:%d: %s\n", file, diag.Pos.Line, diag.Pos.Column, diag.Message)
				} else {
					fmt.Fprintf(out, "    %s: %s\n", file, diag.Message)
				}
			}
			continue
		}
		for _, test := range prog.IR.Tests {
			fullName := file + " " + test.Name
			if match != nil && !match.MatchString(test.Name) && !match.MatchString(fullName) {
				summary.Skipped++
				continue
			}
			fmt.Fprintf(out, "=== RUN %s ? %s\n", file, test.Name)
			testStart := time.Now()
			var testOutput bytes.Buffer
			runner := interpreter.New(prog.IR, interpreter.WithOutput(&testOutput))
			err := runner.RunTest(test)
			elapsed := time.Since(testStart)
			if err != nil {
				summary.Failed++
				fmt.Fprintf(out, "--- FAIL %s ? %s (%s)\n", file, test.Name, elapsed.Round(time.Microsecond))
				if testOutput.Len() > 0 {
					fmt.Fprintf(out, "    output:\n%s", indent(testOutput.String(), "      "))
				}
				fmt.Fprintf(out, "    error: %s\n", err)
				continue
			}
			summary.Passed++
			fmt.Fprintf(out, "--- PASS %s ? %s (%s)\n", file, test.Name, elapsed.Round(time.Microsecond))
			if testOutput.Len() > 0 {
				fmt.Fprintf(out, "    output:\n%s", indent(testOutput.String(), "      "))
			}
		}
	}
	total := summary.Passed + summary.Failed + summary.Skipped
	status := "PASS"
	if summary.Failed > 0 {
		status = "FAIL"
	}
	fmt.Fprintf(out, "%s %d tests, %d passed, %d failed, %d skipped in %s\n", status, total, summary.Passed, summary.Failed, summary.Skipped, time.Since(start).Round(time.Millisecond))
	if summary.Failed > 0 {
		return summary, fmt.Errorf("test failed")
	}
	if total == 0 {
		return summary, fmt.Errorf("no tests matched")
	}
	return summary, nil
}

func runeFiles(path string) ([]string, error) {
	info, err := fs.Stat(osDirFS{}, path)
	if err != nil {
		return nil, err
	}
	if !info.IsDir() {
		return []string{path}, nil
	}
	var files []string
	err = filepath.WalkDir(path, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() {
			return nil
		}
		if filepath.Ext(path) == ".rn" {
			files = append(files, path)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	sort.Strings(files)
	return files, nil
}

func indent(s string, prefix string) string {
	var b bytes.Buffer
	for _, ch := range s {
		if b.Len() == 0 || b.Bytes()[b.Len()-1] == '\n' {
			b.WriteString(prefix)
		}
		b.WriteRune(ch)
	}
	if len(s) > 0 && s[len(s)-1] != '\n' {
		b.WriteByte('\n')
	}
	return b.String()
}

type osDirFS struct{}

func (osDirFS) Open(name string) (fs.File, error) {
	return os.Open(name)
}
