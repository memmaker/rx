package game

import (
	"github.com/memmaker/go/fxtools"
	"strconv"
)

type Params struct {
	kvp map[string]string
}

func NewParams(values map[string]string) Params {
	return Params{kvp: values}
}
func (p Params) Has(key string) bool {
	_, exists := p.kvp[key]
	return exists
}
func (p Params) Get(key string) string {
	val, exists := p.kvp[key]
	if !exists {
		return ""
	}
	return val
}
func (p Params) GetInt(key string) int {
	if !p.Has(key) {
		return 0
	}
	atoi, _ := strconv.Atoi(p.kvp[key])
	return atoi
}
func (p Params) GetIntOrDefault(key string, defaultValue int) int {
	if !p.Has(key) {
		return defaultValue
	}
	atoi, _ := strconv.Atoi(p.kvp[key])
	return atoi
}

func (p Params) GetBool(key string) bool {
	if !p.Has(key) {
		return false
	}
	parseBool, _ := strconv.ParseBool(p.kvp[key])
	return parseBool
}
func (p Params) GetBoolOrDefault(key string, defaultValue bool) bool {
	if !p.Has(key) {
		return defaultValue
	}
	parseBool, _ := strconv.ParseBool(p.kvp[key])
	return parseBool
}
func (p Params) GetFloat(key string) float64 {
	if !p.Has(key) {
		return 0
	}
	parseFloat, _ := strconv.ParseFloat(p.kvp[key], 64)
	return parseFloat
}
func (p Params) GetFloatOrDefault(key string, defaultValue float64) float64 {
	if !p.Has(key) {
		return defaultValue
	}
	parseFloat, _ := strconv.ParseFloat(p.kvp[key], 64)
	return parseFloat
}

func (p Params) GetIntervalOrDefault(s string, interval fxtools.Interval) fxtools.Interval {
	if !p.Has(s) {
		return interval
	}
	return fxtools.ParseInterval(p.Get(s))
}
