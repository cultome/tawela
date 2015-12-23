package camera

import (
	"fmt"
	"net/http"
	"time"
)

type CameraControl struct {
	CameraIsMoving bool
}

func NewCameraControl() *CameraControl {
	return &CameraControl{false}
}

func (control *CameraControl) Move(direction CameraDirection) {
	command := control.directionCommand(direction)
	control.callCamera(command)
}

func (control *CameraControl) MoveStep(direction CameraDirection) {
	command := control.directionCommand(direction)
	control.moveAndStop(command, StepTime)
}

func (control *CameraControl) Scan(direction ScanDirection) {
	command := control.scanCommand(direction)
	control.moveAndStop(command, ScanTime)
}

func (control *CameraControl) Stop() {
	if control.CameraIsMoving {
		control.callCamera("/ptzstop.cgi")
		control.CameraIsMoving = false
	}
}

func (control *CameraControl) SetPoint(position CameraPosition) {
	control.Stop()
	command := fmt.Sprintf("/ptzsetpoint.cgi?-point=%d", position)
	control.callCamera(command)
}

func (control *CameraControl) GotoPoint(position CameraPosition) {
	control.Stop()
	command := fmt.Sprintf("/ptzgotopoint.cgi?-point=%d", position)
	control.callCamera(command)
}

func (control *CameraControl) moveCamera(command string) {
	control.Stop()
	control.callCamera(command)
	control.CameraIsMoving = true
}

func (control *CameraControl) moveAndStop(command string, wait int) {
	control.moveCamera(command)
	stepTime := time.Duration(wait) * time.Second
	time.Sleep(stepTime)
	control.Stop()
}

func (control *CameraControl) callCamera(command string) {
	http.Get(Server + command)
}

func (control *CameraControl) scanCommand(direction ScanDirection) string {
	if direction == Vertical {
		return "/ptzctrl.cgi?-act=vscan"
	}
	return "/ptzctrl.cgi?-act=hscan"
}

func (control *CameraControl) directionCommand(direction CameraDirection) string {
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
	}
	return ""
}
