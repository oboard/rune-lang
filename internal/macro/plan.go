package macro

import (
	"sort"

	"github.com/oboard/rune-lang/internal/ast"
	"github.com/oboard/rune-lang/internal/checker"
	"github.com/oboard/rune-lang/internal/lexer"
	"github.com/oboard/rune-lang/internal/stdlib"
)

type TargetKind string

const (
	StructTarget     TargetKind = "struct"
	FieldTarget      TargetKind = "field"
	EnumTarget       TargetKind = "enum"
	EnumMemberTarget TargetKind = "enumMember"
	FunctionTarget   TargetKind = "function"
)

type Target struct {
	Kind       TargetKind
	Name       string
	ParentName string
	SourcePath string
	Pos        lexer.Position

	Struct     *ast.StructType
	Field      *ast.Field
	Enum       *ast.EnumType
	EnumMember *ast.EnumMember
	Function   *ast.Function
}

type Invocation struct {
	Order      int
	Target     Target
	Annotation *ast.Annotation
	Macro      *stdlib.Function
	LocalMacro *checker.FuncInfo
}

func Plan(file *ast.File, info *checker.Info) []Invocation {
	if file == nil || info == nil {
		return nil
	}
	targets := collectTargets(file)
	sort.SliceStable(targets, func(i, j int) bool {
		if targets[i].SourcePath != targets[j].SourcePath {
			return targets[i].SourcePath < targets[j].SourcePath
		}
		left := targets[i].Pos.Offset
		right := targets[j].Pos.Offset
		if left != right {
			return left < right
		}
		if targets[i].Pos.Line != targets[j].Pos.Line {
			return targets[i].Pos.Line < targets[j].Pos.Line
		}
		return targets[i].Pos.Column < targets[j].Pos.Column
	})

	var invocations []Invocation
	for _, target := range targets {
		for _, annotation := range targetAnnotations(target) {
			fn := info.ResolvedMacros[annotation]
			local := info.ResolvedMacroFunctions[annotation]
			if fn == nil && local == nil {
				continue
			}
			invocations = append(invocations, Invocation{
				Order:      len(invocations),
				Target:     target,
				Annotation: annotation,
				Macro:      fn,
				LocalMacro: local,
			})
		}
	}
	return invocations
}

func collectTargets(file *ast.File) []Target {
	var targets []Target
	for _, typ := range file.Types {
		targets = append(targets, Target{
			Kind:       StructTarget,
			Name:       typ.Name,
			SourcePath: typ.SourcePath,
			Pos:        typ.Pos,
			Struct:     typ,
		})
		for i := range typ.Fields {
			field := &typ.Fields[i]
			targets = append(targets, Target{
				Kind:       FieldTarget,
				Name:       field.Name,
				ParentName: typ.Name,
				SourcePath: typ.SourcePath,
				Pos:        field.Pos,
				Struct:     typ,
				Field:      field,
			})
		}
		for _, method := range typ.Methods {
			targets = append(targets, Target{
				Kind:       FunctionTarget,
				Name:       method.Name,
				ParentName: typ.Name,
				SourcePath: method.SourcePath,
				Pos:        method.Pos,
				Struct:     typ,
				Function:   method,
			})
		}
	}
	for _, enum := range file.Enums {
		targets = append(targets, Target{
			Kind:       EnumTarget,
			Name:       enum.Name,
			SourcePath: enum.SourcePath,
			Pos:        enum.Pos,
			Enum:       enum,
		})
		for i := range enum.Members {
			member := &enum.Members[i]
			targets = append(targets, Target{
				Kind:       EnumMemberTarget,
				Name:       member.Name,
				ParentName: enum.Name,
				SourcePath: enum.SourcePath,
				Pos:        member.Pos,
				Enum:       enum,
				EnumMember: member,
			})
		}
	}
	for _, fn := range file.Functions {
		targets = append(targets, Target{
			Kind:       FunctionTarget,
			Name:       fn.Name,
			SourcePath: fn.SourcePath,
			Pos:        fn.Pos,
			Function:   fn,
		})
	}
	return targets
}

func targetAnnotations(target Target) []*ast.Annotation {
	var annotations []ast.Annotation
	switch target.Kind {
	case StructTarget:
		annotations = target.Struct.Annotations
	case FieldTarget:
		annotations = target.Field.Annotations
	case EnumTarget:
		annotations = target.Enum.Annotations
	case EnumMemberTarget:
		annotations = target.EnumMember.Annotations
	case FunctionTarget:
		annotations = target.Function.Annotations
	}
	out := make([]*ast.Annotation, 0, len(annotations))
	for i := range annotations {
		out = append(out, &annotations[i])
	}
	return out
}
