package game

import "strings"

type CyberWare int

const (
	CyberWareInvalid CyberWare = -1
	CyberWareLight   CyberWare = iota
	CyberWareClock
)

func NewCyberWareFromString(s string) CyberWare {
	switch strings.ToLower(s) {
	case "light":
		return CyberWareLight
	case "clock":
		return CyberWareClock
	}
	return CyberWareInvalid
}
