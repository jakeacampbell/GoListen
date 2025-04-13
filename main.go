package main

import (
	"fmt"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	"github.com/faiface/beep/mp3"
	"github.com/faiface/beep/speaker"
	"os"
	"time"
)

type Audio interface {
	GetName() string
	Play() error
}

type Song struct {
	fileName   string
	SongName   string
	ArtistName string
	AlbumName  string
	SongLength int
}

func (s *Song) GetName() string { return s.SongName }

func (s *Song) Play() error {
	fmt.Println("GoListen")

	file, err := os.Open("GoListen.mp3")
	if err != nil {
		return err
	}
	defer file.Close()

	streamer, format, err := mp3.Decode(file)
	if err != nil {
		return err
	}
	defer streamer.Close()

	speaker.Init(format.SampleRate, format.SampleRate.N(time.Second/10))

	speaker.Play(streamer)

	return nil
}

func main() {
	var playlist []Audio

	playlist = append(playlist, &Song{"", "Mine", "Bazzi", "Eyes", 153})

	myApp := app.New()
	myWindow := myApp.NewWindow("Hello")

	myWindow.Resize(fyne.NewSize(1920, 1080))

	windowContent := container.NewBorder(nil, nil, nil, nil, container.NewVScroll(MakePlaylistView(playlist)))

	myWindow.SetContent(windowContent)
	myWindow.Show()

	myApp.Run()
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
