package camera

import (
	"fmt"
	vlc "github.com/jteeuwen/go-vlc"
	"os"
	"time"
)

type CameraViewer struct {
	control             *CameraControl
	currentPanDirection CameraDirection
	lastDirectionChange time.Time
	escapeSequence      []CameraDirection
	escapeSeqId         int
	keepPlaying         bool
	isFullscreen        bool
}

func NewCameraViewer() *CameraViewer {
	escapeSeq := make([]CameraDirection, 4)
	return &CameraViewer{NewCameraControl(), Center, time.Now(), escapeSeq, 0, true, false}
}

func (viewer *CameraViewer) Start() {
	var inst *vlc.Instance
	var player *vlc.Player
	var evt *vlc.EventManager
	var err error

	if inst, err = vlc.New([]string{"-v"}); err != nil {
		fmt.Fprintf(os.Stderr, "[e] New(): %v", err)
		return
	}
	defer inst.Release()

	if player, err = viewer.loadMedia(RtspStreamUri, inst); err != nil {
		fmt.Fprintf(os.Stderr, "[e] Player(): %v", err)
		return
	}
	defer player.Release()

	if evt, err = player.Events(); err != nil {
		fmt.Fprintf(os.Stderr, "[e] Events(): %v", err)
		return
	}

	evt.Attach(vlc.MediaPlayerPositionChanged, viewer.onPositionChange, player)

	player.SetMouseInput(false)
	player.SetKeyInput(false)
	player.Play()

	if viewer.isFullscreen {
		player.ToggleFullscreen()
	}

	// setting intial position
	viewer.control.SetPoint(Default)

	for viewer.keepPlaying {
		time.Sleep(1e8)
	}

	player.Stop()

	if viewer.escapeSeqId == 2 {
		viewer.control.GotoPoint(Default)
	}
}

func (viewer *CameraViewer) loadMedia(uri string, inst *vlc.Instance) (*vlc.Player, error) {
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

func (viewer *CameraViewer) onPositionChange(evt *vlc.Event, data interface{}) {
	newDirection := viewer.cursorDirection(data.(*vlc.Player))

	if newDirection != viewer.currentPanDirection {
		viewer.currentPanDirection = newDirection
		viewer.lastDirectionChange = time.Now()
		viewer.addToEscapeSequence(viewer.currentPanDirection)

		if viewer.currentPanDirection != Center {
			viewer.control.Move(viewer.currentPanDirection)
		} else {
			viewer.control.Stop()
		}
	}

	if viewer.control.CameraIsMoving && time.Since(viewer.lastDirectionChange) > time.Duration(StepTime)*time.Second {
		viewer.control.Stop()
	}

	if viewer.escapeSeqId = viewer.isEscapeSequence(); viewer.escapeSeqId != 0 {
		viewer.keepPlaying = false
	}
}

func (viewer *CameraViewer) isEscapeSequence() int {
	if viewer.escapeSequence[0] == UpLeft && viewer.escapeSequence[1] == UpRight && viewer.escapeSequence[2] == DownRight && viewer.escapeSequence[3] == DownLeft {
		return 1
	} else if viewer.escapeSequence[0] == UpLeft && viewer.escapeSequence[1] == DownLeft && viewer.escapeSequence[2] == DownRight && viewer.escapeSequence[3] == UpRight {
		return 2
	}
	return 0
}

func (viewer *CameraViewer) addToEscapeSequence(direction CameraDirection) {
	viewer.escapeSequence = append(viewer.escapeSequence, direction)[1:5]
}

func (viewer *CameraViewer) cursorDirection(player *vlc.Player) CameraDirection {
	width, height, _ := player.Size(0)
	limR, limL, limT, limB := viewer.calculateLimits(float64(width), float64(height), 0.15)

	x, y, _ := player.Cursor(0)
	panDirection := viewer.cursorPosition(limR, limL, limT, limB, float64(x), float64(y))
	return panDirection
}

func (viewer *CameraViewer) cursorPosition(limR, limL, limT, limB float64, x, y float64) CameraDirection {
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

func (viewer *CameraViewer) calculateLimits(width, height, percent float64) (float64, float64, float64, float64) {
	limR := width / (1.0 + percent)
	limL := width * percent
	limT := height * percent
	limB := height / (1.0 + percent)

	return limR, limL, limT, limB
}
