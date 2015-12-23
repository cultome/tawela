package camera

import (
	"fmt"
	"testing"
)

func TestMove(t *testing.T) {
	control := NewCameraControl()

	fmt.Println("Moving left...")
	control.MoveStep(Left)
	fmt.Println("Moving up...")
	control.MoveStep(Up)
	fmt.Println("Moving right...")
	control.MoveStep(Right)
	fmt.Println("Moving down...")
	control.MoveStep(Down)

	fmt.Println("Moving up-right...")
	control.MoveStep(UpRight)
	fmt.Println("Moving up-left...")
	control.MoveStep(UpLeft)
	fmt.Println("Moving down-right...")
	control.MoveStep(DownRight)
	fmt.Println("Moving down-left...")
	control.MoveStep(DownLeft)
}

func TestScan(t *testing.T) {
	control := NewCameraControl()

	fmt.Println("Scan horizontal...")
	control.Scan(Horizontal)
}
