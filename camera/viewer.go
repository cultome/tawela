package camera

import (
	"fmt"
	vlc "github.com/jteeuwen/go-vlc"
	"os"
	"time"
)

type playbackWindow struct {
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
	screen              playbackWindow
	initialized         bool
}

type Viewer interface {
	RegisterEventListeners(*vlc.EventManager, *vlc.Player)
	PlayLoop(*vlc.Player)
	KeepPlaying() bool
	IsInitialized() bool
}

func (p *playbackWindow) String() string {
	return fmt.Sprintf("[%vx%v] R: %v L: %v T: %v B: %v (%v%%)", p.width, p.height, p.limR, p.limL, p.limT, p.limB, p.limitSizePercentage)
}

func NewCameraViewer() *CameraViewer {
	escapeSeq := make([]CameraDirection, 4)
	return &CameraViewer{NewCameraControl(), Center, time.Now(), escapeSeq, 0, true, playbackWindow{0, 0, 0, 0, 0, 0, 0.15}, false}
}

func (viewer *CameraViewer) Start() {
	preparePlayback(viewer)
}

func (viewer *CameraViewer) IsInitialized() bool {
	return viewer.initialized
}

func (viewer *CameraViewer) RegisterEventListeners(evt *vlc.EventManager, player *vlc.Player) {
	evt.Attach(vlc.MediaPlayerPlaying, viewer.onPlaying, player)
	evt.Attach(vlc.MediaPlayerPositionChanged, viewer.onPositionChanged, player)
	evt.Attach(vlc.MediaPlayerStopped, viewer.onStop, player)
}

func (viewer *CameraViewer) PlayLoop(player *vlc.Player) {
	readMousePosition(viewer, player)
}

func (viewer *CameraViewer) KeepPlaying() bool {
	return viewer.keepPlaying
}

func (viewer *CameraViewer) onPositionChanged(evt *vlc.Event, data interface{}) {
	if viewer.screen.width == 0 || viewer.screen.height == 0 {
		width, height, _ := data.(*vlc.Player).Size(0)
		viewer.screen.width, viewer.screen.height = float64(width), float64(height)

		viewer.screen.limR = viewer.screen.width / (1.0 + viewer.screen.limitSizePercentage)
		viewer.screen.limL = viewer.screen.width * viewer.screen.limitSizePercentage
		viewer.screen.limT = viewer.screen.height * viewer.screen.limitSizePercentage
		viewer.screen.limB = viewer.screen.height / (1.0 + viewer.screen.limitSizePercentage)
	}
}

func (viewer *CameraViewer) onPlaying(evt *vlc.Event, data interface{}) {
	// setting intial position
	viewer.control.SetPoint(Default)
	viewer.initialized = true
}

func (viewer *CameraViewer) onStop(evt *vlc.Event, data interface{}) {
	if viewer.escapeSeqId == 2 {
		// restoring initial position
		viewer.control.GotoPoint(Default)
	}
	viewer.keepPlaying = false
}

func preparePlayback(viewer Viewer) {
	var inst *vlc.Instance
	var player *vlc.Player
	var evt *vlc.EventManager
	var err error

	if inst, err = vlc.New([]string{"-v"}); err != nil {
		fmt.Fprintf(os.Stderr, "[e] New(): %v", err)
		return
	}
	defer inst.Release()

	if player, err = loadMedia(RtspStreamUri, inst); err != nil {
		fmt.Fprintf(os.Stderr, "[e] Player(): %v", err)
		return
	}
	defer player.Release()

	if evt, err = player.Events(); err != nil {
		fmt.Fprintf(os.Stderr, "[e] Events(): %v", err)
		return
	}

	viewer.RegisterEventListeners(evt, player)

	player.SetMouseInput(false)
	player.SetKeyInput(false)
	//player.ToggleFullscreen()

	player.Play()

	for viewer.KeepPlaying() {
		time.Sleep(1e8)
		if viewer.IsInitialized() {
			viewer.PlayLoop(player)
		}
	}

	player.Stop()

	time.Sleep(time.Duration(2) * time.Second)
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

func readMousePosition(viewer *CameraViewer, player *vlc.Player) {
	newDirection := cursorDirection(viewer, player)

	if newDirection != viewer.currentPanDirection {
		viewer.currentPanDirection = newDirection
		viewer.lastDirectionChange = time.Now()
		addToEscapeSequence(viewer, viewer.currentPanDirection)

		if viewer.currentPanDirection != Center {
			viewer.control.Move(viewer.currentPanDirection)
		} else {
			viewer.control.Stop()
		}
	}

	if viewer.control.CameraIsMoving && time.Since(viewer.lastDirectionChange) > time.Duration(StepTime)*time.Second {
		viewer.control.Stop()
	}

	if viewer.escapeSeqId = isEscapeSequence(viewer); viewer.escapeSeqId != 0 {
		viewer.keepPlaying = false
		viewer.control.Stop()
	}
}

func isEscapeSequence(viewer *CameraViewer) int {
	if viewer.escapeSequence[0] == UpLeft && viewer.escapeSequence[1] == UpRight && viewer.escapeSequence[2] == DownRight && viewer.escapeSequence[3] == DownLeft {
		return 2
	} else if viewer.escapeSequence[0] == UpLeft && viewer.escapeSequence[1] == DownLeft && viewer.escapeSequence[2] == DownRight && viewer.escapeSequence[3] == UpRight {
		return 1
	}
	return 0
}

func addToEscapeSequence(viewer *CameraViewer, direction CameraDirection) {
	viewer.escapeSequence = append(viewer.escapeSequence, direction)[1:5]
}

func cursorDirection(viewer *CameraViewer, player *vlc.Player) CameraDirection {
	x, y, _ := player.Cursor(0)
	panDirection := cursorPosition(viewer, viewer.screen.limR, viewer.screen.limL, viewer.screen.limT, viewer.screen.limB, float64(x), float64(y))
	return panDirection
}

func cursorPosition(viewer *CameraViewer, limR, limL, limT, limB float64, x, y float64) CameraDirection {
	// This is to handle the initial-position-move problem
	if x == 0 && y == 0 {
		return Center
	}

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
