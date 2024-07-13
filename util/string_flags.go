package util

import "strconv"

type StringFlags map[string]int

func (sf StringFlags) Get(key string) int {
	if val, ok := sf[key]; ok {
		return val
	}
	return 0
}

func (sf StringFlags) Set(key string, val int) {
	sf[key] = val
}

func (sf StringFlags) HasFlag(key string) bool {
	return sf.Get(key) != 0
}

func (sf StringFlags) SetFlag(key string) {
	sf.Set(key, 1)
}

func (sf StringFlags) ClearFlag(key string) {
	delete(sf, key)
}

func (sf StringFlags) ToStringArray() []string {
	var rows []TableRow
	for key, val := range sf {
		if val != 0 {
			rows = append(rows, TableRow{Columns: []string{key, strconv.Itoa(val)}})
		}
	}
	return TableLayout(rows, []TextAlignment{AlignLeft, AlignRight})
}
