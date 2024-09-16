package game

import (
	"github.com/memmaker/go/fxtools"
)

type Params map[string]interface{}

func (p Params) Has(key string) bool {
	_, exists := p[key]
	return exists
}
func (p Params) Get(key string) string {
	val, exists := p[key]
	if !exists {
		return ""
	}
	return val.(string)
}
func (p Params) GetInt(key string) int {
	return p.GetIntOrDefault(key, -1)
}
func (p Params) GetIntOrDefault(key string, defaultValue int) int {
	if !p.Has(key) {
		return defaultValue
	}
	return p[key].(int)
}

func (p Params) GetBool(key string) bool {
	if !p.Has(key) {
		return false
	}
	return p.GetBoolOrDefault(key, false)
}
func (p Params) GetBoolOrDefault(key string, defaultValue bool) bool {
	if !p.Has(key) {
		return defaultValue
	}
	return p[key].(bool)
}
func (p Params) GetFloat(key string) float64 {
	if !p.Has(key) {
		return 0
	}
	return p.GetFloatOrDefault(key, 0)
}
func (p Params) GetFloatOrDefault(key string, defaultValue float64) float64 {
	if !p.Has(key) {
		return defaultValue
	}
	return p[key].(float64)
}

func (p Params) GetIntervalOrDefault(s string, interval fxtools.Interval) fxtools.Interval {
	if !p.Has(s) {
		return interval
	}
	return p[s].(fxtools.Interval)
}

func (p Params) GetInterval(key string) fxtools.Interval {
	return p.GetIntervalOrDefault(key, fxtools.Interval{})
}

func (p Params) GetDamageOrDefault(defaultDamage int) int {
	if p.Has("damage") {
		defaultDamage = p.GetInt("damage")
	} else if p.Has("damage_interval") {
		defaultDamage = p.GetInterval("damage_interval").Roll()
	}
	return defaultDamage
}

func (p Params) HasDamage() bool {
	return p.Has("damage") || p.Has("damage_interval")
}
