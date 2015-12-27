package camera

import (
	"fmt"
	vlc "github.com/jteeuwen/go-vlc"
	"os"
	"time"
)

type playbackWindow struct {
	Width, Height          float64
	LimR, LimL, LimT, LimB float64
	LimitSizePercentage    float64
}

type viewer interface {
	registerEventListeners(*vlc.EventManager, *vlc.Player)
	playLoop(*vlc.Player)
	keepPlaying() bool
	isInitialized() bool
}

type VlcViewer struct {
	currentPanDirection CameraDirection
	lastDirectionChange time.Time
	escapeSequence      []CameraDirection
	escapeSeqId         int
	playing             bool
	screen              playbackWindow
	initialized         bool
}

type VideoViewer struct {
	*VlcViewer
}

type CameraViewer struct {
	control *CameraControl
	*VlcViewer
}

func (viewer *VlcViewer) isInitialized() bool {
	return viewer.initialized
}

func (viewer *VlcViewer) keepPlaying() bool {
	return viewer.playing
}

func NewVideoViewer() *VideoViewer {
	escapeSeq := make([]CameraDirection, 4)
	vlc := VlcViewer{Center, time.Now(), escapeSeq, 0, true, playbackWindow{0, 0, 0, 0, 0, 0, 0.15}, false}
	return &VideoViewer{&vlc}
}

func (viewer *VideoViewer) Play(mediaPath string) {
	preparePlayback(viewer, mediaPath)
}

func (viewer *VideoViewer) playLoop(player *vlc.Player) {
	x, y, _ := player.Cursor(0)
	newDirection := cursorPosition(viewer.screen.LimR, viewer.screen.LimL, viewer.screen.LimT, viewer.screen.LimB, float64(x), float64(y))

	if newDirection != viewer.currentPanDirection {
		viewer.currentPanDirection = newDirection
		viewer.lastDirectionChange = time.Now()
		viewer.escapeSequence = append(viewer.escapeSequence, viewer.currentPanDirection)[1:5]

		switch viewer.currentPanDirection {
		case Up:
		case Down:
		case Right:
		case Left:
		}
	}

	if viewer.escapeSeqId = isEscapeSequence(viewer.escapeSequence); viewer.escapeSeqId != 0 {
		viewer.playing = false
	}
}

func (viewer *VideoViewer) registerEventListeners(evt *vlc.EventManager, player *vlc.Player) {
	// implement
}

func NewCameraViewer() *CameraViewer {
	escapeSeq := make([]CameraDirection, 4)
	vlc := VlcViewer{Center, time.Now(), escapeSeq, 0, true, playbackWindow{0, 0, 0, 0, 0, 0, 0.15}, false}
	return &CameraViewer{NewCameraControl(), &vlc}
}

func (viewer *CameraViewer) Play() {
	preparePlayback(viewer, RtspStreamUri)
}

func (viewer *CameraViewer) registerEventListeners(evt *vlc.EventManager, player *vlc.Player) {
	evt.Attach(vlc.MediaPlayerPlaying, viewer.onPlaying, player)
	evt.Attach(vlc.MediaPlayerPositionChanged, viewer.onPositionChanged, player)
	evt.Attach(vlc.MediaPlayerStopped, viewer.onStop, player)
}

func (viewer *CameraViewer) playLoop(player *vlc.Player) {
	readMousePosition(viewer, player)
}

func (viewer *CameraViewer) onPositionChanged(evt *vlc.Event, data interface{}) {
	if viewer.screen.Width == 0 || viewer.screen.Height == 0 {
		width, height, _ := data.(*vlc.Player).Size(0)
		viewer.screen.Width, viewer.screen.Height = float64(width), float64(height)

		viewer.screen.LimR = viewer.screen.Width / (1.0 + viewer.screen.LimitSizePercentage)
		viewer.screen.LimL = viewer.screen.Width * viewer.screen.LimitSizePercentage
		viewer.screen.LimT = viewer.screen.Height * viewer.screen.LimitSizePercentage
		viewer.screen.LimB = viewer.screen.Height / (1.0 + viewer.screen.LimitSizePercentage)
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
	viewer.playing = false
}

func preparePlayback(viewer viewer, mediaUri string) {
	var inst *vlc.Instance
	var player *vlc.Player
	var evt *vlc.EventManager
	var err error

	if inst, err = vlc.New([]string{"-vvv"}); err != nil {
		fmt.Fprintf(os.Stderr, "[e] New(): %v", err)
		return
	}
	defer inst.Release()

	if player, err = loadMedia(mediaUri, inst); err != nil {
		fmt.Fprintf(os.Stderr, "[e] Player(): %v", err)
		return
	}
	defer player.Release()

	if evt, err = player.Events(); err != nil {
		fmt.Fprintf(os.Stderr, "[e] Events(): %v", err)
		return
	}

	viewer.registerEventListeners(evt, player)

	player.SetMouseInput(false)
	player.SetKeyInput(false)
	player.ToggleFullscreen()

	player.Play()

	for viewer.keepPlaying() {
		time.Sleep(1e8)
		if viewer.isInitialized() {
			viewer.playLoop(player)
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
	x, y, _ := player.Cursor(0)
	newDirection := cursorPosition(viewer.screen.LimR, viewer.screen.LimL, viewer.screen.LimT, viewer.screen.LimB, float64(x), float64(y))

	if newDirection != viewer.currentPanDirection {
		viewer.currentPanDirection = newDirection
		viewer.lastDirectionChange = time.Now()
		viewer.escapeSequence = append(viewer.escapeSequence, viewer.currentPanDirection)[1:5]

		if viewer.currentPanDirection != Center {
			viewer.control.Move(viewer.currentPanDirection)
		} else {
			viewer.control.Stop()
		}
	}

	if viewer.control.CameraIsMoving && time.Since(viewer.lastDirectionChange) > time.Duration(StepTime)*time.Second {
		viewer.control.Stop()
	}

	if viewer.escapeSeqId = isEscapeSequence(viewer.escapeSequence); viewer.escapeSeqId != 0 {
		viewer.playing = false
		viewer.control.Stop()
	}
}

func isEscapeSequence(escapeSeq []CameraDirection) int {
	if escapeSeq[0] == UpLeft && escapeSeq[1] == UpRight && escapeSeq[2] == DownRight && escapeSeq[3] == DownLeft {
		return 2
	} else if escapeSeq[0] == UpLeft && escapeSeq[1] == DownLeft && escapeSeq[2] == DownRight && escapeSeq[3] == UpRight {
		return 1
	}
	return 0
}

func cursorPosition(limR, limL, limT, limB float64, x, y float64) CameraDirection {
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
