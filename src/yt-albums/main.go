package main

import (
	"bufio"
	"fmt"

	"sync"

	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"

	"strings"

	id3 "github.com/mikkyang/id3-go"
)

type UserInput struct {
	artist, album, url string
}

func main() {
	scanner := bufio.NewScanner(os.Stdin)

	var userInputs = []UserInput{}
	moreInput := true
	for moreInput {
		user := getInput(scanner)
		userInputs = append(userInputs, user)
		moreInput = confirmInput("Add another artist/album (Y/N): ", scanner)
	}

	var wg sync.WaitGroup
	wg.Add(len(userInputs))
	for _, user := range userInputs {

		path := makeDirs(user.artist, user.album)
		ytdlpath := path + "/%(title)s.%(ext)s"

		args := []string{
			"--extract-audio",
			"--audio-format", "mp3",
			"-i",
			"-o", ytdlpath,
			user.url}

		cmd := exec.Command("youtube-dl", args...)
		pipe, _ := cmd.StdoutPipe()
		cmd.Start()
		ytdl := bufio.NewScanner(pipe)
		for ytdl.Scan() {
			fmt.Println(string(ytdl.Text()))
		}

		fmt.Println("**Download Complete**: Spawning subroutine to set ID3 Tags")
		go updateID3Tags(&wg, path, user) //send additional args struct
	}
	fmt.Println("Waiting for subroutines to complete")
	wg.Wait()
	fmt.Println("Press enter to run again, Control-C to quit")
	scanner.Scan()
	main()
}

func updateID3Tags(wg *sync.WaitGroup, path string, user UserInput) {
	defer wg.Done()
	files, _ := ioutil.ReadDir(path)
	for _, f := range files {
		mp3File, _ := id3.Open(path + "/" + f.Name())
		mp3File.SetArtist(user.artist)
		mp3File.SetAlbum(user.album)
		mp3File.Close()
	}
	fmt.Println("**ID3 Tags Set** for " + user.artist + ": " + user.album)
}

func confirmInput(msg string, scanner *bufio.Scanner) bool {
	fmt.Print(msg)
	scanner.Scan()
	response := strings.TrimSpace(strings.ToLower(scanner.Text()))
	if response == "y" {
		return true
	} else {
		return false
	}

}

func makeDirs(parent, child string) string {
	parentPath := filepath.Join(".", parent)
	os.MkdirAll(parentPath, os.ModePerm)
	childPath := filepath.Join(".", parentPath, child)
	os.MkdirAll(childPath, os.ModePerm)
	return childPath
}

func getInput(scanner *bufio.Scanner) UserInput {

	fmt.Print("Enter artist name: ")
	scanner.Scan()
	artistName := scanner.Text()

	fmt.Print("Enter album name: ")
	scanner.Scan()
	albumName := scanner.Text()

	fmt.Print("Playlist URL: ")
	scanner.Scan()
	url := scanner.Text()

	return UserInput{artistName, albumName, url}

}
