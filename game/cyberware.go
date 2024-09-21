package game

import "strings"

type CyberWare int

const (
	CyberWareInvalid CyberWare = -1
	CyberWareLight   CyberWare = iota
)

func NewCyberWareFromString(s string) CyberWare {
	switch strings.ToLower(s) {
	case "light":
		return CyberWareLight
	}
	return CyberWareInvalid
}
