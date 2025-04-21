package main

import (
	"fmt"
	"fyne.io/fyne/v2"
	"github.com/faiface/beep"
	"github.com/faiface/beep/effects"
	"github.com/faiface/beep/mp3"
	"github.com/faiface/beep/speaker"
	"log"
	"os"
	"time"
)

var (
	playback  PlaybackController
	done      chan bool
	playlists []Playlist
)

type Playlist []Audio

// Audio represents an interface for audio-related metadata and file access.
// GetPath retrieves the file path of the audio.
// GetName retrieves the name/title of the audio.
// GetArtist retrieves the artist of the audio.
// GetAlbum retrieves the album associated with the audio.
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

func Seek(t float64) {
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

func TogglePause() {
	if playback.ctrl == nil {
		log.Println("No audio is currently playing.")
		return
	}

	speaker.Lock()
	playback.ctrl.Paused = !playback.ctrl.Paused
	speaker.Unlock()
}
