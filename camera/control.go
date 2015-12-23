package camera

import (
	"fmt"
	"net/http"
	"time"
)

const (
	CameraIp      = "192.168.1.128"
	RtspStreamUri = "rtsp://" + CameraIp + ":554/12"
	Server        = "http://" + CameraIp + "/cgi-bin/hi3510"
	StepTime      = 2
	ScanTime      = 20
)

func Move(direction CameraDirection) {
	moveUrl := directionContext(direction)
	callCamera(moveUrl)
}

func MoveStep(direction CameraDirection) {
	moveUrl := directionContext(direction)
	moveAndStop(moveUrl, StepTime)
}

func Scan(direction ScanDirection) {
	scanUrl := scanContext(direction)
	moveAndStop(scanUrl, ScanTime)
}

func moveAndStop(context string, wait int) {
	stopUrl := directionContext(Stop)
	callCamera(context)

	stepTime := time.Duration(wait) * time.Second
	time.Sleep(stepTime)

	callCamera(stopUrl)
}

func callCamera(context string) {
	http.Get(Server + context)
}

func scanContext(direction ScanDirection) string {
	if direction == Vertical {
		return "/ptzctrl.cgi?-act=vscan"
	}
	return "/ptzctrl.cgi?-act=hscan"
}

func directionContext(direction CameraDirection) string {
	switch direction {
	case UpLeft:
		return "/ptzctrl.cgi?&-act=upleft"
	case UpRight:
		return "/ptzctrl.cgi?&-act=upright"
	case DownRight:
		return "/ptzctrl.cgi?&-act=downright"
	case DownLeft:
		return "/ptzctrl.cgi?&-act=downleft"
	case Up:
		return "/ptzup.cgi"
	case Right:
		return "/ptzright.cgi"
	case Down:
		return "/ptzdown.cgi"
	case Left:
		return "/ptzleft.cgi"
	case Stop:
		return "/ptzstop.cgi"
	}
	return ""
}

func SetPoint(position CameraPosition) string {
	return fmt.Sprintf("/ptzsetpoint.cgi?-point=%d", position)
}

func GotoPoint(position CameraPosition) string {
	return fmt.Sprintf("/ptzgotopoint.cgi?-point=%v", position)
}
