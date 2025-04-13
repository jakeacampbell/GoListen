package main

import (
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
	GetPath() string
	GetName() string
	GetArtist() string
	GetAlbum() string
}

type Song struct {
	FileName   string
	SongName   string
	ArtistName string
	AlbumName  string
}

func (s *Song) GetName() string   { return s.SongName }
func (s *Song) GetPath() string   { return s.FileName }
func (s *Song) GetArtist() string { return s.ArtistName }
func (s *Song) GetAlbum() string  { return s.AlbumName }

var (
	ctrl           *beep.Ctrl
	done           chan bool
	playbackSlider *widget.Slider
)

func main() {
	var playlist []Audio

	playlist = append(playlist,
		&Song{"C:\\Users\\campb\\Music\\Let.mp3", "Mine", "Bazzi", "Eyes"},
		&Song{"", "Paradise", "Bazzi", "Eyes"},
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

	playbackSlider = widget.NewSlider(0, 60)
	playbackSlider.Step = 1.0
	playbackSlider.OnChanged = func(t float64) {
		speaker.Lock()
		// Seek new time
		speaker.Unlock()
	}

	windowContent := container.NewBorder(
		playbackSlider,
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

	done = make(chan bool)

	speaker.Init(format.SampleRate, format.SampleRate.N(time.Second/10))
	speaker.Play(beep.Seq(ctrl, beep.Callback(func() {
		CloseAudio()
	})))

	playbackSlider.Max = float64(format.SampleRate.D(streamer.Len()).Round(time.Second))

	for {
		select {
		case <-done:
			return
		case <-time.After(time.Second):
			speaker.Lock()
			fyne.Do(func() {
				playbackSlider.Value = float64(format.SampleRate.D(streamer.Position()).Round(time.Second))
				playbackSlider.Refresh()
			})
			//fmt.Println(format.SampleRate.D(streamer.Position()).Round(time.Second))
			speaker.Unlock()
		}
	}
}

func MakePlaylistView(source []Audio) *widget.List {
	return widget.NewList(
		func() int { return len(source) },
		func() fyne.CanvasObject {
			return container.NewHBox(
				widget.NewLabel("Song"),
				widget.NewLabel("Artist"),
				widget.NewLabel("Album"),
			)
		},
		func(i int, object fyne.CanvasObject) {
			row := object.(*fyne.Container)
			audio := source[i]

			row.Objects[0].(*widget.Label).SetText(audio.GetName())
			row.Objects[1].(*widget.Label).SetText(audio.GetArtist())
			row.Objects[2].(*widget.Label).SetText(audio.GetAlbum())
		},
	)
}
