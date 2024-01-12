package main

import (
	"dtmf-decoder/dtmf"
	"fmt"
	"log"
	"os"
	"os/exec"
)

//playlistURL := "output.m3u8"
//
//cmd := exec.Command("ffmpeg", "-i", playlistURL, "-vn", "-acodec", "pcm_s16le", "-ar", "44100", "-ac", "2", "-f", "wav", "output.wav")
//err := cmd.Run()
//if err != nil {
//	log.Fatal(err)
//}

func main() {
	path := "examples/test/output.wav"
	str, err := dtmf.DecodeDTMFFromFile(path, 7)
	if err != nil {
		fmt.Println(err)
	}

	cmd := exec.Command("ffplay", path)
	cmd.Stderr = os.Stdout
	err = cmd.Run()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Decoded DTMF:", str)
}
