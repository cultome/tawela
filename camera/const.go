package camera

type CameraDirection int

const (
	Center CameraDirection = iota
	UpLeft
	Up
	UpRight
	Right
	DownRight
	Down
	DownLeft
	Left
)

type ScanDirection int

const (
	Vertical ScanDirection = iota
	Horizontal
)

type CameraPosition int

const (
	One     CameraPosition = 1
	Two     CameraPosition = 2
	Three   CameraPosition = 3
	Four    CameraPosition = 4
	Default CameraPosition = 5
)

const (
	CameraIp            = "192.168.1.128"
	RtspStreamUri       = "rtsp://" + CameraIp + ":554/12"
	Server              = "http://" + CameraIp + "/cgi-bin/hi3510"
	StepTime            = 2
	ScanTime            = 20
	VideoFilenameRegexp = "^([\\d]{2})([\\d]{2})([\\d]{2})_([\\d]{2})([\\d]{2})([\\d]{2})\\.mp4$"
)
