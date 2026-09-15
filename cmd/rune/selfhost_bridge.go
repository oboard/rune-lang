package main

import shared "github.com/oboard/rune-lang/internal/selfhost"

type SourceFile struct {
	path   string
	source string
}
type CompileResult struct {
	ok     bool
	output string
	errors []string
}

func bridgeFiles(files []SourceFile) []shared.SourceFileInput {
	out := make([]shared.SourceFileInput, len(files))
	for i, f := range files {
		out[i] = shared.SourceFileInput{Path: f.path, Source: f.source}
	}
	return out
}
func bridgeResult(r shared.CompileOutput) CompileResult {
	return CompileResult{ok: r.Ok, output: r.Output, errors: r.Errors}
}
func formatSource(source string) string     { return shared.FormatSource(source) }
func compileGo(source string) CompileResult { return bridgeResult(shared.CompileGo(source)) }
func compileGoFiles(files []SourceFile) CompileResult {
	return bridgeResult(shared.CompileGoFiles(bridgeFiles(files)))
}
func compileTypeScriptFiles(files []SourceFile) CompileResult {
	return bridgeResult(shared.CompileTypeScriptFiles(bridgeFiles(files)))
}
func checkSourceWithCore(source, path string, files []SourceFile) CompileResult {
	return bridgeResult(shared.CheckSourceWithCore(source, path, bridgeFiles(files)))
}
func compileMoonBitFiles(files []SourceFile) CompileResult {
	return bridgeResult(shared.CompileMoonBitFiles(bridgeFiles(files)))
}
func compileDeclarations(source string) CompileResult {
	return bridgeResult(shared.CompileDeclarations(source))
}
func compileDeclarationsFiles(files []SourceFile) CompileResult {
	return bridgeResult(shared.CompileDeclarationsFiles(bridgeFiles(files)))
}
