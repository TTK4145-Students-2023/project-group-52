package music

import (
	"os"
	"fmt"

	"github.com/hajimehoshi/oto/v2"

	"github.com/hajimehoshi/go-mp3"
)

func MusicPlayer(enableCh <-chan bool){
	f, err := os.Open("music/elevator_music.mp3")
	if err != nil {
		fmt.Printf("could not open file")
		return
	}
	defer f.Close()

	d, err := mp3.NewDecoder(f)
	if err != nil {
		return
	}

	c, ready, err := oto.NewContext(d.SampleRate(), 2, 2)
	if err != nil {
		return
	}
	<-ready


	p := c.NewPlayer(d)
	defer p.Close()

	for{
		select {
		case enable := <-enableCh:
			if enable {
				p.Play()
			} else {
				p.Reset()
			}
		}
	}
}
