package main

import (
	"fmt"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/widget"
	"github.com/faiface/beep/mp3"
	"github.com/faiface/beep/speaker"
	"os"
	"time"
)

func main() {
	myApp := app.New()
	myWindow := myApp.NewWindow("Hello")
	myWindow.SetContent(widget.NewLabel("Hello World!"))

	myWindow.Show()

	DemoAudio()

	myApp.Run()
}

func DemoAudio() error {
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
