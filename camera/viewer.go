package camera

import (
	"fmt"
	vlc "github.com/jteeuwen/go-vlc"
	"os"
	"time"
)

const uri = "rtsp://192.168.1.128:554/12"

var currentPanDirection CameraDirection
var lastDirectionChange time.Time
var cameraIsMoving = false
var escapeSequence = make([]CameraDirection, 4)
var keepPlaying = true

func NewViewer() {
	var inst *vlc.Instance
	var player *vlc.Player
	var evt *vlc.EventManager
	var err error

	if inst, err = vlc.New([]string{"-v"}); err != nil {
		fmt.Fprintf(os.Stderr, "[e] New(): %v", err)
		return
	}
	defer inst.Release()

	if player, err = loadMedia(uri, inst); err != nil {
		fmt.Fprintf(os.Stderr, "[e] Player(): %v", err)
		return
	}
	defer player.Release()

	if evt, err = player.Events(); err != nil {
		fmt.Fprintf(os.Stderr, "[e] Events(): %v", err)
		return
	}

	evt.Attach(vlc.MediaPlayerPositionChanged, onPositionChange, player)

	player.SetMouseInput(false)
	player.SetKeyInput(false)
	player.Play()
	//player.ToggleFullscreen()

	for keepPlaying {
		time.Sleep(1e8)
	}

	player.Stop()
}

func MoveCamera(direction CameraDirection) {
	cameraIsMoving = true
	fmt.Printf("[*] Moving camera to %v\n", direction)
	Move(direction)
}

func StopCamera() {
	cameraIsMoving = false
	fmt.Printf("[*] Stopping camera\n")
	Move(Stop)
}

func loadMedia(uri string, inst *vlc.Instance) (*vlc.Player, error) {
	var media *vlc.Media
	var player *vlc.Player
	var err error

	if media, err = inst.OpenMediaUri(uri); err != nil {
		fmt.Fprintf(os.Stderr, "[e] OpenMediaUri(): %v", err)
		return nil, err
	}

	if player, err = media.NewPlayer(); err != nil {
		fmt.Fprintf(os.Stderr, "[e] NewPlayer(): %v", err)
		media.Release()
		return nil, err
	}

	media.Release()
	media = nil

	return player, nil
}

func onPositionChange(evt *vlc.Event, data interface{}) {
	newDirection := cursorDirection(data.(*vlc.Player))

	if newDirection != currentPanDirection {
		currentPanDirection = newDirection
		lastDirectionChange = time.Now()
		addToEscapeSequence(currentPanDirection)

		if currentPanDirection != Center {
			MoveCamera(currentPanDirection)
		} else {
			StopCamera()
		}
	}

	if cameraIsMoving && time.Since(lastDirectionChange) > time.Duration(StepTime)*time.Second {
		StopCamera()
	}

	if isEscapeSequence() {
		keepPlaying = false
		StopCamera()
	}
}

func isEscapeSequence() bool {
	return escapeSequence[0] == UpLeft && escapeSequence[1] == UpRight && escapeSequence[2] == DownRight && escapeSequence[3] == DownLeft
}

func addToEscapeSequence(direction CameraDirection) {
	escapeSequence = append(escapeSequence, direction)[1:5]
}

func cursorDirection(player *vlc.Player) CameraDirection {
	width, height, _ := player.Size(0)
	limR, limL, limT, limB := calculateLimits(float64(width), float64(height), 0.15)

	x, y, _ := player.Cursor(0)
	panDirection := cursorPosition(limR, limL, limT, limB, float64(x), float64(y))
	return panDirection
}

func cursorPosition(limR, limL, limT, limB float64, x, y float64) CameraDirection {
	switch {
	case x <= limL && y <= limT:
		return UpLeft
	case x <= limL && y >= limB:
		return DownLeft
	case x >= limR && y <= limT:
		return UpRight
	case x >= limR && y >= limB:
		return DownRight
	case x <= limL:
		return Left
	case x >= limR:
		return Right
	case y <= limT:
		return Up
	case y >= limB:
		return Down
	}
	return Center
}

func calculateLimits(width, height, percent float64) (float64, float64, float64, float64) {
	limR := width / (1.0 + percent)
	limL := width * percent
	limT := height * percent
	limB := height / (1.0 + percent)

	return limR, limL, limT, limB
}
