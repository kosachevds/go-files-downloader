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
	err = downloadAllSymultaneously(infos, resultDir)
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

func downloadAllSymultaneously(infos []downloader.FileInfo, directory string) error {
	if _, err := os.Stat(directory); os.IsNotExist(err) {
		os.MkdirAll(directory, os.ModePerm)
	}
	urls := make([]string, len(infos))
	for i, fi := range infos {
		urls[i] = fi.URL
	}
	// TODO: with max simultaneous limit
	// TODO: fix names
	bytes, err := downloader.DownloadFilesSimultaneously(urls)
	for i, fileBytes := range bytes {
		if fileBytes == nil {
			continue
		}
		fullPath := path.Join(directory, infos[i].Filename)
		file, err := os.OpenFile(fullPath, os.O_CREATE, 0644)
		if err != nil {
			return err
		}
		defer file.Close()

		_, err = file.Write(fileBytes)
		if err != nil {
			return nil
		}
	}
	return err
}
