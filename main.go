package main

import (
	"fmt"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	"github.com/faiface/beep"
	"github.com/faiface/beep/effects"
	"github.com/faiface/beep/mp3"
	"github.com/faiface/beep/speaker"
	"log"
	"os"
	"path/filepath"
	"strings"
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

type PlaybackController struct {
	ctrl       *beep.Ctrl
	sampleRate beep.SampleRate
	volume     *effects.Volume
}

var (
	playback           PlaybackController
	done               chan bool
	playbackSlider     *widget.Slider
	currTimeText       *widget.Label
	currSongLengthText *widget.Label
	playlist           []Audio
)

func ScanDirectory(dir string) error {
	files, err := os.ReadDir(dir)
	if err != nil {
		return err
	}

	for _, file := range files {
		if file.IsDir() {
			continue
		}

		extension := filepath.Ext(file.Name())

		switch strings.ToLower(extension) {
		case ".mp3":
			full := filepath.Join(dir, file.Name())
			playlist = append(playlist, &Song{full, file.Name(), "", ""})
		default:
			continue
		}
	}

	return nil
}

func main() {

	//playlist = append(playlist,
	//	&Song{"C:\\Users\\campb\\Music\\Paradise.mp3", "Paradise", "Bazzi", "Eyes"},
	//	&Song{"C:\\Users\\campb\\Music\\Mine.mp3", "Mine", "Bazzi", "Eyes"},
	//	&Song{"C:\\Users\\campb\\Music\\get_up.mp3", "Get up for a ride", "", ""},
	//	&Song{"C:\\Users\\campb\\Music\\Feel-It.mp3", "Feel It", "d4vd", "Invincible OST"},
	//	&Song{"C:\\Users\\campb\\Music\\Headlock.mp3", "Headlock", "Imogen Heap", "Invincible OST"},
	//	&Song{"C:\\Users\\campb\\Music\\Might-not.mp3", "Might Not Make It Home", "LPX", "Invincible OST"},
	//	&Song{"C:\\Users\\campb\\Music\\Endorsi.mp3", "Endorsi", "Tower Of God", "Tower Of God OST"},
	//)

	if err := ScanDirectory("C:\\Users\\campb\\Music"); err != nil {
		log.Fatal(err)
	}

	myApp := app.New()
	myWindow := myApp.NewWindow("GoListen")

	myWindow.Resize(fyne.NewSize(1920, 1080))

	playlistView := MakePlaylistView(playlist)

	windowContent := container.NewBorder(
		nil,
		MakePlaybackContent(),
		nil,
		nil,
		container.NewHSplit(
			container.NewBorder(
				nil,
				nil,
				nil,
				nil,
				widget.NewLabel("Test")),
			container.NewVScroll(playlistView),
		),
	)

	myWindow.SetContent(windowContent)
	myWindow.ShowAndRun()
}

func MakePlaybackContent() fyne.CanvasObject {
	currTimeText = widget.NewLabel("0.00")
	currSongLengthText = widget.NewLabel("0.00")

	playbackSlider = widget.NewSlider(0, 60)
	playbackSlider.Step = 1.0
	playbackSlider.OnChanged = func(t float64) {
		speaker.Lock()
		defer speaker.Unlock()

		// Debugging: Check if the streamer supports seeking
		if streamSeekCloser, ok := playback.ctrl.Streamer.(beep.StreamSeekCloser); ok {
			if err := streamSeekCloser.Seek(int(float64(playback.sampleRate) * t)); err != nil {
				log.Println("Failed to seek:", err)
			}
		} else {
			log.Println("Streamer does not support seeking.")
		}

	}

	playbackProgress := container.NewBorder(
		nil,
		nil,
		currTimeText,
		currSongLengthText,
		playbackSlider,
	)

	playbackButton := widget.NewButton("Pause/Resume", func() {
		if playback.ctrl == nil {
			log.Println("No audio is currently playing.")
			return
		}

		speaker.Lock()
		playback.ctrl.Paused = !playback.ctrl.Paused
		speaker.Unlock()
	})

	volumeSlider := widget.NewSlider(-10, 0)
	volumeSlider.Value = 0 // Default volume
	volumeSlider.Step = 0.01
	volumeSlider.OnChanged = func(v float64) {
		SetVolume(v)
	}

	volumeControl := container.NewBorder(
		nil, nil,
		widget.NewLabel("Volume:"),
		nil,
		volumeSlider,
	)

	return container.NewBorder(
		playbackButton,
		nil,
		nil,
		nil,
		container.NewBorder(
			nil,
			nil,
			nil,
			volumeControl,
			playbackProgress,
		),
	)
}

func SetVolume(volume float64) {
	if playback.volume == nil {
		return
	}

	speaker.Lock()
	defer speaker.Unlock()

	if volume == -10 {
		playback.volume.Silent = true
	} else {
		playback.volume.Silent = false
	}

	playback.volume.Volume = volume
}

func CloseAudio() {
	if done != nil {
		speaker.Lock()
		if playback.ctrl != nil {
			playback.ctrl.Paused = true
		}
		speaker.Unlock()

		select {
		case <-done:
			// Do nothing we have already closed
		default:
			close(done)
		}

		done = nil
		playback.ctrl = nil
		playback.volume = nil
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

	playback.sampleRate = format.SampleRate

	// Clear up existing playback
	if playback.ctrl != nil {
		speaker.Lock()
		playback.ctrl.Paused = true
		speaker.Unlock()
	}

	playback.ctrl = &beep.Ctrl{
		Streamer: streamer,
		Paused:   false,
	}

	playback.volume = &effects.Volume{
		Streamer: playback.ctrl,
		Base:     2,
		Volume:   -2,
		Silent:   false,
	}

	done = make(chan bool)

	speaker.Init(format.SampleRate, format.SampleRate.N(time.Second/10))
	speaker.Play(beep.Seq(playback.volume, beep.Callback(func() {
		CloseAudio()
	})))

	fyne.Do(func() {
		totalSeconds := int(format.SampleRate.D(streamer.Len()).Round(time.Second).Seconds())
		minutes := totalSeconds / 60
		seconds := totalSeconds % 60

		currSongLengthText.SetText(fmt.Sprintf("%d:%02d", minutes, seconds))
		playbackSlider.Max = float64(streamer.Len()) / float64(playback.sampleRate)
	})

	for {
		select {
		case <-done:
			return
		case <-time.After(time.Second):
			speaker.Lock()

			fyne.Do(func() {
				totalSeconds := int(format.SampleRate.D(streamer.Position()).Round(time.Second).Seconds())

				minutes := totalSeconds / 60
				seconds := totalSeconds % 60

				currTimeText.SetText(fmt.Sprintf("%d:%02d", minutes, seconds))
				playbackSlider.Value = format.SampleRate.D(streamer.Position()).Round(time.Second).Seconds()
				playbackSlider.Refresh()
			})
			//fmt.Println(format.SampleRate.D(streamer.Position()).Round(time.Second))
			speaker.Unlock()
		}
	}
}

func MakePlaylistView(source []Audio) *fyne.Container {

	myPlaylist := widget.NewList(
		func() int { return len(source) },
		func() fyne.CanvasObject {
			return container.NewGridWithColumns(4, widget.NewLabel("#"), widget.NewLabel("Song"), widget.NewLabel("Artist"), widget.NewLabel("Album"))
		},
		func(i int, object fyne.CanvasObject) {
			row := object.(*fyne.Container)
			audio := source[i]

			row.Objects[0].(*widget.Label).SetText(fmt.Sprintf("%d", i+1))
			row.Objects[1].(*widget.Label).SetText(audio.GetName())
			row.Objects[2].(*widget.Label).SetText(audio.GetArtist())
			row.Objects[3].(*widget.Label).SetText(audio.GetAlbum())
		},
	)

	myPlaylist.OnSelected = func(id widget.ListItemID) {
		CloseAudio()

		go PlayAudio(playlist[id])
	}

	return container.NewBorder(
		container.NewGridWithColumns(4, widget.NewLabel("#"), widget.NewLabel("Song"), widget.NewLabel("Artist"), widget.NewLabel("Album")),
		nil, nil, nil,
		myPlaylist,
	)

}
