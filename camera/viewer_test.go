package camera

import "testing"

func TestPlayCamera(t *testing.T) {
	NewCameraViewer().Play()
}

func TestNewVideoViewer(t *testing.T) {
	NewVideoViewer().Play("/home/csoria/tmp/cam/151223_142945.mp4")
}
