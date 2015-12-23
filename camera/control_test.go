package camera

import (
	"fmt"
	"testing"
)

func TestMove(t *testing.T) {
	fmt.Println("Moving left...")
	MoveStep(Left)
	fmt.Println("Moving up...")
	MoveStep(Up)
	fmt.Println("Moving right...")
	MoveStep(Right)
	fmt.Println("Moving down...")
	MoveStep(Down)

	fmt.Println("Moving up-right...")
	MoveStep(UpRight)
	fmt.Println("Moving up-left...")
	MoveStep(UpLeft)
	fmt.Println("Moving down-right...")
	MoveStep(DownRight)
	fmt.Println("Moving down-left...")
	MoveStep(DownLeft)
}

func TestScan(t *testing.T) {
	fmt.Println("Scan horizontal...")
	Scan(Horizontal)
}
