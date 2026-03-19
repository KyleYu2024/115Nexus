package web

import (
	"embed"
)

//go:embed all:frontend
var frontend embed.FS

const AppLogoSVG = `<svg width="512" height="512" viewBox="0 0 512 512" fill="none" xmlns="http://www.w3.org/2000/svg"><defs><linearGradient id="iconGradient" x1="0" y1="0" x2="512" y2="512" gradientUnits="userSpaceOnUse"><stop stop-color="#3A8DFF"/><stop offset="1" stop-color="#003366"/></linearGradient><filter id="shadowFilter" x="-5%" y="-5%" width="110%" height="110%"><feDropShadow dx="4" dy="6" stdDeviation="5" flood-opacity="0.2"/></filter></defs><rect x="0" y="0" width="512" height="512" rx="114" fill="url(#iconGradient)"/><g filter="url(#shadowFilter)"><path d="M256 96 L416 384 L336 384 L288 288 L224 288 L176 384 L96 384 Z M256 168 L200 288 L312 288 Z" fill="#FFFFFF" fill-opacity="0.95"/></g></svg>`

func GetFrontendFS() embed.FS {
	return frontend
}
