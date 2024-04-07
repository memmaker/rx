package util

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// pair with: https://github.com/Knetic/govaluate
// for full effect
// or checkout (b/c still maintained) https://github.com/expr-lang/expr

type Arguments []string

func (a Arguments) Get(index int) string {
	if index >= len(a) || index < 0 {
		return ""
	}
	return a[index]
}

func (a Arguments) GetInt(index int) int {
	if index >= len(a) || index < 0 {
		return 0
	}
	result, err := strconv.Atoi(a[index])
	if err != nil {
		return 0
	}
	return result
}

func (a Arguments) GetFloat(i int) float64 {
	if i >= len(a) || i < 0 {
		return 0
	}
	result, err := strconv.ParseFloat(a[i], 64)
	if err != nil {
		return 0
	}
	return result
}

func (a Arguments) IsEmpty() bool {
	return len(a) == 0
}

func LooksLikeAFunction(line string) bool {
	regexpPattern := regexp.MustCompile(`([A-Za-z_]+)\((.*)\)`)
	return regexpPattern.MatchString(line)
}
func NewArgumentsFromString(args string) Arguments {
	return trimAll(strings.Split(args, "|"))
}
func GetNameAndArgs(line string) (string, Arguments) {
	regexpPattern := regexp.MustCompile(`([A-Za-z_]+)\((.*)\)`)
	matches := regexpPattern.FindStringSubmatch(line)
	if matches == nil {
		return line, Arguments{}
	}
	name := matches[1]
	argString := matches[2]
	if argString == "" {
		return name, Arguments{}
	}
	args := trimAll(strings.Split(argString, "|"))
	return name, args
}

func trimAll(split []string) []string {
	for i, s := range split {
		split[i] = strings.TrimSpace(s)
	}
	return split
}

type CombinedPredicate struct {
	andPredicates []func() bool
	orPredicates  []func() bool
}

func (p CombinedPredicate) And(predicates ...func() bool) CombinedPredicate {
	p.andPredicates = append(p.andPredicates, predicates...)
	return p
}
func (p CombinedPredicate) Or(predicates ...func() bool) CombinedPredicate {
	p.orPredicates = append(p.orPredicates, predicates...)
	return p
}
func (p CombinedPredicate) Evaluate() bool {
	and := true
	for _, predicate := range p.andPredicates {
		and = and && predicate()
	}
	if len(p.orPredicates) == 0 {
		return and
	}
	or := false
	for _, predicate := range p.orPredicates {
		or = or || predicate()
	}
	return and && or
}

func (p CombinedPredicate) IsEmpty() bool {
	return len(p.andPredicates) == 0 && len(p.orPredicates) == 0
}

type AnyFunc func(...any) any
type ActionFunc func(...any)

type Predicate func(...any) bool

type Logic struct {
	PredicateMap map[string]Predicate
	actionMap    map[string]ActionFunc
	anyMap       map[string]AnyFunc
	Variables    map[string]func() any
}

func NewLogicCore() *Logic {
	return &Logic{
		PredicateMap: make(map[string]Predicate),
		actionMap:    make(map[string]ActionFunc),
		anyMap:       make(map[string]AnyFunc),
		Variables:    make(map[string]func() any),
	}
}

func (p *Logic) RegisterPredicate(name string, predicate Predicate) {
	p.PredicateMap[name] = predicate
}

func (p *Logic) RegisterAction(name string, action ActionFunc) {
	p.actionMap[name] = action
}
func (p *Logic) RegisterAny(name string, any AnyFunc) {
	p.anyMap[name] = any
}

func (p *Logic) HandleAssignment(line string, resolveNow bool) {
	parts := strings.Split(line, "=")
	assignParts := trimAll(parts)
	variableName := assignParts[0]
	if LooksLikeAFunction(assignParts[1]) {
		// variable is a function call
		functionName, stringArgs := GetNameAndArgs(assignParts[1])
		functionToCall := p.anyMap[functionName]
		if resolveNow { // resolve now
			args := p.StringResolve(stringArgs)
			resolvedArgs := p.ResolveArgs(args)
			if functionToCall == nil {
				println(fmt.Sprintf("Script ERR: Function %s not found", functionName))
				return
			}
			functionResult := functionToCall(resolvedArgs...)
			p.Variables[variableName] = func() any {
				return functionResult
			}
		} else { // lazy resolve when used
			p.Variables[variableName] = func() any {
				args := p.StringResolve(stringArgs)
				functionResult := functionToCall(p.ResolveArgs(args)...)
				return functionResult
			}
		}
	} else {
		// variable is a value
		p.Variables[variableName] = func() any {
			return assignParts[1]
		}
	}
}
func (p *Logic) LineToAction(line string) func() {
	name, stringArgs := GetNameAndArgs(line)
	println(fmt.Sprintf("Parsed action '%s(%s)'", name, stringArgs))
	args := p.StringResolve(stringArgs)
	action := p.actionMap[name]
	if action == nil {
		println(fmt.Sprintf("Script ERR: Action %s not found. Returning empty action.", name))
		return func() {}
	}
	actionCall := func() {
		action(p.ResolveArgs(args)...)
	}
	return actionCall
}

func (p *Logic) LineToPredicate(line string) func() bool {
	name, stringArgs := GetNameAndArgs(line)
	println(fmt.Sprintf("Parsed predicate '%s(%s)'", name, stringArgs))
	args := p.StringResolve(stringArgs)
	predicate := p.PredicateMap[name]
	if predicate == nil {
		println(fmt.Sprintf("Script ERR: Predicate %s not found. Returning always FALSE predicate.", name))
		return func() bool { return false }
	}
	predicateCall := func() bool {
		return predicate(p.ResolveArgs(args)...)
	}
	return predicateCall
}

func (p *Logic) ResolveArgs(args []func() any) []any {
	resolvedArgs := make([]any, len(args))
	for i, arg := range args {
		if arg == nil {
			println("Script ERR: Argument resolved to nil. Returning empty string.")
			resolvedArgs[i] = ""
		} else {
			resolvedArgs[i] = arg()
		}
	}
	return resolvedArgs
}
func (p *Logic) StringResolve(stringArgs []string) []func() any {
	args := make([]func() any, len(stringArgs))
	for i, arg := range stringArgs {
		value := arg
		/*
			if arg == "$PLAYER" {
				args[i] = func() any { return p.Player }
			} else
		*/
		if strings.HasPrefix(value, "$") {
			args[i] = p.Variables[value]
		} else {
			args[i] = func() any { return value }
		}
	}
	return args
}

func (p *Logic) ResolveVariable(name string) any {
	return p.Variables[name]()
}
