package main

import (
	"fmt"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	"log"
	"os"
	"path/filepath"
	"strings"
)

var (
	playbackSlider     *widget.Slider
	currTimeText       *widget.Label
	currSongLengthText *widget.Label
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
	playbackSlider.OnChanged = Seek

	playbackProgress := container.NewBorder(
		nil,
		nil,
		currTimeText,
		currSongLengthText,
		playbackSlider,
	)

	playbackButton := widget.NewButton("Pause/Resume", TogglePause)

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

// MakePlaylistView creates a fyne.Container that displays a list of Audio items with details like Song, Artist, and Album.
// It allows audio selection and playback from the provided source.
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
