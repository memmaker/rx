package recfile

import (
	"regexp"
	"strconv"
	"strings"
)

func IntStr(value int) string {
	return strconv.Itoa(value)
}
func Int64Str(value int64) string {
	return strconv.FormatInt(value, 10)
}

func StrInt(value string) int {
	atoi, _ := strconv.Atoi(value)
	return atoi
}
func Int32Str(value int32) string {
	return strconv.FormatInt(int64(value), 10)
}
func BoolStr(value bool) string {
	if value {
		return "true"
	}
	return "false"
}
func StringsStr(value []string) string {
	return strings.Join(value, "\n")
}
func StrBool(value string) bool {
	return value == "true"
}
func ToPredicate(name string, params ...string) string {
	return name + "(" + strings.Join(params, ",") + ")"
}
func ToPredicateSep(name, sep string, params ...string) string {
	return name + "(" + strings.Join(params, sep) + ")"
}

type StringPredicate []string

func StrPredicate(value string) StringPredicate {
	return StrPredicateSep(value, ",")
}

func StrPredicateSep(value, sep string) StringPredicate {
	// looks like this
	// text(param1, param2, param3..)
	predicateRegex := regexp.MustCompile(`^(\w+)\((.*)\)$`)
	if !predicateRegex.MatchString(value) {
		return nil
	}

	matches := predicateRegex.FindStringSubmatch(value)
	if len(matches) != 3 {
		return nil
	}
	name := matches[1]
	params := strings.Split(matches[2], sep)
	return append(StringPredicate{name}, params...)
}

func (p StringPredicate) Name() string {
	return p[0]
}

func (p StringPredicate) ParamCount() int {
	return len(p) - 1
}

func (p StringPredicate) GetString(index int) string {
	return strings.TrimSpace(p[index+1])
}

func (p StringPredicate) GetInt(index int) int {
	atoi, _ := strconv.Atoi(strings.TrimSpace(p[index+1]))
	return atoi
}

func (p StringPredicate) GetBool(index int) bool {
	return StrBool(strings.TrimSpace(p[index+1]))
}

func (p StringPredicate) GetFloat(index int) float64 {
	atof, _ := strconv.ParseFloat(strings.TrimSpace(p[index+1]), 64)
	return atof
}
func FloatStr(value float64) string {
	return strconv.FormatFloat(value, 'f', -1, 64)
}
func (p StringPredicate) Params() string {
	return strings.Join(p[1:], ", ")
}
