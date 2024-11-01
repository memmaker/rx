package game

import "strings"

type CyberWare int

const (
	CyberWareInvalid CyberWare = -1
	CyberWareLight   CyberWare = iota
	CyberWareClock
	CyberWareThermopticCamouflage
	CyberWareCount
)

func NewCyberWareFromString(s string) CyberWare {
	switch strings.ToLower(s) {
	case "light":
		return CyberWareLight
	case "clock":
		return CyberWareClock
	case "thermoptic_camouflage":
		return CyberWareThermopticCamouflage
	}
	return CyberWareInvalid
}

func (w CyberWare) String() string {
	switch w {
	case CyberWareLight:
		return "Tactical Flashlight"
	case CyberWareClock:
		return "Basic Chronometer"
	case CyberWareThermopticCamouflage:
		return "Thermoptic Camouflage"
	}
	return "Invalid"
}

func (w CyberWare) DefaultPrice() int {
	switch w {
	case CyberWareLight:
		return 250
	case CyberWareClock:
		return 50
	case CyberWareThermopticCamouflage:
		return 9500
	}
	return 0
}
