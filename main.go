package main

import (
	"fmt"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
	"os"
	"path/filepath"
	"strings"
)

var (
	playbackSlider     *widget.Slider
	currTimeText       *widget.Label
	currSongLengthText *widget.Label
	playlistView       *widget.List
)

func ScanDirectory(dir string, output *Playlist) error {
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
			*output = append(*output, &Song{full, file.Name(), "", ""})
		default:
			continue
		}
	}

	return nil
}

func main() {
	myApp := app.New()
	myWindow := myApp.NewWindow("GoListen")

	myWindow.Resize(fyne.NewSize(1920, 1080))

	localPlaylist := Playlist{}

	playlistContent := MakePlaylistView(&localPlaylist)

	playlists = append(playlists, localPlaylist)

	loadButton := widget.NewButton("Load", func() {
		dialog.ShowFolderOpen(func(uri fyne.ListableURI, err error) {
			if uri == nil {
				return
			}

			if err != nil {
				fmt.Println(err)
				return
			}

			CloseAudio()

			scanErr := ScanDirectory(uri.Path(), &localPlaylist)

			if scanErr != nil {
				fmt.Println(scanErr)
			}

			playlistView.Refresh()
		}, myWindow)
	})

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
				loadButton,
			),
			container.NewVScroll(playlistContent),
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
func MakePlaylistView(source *Playlist) *fyne.Container {

	playlistView = widget.NewList(
		func() int { return len(*source) },
		func() fyne.CanvasObject {
			return container.NewGridWithColumns(4, widget.NewLabel("#"), widget.NewLabel("Song"), widget.NewLabel("Artist"), widget.NewLabel("Album"))
		},
		func(i int, object fyne.CanvasObject) {
			row := object.(*fyne.Container)
			audio := (*source)[i]

			row.Objects[0].(*widget.Label).SetText(fmt.Sprintf("%d", i+1))
			row.Objects[1].(*widget.Label).SetText(audio.GetName())
			row.Objects[2].(*widget.Label).SetText(audio.GetArtist())
			row.Objects[3].(*widget.Label).SetText(audio.GetAlbum())
		},
	)

	playlistView.OnSelected = func(id widget.ListItemID) {
		CloseAudio()

		go PlayAudio((*source)[id])
	}

	return container.NewBorder(
		container.NewGridWithColumns(4, widget.NewLabel("#"), widget.NewLabel("Song"), widget.NewLabel("Artist"), widget.NewLabel("Album")),
		nil, nil, nil,
		playlistView,
	)

}
