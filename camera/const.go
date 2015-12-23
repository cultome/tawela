package camera

type CameraDirection int

const (
	UpLeft CameraDirection = iota
	Up
	UpRight
	Right
	DownRight
	Down
	DownLeft
	Left
	Center
	Stop
)

type ScanDirection int

const (
	Vertical ScanDirection = iota
	Horizontal
)

type CameraPosition int

const (
	First  CameraPosition = 1
	Second CameraPosition = 2
	Third  CameraPosition = 3
	Fourth CameraPosition = 4
	Fifth  CameraPosition = 5
)
