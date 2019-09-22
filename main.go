package main

import (
	"fmt"
	"os"
	"os/exec"
	"path"
	"time"

	// TODO: remade with absolute path
	"./internal/downloader"
)

const (
	ResultDir     = "files_"
	InputFile     = "test_tracks.txt"
	Separator     = ";"
	MsecDelay     = 400
	MaxPointCount = 10
	EmptyUrlsFile = "errors.txt"
)

func main() {
	done := make(chan bool)
	go func() {
		// downloadFromFile(InputFile, Separator)
		err := downloadAllFromFileSimultaneously(InputFile, Separator, ResultDir)
		done <- true
		if err != nil {
			panic(err)
		}
	}()

	printPoints(MaxPointCount, MsecDelay, done)

	fmt.Println("Done!")
}

func printPoints(maxPointsCount, msecDelay int, done <-chan bool) {
	curentCount := 0
	delayDuration := time.Duration(msecDelay) * time.Millisecond
	for {
		select {
		case <-done:
			{
				_ = clearConsole()
				return
			}
		default:
			{
				if curentCount == maxPointsCount {
					_ = clearConsole()
					curentCount = 0
				}
				fmt.Print(".")
				curentCount++
			}
		}
		time.Sleep(delayDuration)
	}
}

func downloadAllFromFileSimultaneously(filename, separator, resultDir string) error {
	infos, err := downloader.ReadInfos(filename, separator)
	if err != nil {
		return err
	}
	if _, err := os.Stat(resultDir); os.IsNotExist(err) {
		os.MkdirAll(resultDir, os.ModePerm)
	}
	infos = addToFilename(resultDir, infos)
	err = downloader.DownloadFilesSimultaneous(infos)
	if err != nil {
		return err
	}
	return nil
}

func clearConsole() error {
	command := exec.Command("cmd", "/c", "cls")
	command.Stdout = os.Stdout
	return command.Run()
}

func addToFilename(pathPart string, infos []downloader.FileInfo) []downloader.FileInfo {
	for i := range infos {
		filename := infos[i].Filename
		infos[i].Filename = path.Join(pathPart, filename)
	}
	return infos
}
