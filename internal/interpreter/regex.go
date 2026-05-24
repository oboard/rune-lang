package interpreter

import (
	"fmt"
	"regexp"

	"github.com/oboard/rune-lang/internal/ir"
)

func newRegex(source string, flags string) (*Regex, error) {
	canonical, err := canonicalRegexFlags(flags)
	if err != nil {
		return nil, err
	}
	expr, err := regexp.Compile(goRegexPattern(source, canonical))
	if err != nil {
		return nil, err
	}
	return &Regex{Source: source, Flags: canonical, expr: expr}, nil
}

func goRegexPattern(source string, flags string) string {
	prefix := ""
	if regexHasFlag(flags, 'i') {
		prefix += "i"
	}
	if regexHasFlag(flags, 'm') {
		prefix += "m"
	}
	if regexHasFlag(flags, 's') {
		prefix += "s"
	}
	if prefix == "" {
		return source
	}
	return "(?" + prefix + ")" + source
}

func canonicalRegexFlags(flags string) (string, error) {
	seen := map[rune]bool{}
	for _, flag := range flags {
		switch flag {
		case 'd', 'g', 'i', 'm', 's', 'u', 'v', 'y':
		default:
			return "", fmt.Errorf("invalid regex flag %q", flag)
		}
		if seen[flag] {
			return "", fmt.Errorf("duplicate regex flag %q", flag)
		}
		seen[flag] = true
	}
	if seen['u'] && seen['v'] {
		return "", fmt.Errorf("regex flags 'u' and 'v' cannot be used together")
	}
	out := ""
	for _, flag := range "dgimsuvy" {
		if seen[flag] {
			out += string(flag)
		}
	}
	return out, nil
}

func regexHasFlag(flags string, flag rune) bool {
	for _, item := range flags {
		if item == flag {
			return true
		}
	}
	return false
}

func (i *Interpreter) callRegexMethod(regex *Regex, name string, args []ir.Expr, env *Env) (Value, error) {
	values, err := i.evalArgs(args, env)
	if err != nil {
		return nil, err
	}
	stringArg := func(index int) (string, error) {
		if index >= len(values) {
			return "", fmt.Errorf("regex.%s expects more args", name)
		}
		arg, ok := values[index].(string)
		if !ok {
			return "", fmt.Errorf("regex.%s argument %d expects String", name, index+1)
		}
		return arg, nil
	}
	switch name {
	case "exec":
		input, err := stringArg(0)
		if err != nil {
			return nil, err
		}
		return stringsToArray(regex.exec(input)), nil
	case "match":
		input, err := stringArg(0)
		if err != nil {
			return nil, err
		}
		return stringsToArray(regex.match(input)), nil
	case "matchAll":
		input, err := stringArg(0)
		if err != nil {
			return nil, err
		}
		return stringGroupsToArray(regex.matchAll(input)), nil
	case "test":
		input, err := stringArg(0)
		if err != nil {
			return nil, err
		}
		return regex.test(input), nil
	case "replace":
		input, err := stringArg(0)
		if err != nil {
			return nil, err
		}
		replacement, err := stringArg(1)
		if err != nil {
			return nil, err
		}
		return regex.replace(input, replacement), nil
	case "replaceAll":
		input, err := stringArg(0)
		if err != nil {
			return nil, err
		}
		replacement, err := stringArg(1)
		if err != nil {
			return nil, err
		}
		return regex.replaceAll(input, replacement), nil
	case "search":
		input, err := stringArg(0)
		if err != nil {
			return nil, err
		}
		return regex.search(input), nil
	case "split":
		input, err := stringArg(0)
		if err != nil {
			return nil, err
		}
		return stringsToArray(regex.expr.Split(input, -1)), nil
	case "source":
		return regex.Source, nil
	case "flags":
		return regex.Flags, nil
	case "global":
		return regexHasFlag(regex.Flags, 'g'), nil
	case "ignoreCase":
		return regexHasFlag(regex.Flags, 'i'), nil
	case "multiline":
		return regexHasFlag(regex.Flags, 'm'), nil
	case "dotAll":
		return regexHasFlag(regex.Flags, 's'), nil
	case "unicode":
		return regexHasFlag(regex.Flags, 'u'), nil
	case "unicodeSets":
		return regexHasFlag(regex.Flags, 'v'), nil
	case "sticky":
		return regexHasFlag(regex.Flags, 'y'), nil
	case "hasIndices":
		return regexHasFlag(regex.Flags, 'd'), nil
	case "lastIndex":
		return regex.LastIndex, nil
	case "setLastIndex":
		if len(values) != 1 {
			return nil, fmt.Errorf("regex.setLastIndex expects 1 arg, got %d", len(values))
		}
		index, ok := values[0].(int)
		if !ok {
			return nil, fmt.Errorf("regex.setLastIndex expects Int")
		}
		regex.LastIndex = index
		return regex.LastIndex, nil
	default:
		return nil, fmt.Errorf("regex.%s is not supported by the interpreter", name)
	}
}

func (r *Regex) exec(input string) []string {
	matches := r.execSubmatch(input)
	if matches == nil {
		return []string{}
	}
	return matches
}

func (r *Regex) execSubmatch(input string) []string {
	start := 0
	stateful := regexHasFlag(r.Flags, 'g') || regexHasFlag(r.Flags, 'y')
	if stateful {
		start = r.LastIndex
		if start < 0 || start > len(input) {
			r.LastIndex = 0
			return nil
		}
	}
	loc := r.expr.FindStringSubmatchIndex(input[start:])
	if loc == nil || (regexHasFlag(r.Flags, 'y') && loc[0] != 0) {
		if stateful {
			r.LastIndex = 0
		}
		return nil
	}
	if stateful {
		r.LastIndex = start + loc[1]
	}
	return r.expr.FindStringSubmatch(input[start+loc[0] : start+loc[1]])
}

func (r *Regex) match(input string) []string {
	if regexHasFlag(r.Flags, 'g') {
		matches := r.expr.FindAllString(input, -1)
		if matches == nil {
			return []string{}
		}
		return matches
	}
	return r.exec(input)
}

func (r *Regex) matchAll(input string) [][]string {
	matches := r.expr.FindAllStringSubmatch(input, -1)
	if matches == nil {
		return [][]string{}
	}
	return matches
}

func (r *Regex) test(input string) bool {
	return r.execSubmatch(input) != nil
}

func (r *Regex) replace(input string, replacement string) string {
	if regexHasFlag(r.Flags, 'g') {
		return r.replaceAll(input, replacement)
	}
	loc := r.expr.FindStringSubmatchIndex(input)
	if loc == nil {
		return input
	}
	out := []byte(input[:loc[0]])
	out = r.expr.ExpandString(out, replacement, input, loc)
	out = append(out, input[loc[1]:]...)
	return string(out)
}

func (r *Regex) replaceAll(input string, replacement string) string {
	return r.expr.ReplaceAllString(input, replacement)
}

func (r *Regex) search(input string) int {
	loc := r.expr.FindStringIndex(input)
	if loc == nil {
		return -1
	}
	return loc[0]
}

func stringsToArray(values []string) *Array {
	out := &Array{Elements: make([]Value, 0, len(values))}
	for _, value := range values {
		out.Elements = append(out.Elements, value)
	}
	return out
}

func stringGroupsToArray(groups [][]string) *Array {
	out := &Array{Elements: make([]Value, 0, len(groups))}
	for _, group := range groups {
		out.Elements = append(out.Elements, stringsToArray(group))
	}
	return out
}
