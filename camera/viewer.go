package camera

import (
	"fmt"
	vlc "github.com/jteeuwen/go-vlc"
	"os"
	"time"
)

type screen struct {
	width, height          float64
	limR, limL, limT, limB float64
	limitSizePercentage    float64
}

type CameraViewer struct {
	control             *CameraControl
	currentPanDirection CameraDirection
	lastDirectionChange time.Time
	escapeSequence      []CameraDirection
	escapeSeqId         int
	keepPlaying         bool
	isFullscreen        bool
	screen              screen
}

func NewCameraViewer() *CameraViewer {
	escapeSeq := make([]CameraDirection, 4)
	return &CameraViewer{NewCameraControl(), Center, time.Now(), escapeSeq, 0, true, true, screen{0, 0, 0, 0, 0, 0, 0.15}}
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

	evt.Attach(vlc.MediaPlayerPlaying, viewer.onPlaying, player)

	player.SetMouseInput(false)
	player.SetKeyInput(false)

	if viewer.isFullscreen {
		player.ToggleFullscreen()
	}

	player.Play()

	// setting intial position
	viewer.control.SetPoint(Default)

	for viewer.keepPlaying {
		time.Sleep(1e8)
		viewer.readMousePosition(player)
	}

	player.Stop()

	if viewer.escapeSeqId == 2 {
		// restoring initial position
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

func (viewer *CameraViewer) onPlaying(evt *vlc.Event, data interface{}) {
	width, height, _ := data.(*vlc.Player).Size(0)
	viewer.screen.width, viewer.screen.height = float64(width), float64(height)

	viewer.screen.limR = viewer.screen.width / (1.0 + viewer.screen.limitSizePercentage)
	viewer.screen.limL = viewer.screen.width * viewer.screen.limitSizePercentage
	viewer.screen.limT = viewer.screen.height * viewer.screen.limitSizePercentage
	viewer.screen.limB = viewer.screen.height / (1.0 + viewer.screen.limitSizePercentage)

}

func (viewer *CameraViewer) readMousePosition(player *vlc.Player) {
	newDirection := viewer.cursorDirection(player)

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
	x, y, _ := player.Cursor(0)
	panDirection := viewer.cursorPosition(viewer.screen.limR, viewer.screen.limL, viewer.screen.limT, viewer.screen.limB, float64(x), float64(y))
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
