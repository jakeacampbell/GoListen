package main

import (
	"fmt"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	"github.com/faiface/beep"
	"github.com/faiface/beep/mp3"
	"github.com/faiface/beep/speaker"
	"log"
	"os"
	"time"
)

type Audio interface {
	GetName() string
	GetPath() string
}

type Song struct {
	FileName   string
	SongName   string
	ArtistName string
	AlbumName  string
	SongLength int
	streamer   beep.StreamSeekCloser
}

func (s *Song) GetName() string { return s.SongName }
func (s *Song) GetPath() string { return s.FileName }

var (
	ctrl *beep.Ctrl
	done chan bool
)

func main() {
	var playlist []Audio

	playlist = append(playlist,
		&Song{"C:\\Users\\campb\\Music\\Let.mp3", "Mine", "Bazzi", "Eyes", 153, nil},
		&Song{"", "Paradise", "Bazzi", "Eyes", 153, nil},
	)

	myApp := app.New()
	myWindow := myApp.NewWindow("GoListen")

	myWindow.Resize(fyne.NewSize(1920, 1080))

	myPlaylist := MakePlaylistView(playlist)
	myPlaylist.OnSelected = func(id widget.ListItemID) {
		CloseAudio()

		go PlayAudio(playlist[id])
	}

	playbackButton := widget.NewButton("Pause/Resume", func() {
		if ctrl == nil {
			log.Println("No audio is currently playing.")
			return
		}

		speaker.Lock()
		ctrl.Paused = !ctrl.Paused
		speaker.Unlock()
	})

	windowContent := container.NewBorder(
		nil,
		playbackButton,
		nil,
		nil,
		container.NewVScroll(myPlaylist),
	)

	myWindow.SetContent(windowContent)
	myWindow.ShowAndRun()
}

func CloseAudio() {
	if done != nil {
		speaker.Lock()
		if ctrl != nil {
			ctrl.Paused = true
		}
		speaker.Unlock()

		select {
		case <-done:
			// Do nothing we have already closed
		default:
			close(done)
		}

		done = nil
		ctrl = nil
	}
}

func PlayAudio(AudioInfo Audio) {
	log.Println("Now playing: " + AudioInfo.GetName())

	// Open the file
	archive, err := os.Open(AudioInfo.GetPath())
	if err != nil {
		log.Println(err)
		return
	}
	defer archive.Close()

	// Decode into a streamer
	streamer, format, err := mp3.Decode(archive)
	if err != nil {
		log.Println(err)
		return
	}
	defer streamer.Close()

	// Clear up existing playback
	if ctrl != nil {
		speaker.Lock()
		ctrl.Paused = true
		speaker.Unlock()
	}

	ctrl = &beep.Ctrl{
		Streamer: streamer,
		Paused:   false,
	}

	speaker.Init(format.SampleRate, format.SampleRate.N(time.Second/10))

	done = make(chan bool)

	speaker.Play(beep.Seq(ctrl, beep.Callback(func() {
		CloseAudio()
	})))

	for {
		select {
		case <-done:
			return
		case <-time.After(time.Second):
			speaker.Lock()
			fmt.Println(format.SampleRate.D(streamer.Position()).Round(time.Second))
			speaker.Unlock()
		}
	}
}

func MakePlaylistView(source []Audio) *widget.List {
	return widget.NewList(
		func() int { return len(source) },
		func() fyne.CanvasObject { return widget.NewLabel("Playlist") },
		func(i int, object fyne.CanvasObject) {
			object.(*widget.Label).SetText(source[i].GetName())
		},
	)
}
