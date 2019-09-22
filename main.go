package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"strings"
	"time"

	// TODO: remade with absolute path
	"./internal/downloader"
)

const (
	Directory     = "files_"
	InputFile     = "test_tracks.txt"
	Separator     = ";"
	MsecDelay     = 400
	MaxPointCount = 10
	EmptyUrlsFile = "errors.txt"
)

type FileInfo struct {
	filename string
	url      string
}

func main() {
	done := make(chan bool)
	go func() {
		// downloadFromFile(InputFile, Separator)
		err := downloadAllFromFileSimultaneously(InputFile, Separator)
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

func downloadAllFromFileSimultaneously(filename, separator string) error {
	infos, err := readInfos(filename, separator)
	if err != nil {
		return err
	}
	err = downloadAllSymultaneously(infos, Directory)
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

func readInfos(filename string, separator string) ([]FileInfo, error) {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("reading file failed with %v\n", err)
	}
	strInfos := strings.Split(string(data), separator)
	result := make([]FileInfo, len(strInfos))
	resultCount := 0
	for _, info := range strInfos {
		if len(info) < 1 {
			continue
		}
		info, err := parseInfo(info)
		if err != nil {
			continue
		}
		result[resultCount] = *info
		resultCount++
	}
	return result[:resultCount], nil
}

func parseInfo(strInfo string) (*FileInfo, error) {
	lastSpace := strings.LastIndex(strInfo, " ")
	if lastSpace < 0 {
		return nil, fmt.Errorf("wrong FileInfo format")
	}
	url := strings.TrimSpace(strInfo[lastSpace+1:])
	// TODO: trim
	filename := strings.TrimSpace(strInfo[:lastSpace])
	return &FileInfo{filename, url}, nil
}

func appendToFile(filename, text string) error {
	fout, err := os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer fout.Close()
	_, err = fout.WriteString(text)
	return err
}

func downloadAllSymultaneously(infos []FileInfo, directory string) error {
	if _, err := os.Stat(directory); os.IsNotExist(err) {
		os.MkdirAll(directory, os.ModePerm)
	}
	urls := make([]string, len(infos))
	for i, fi := range infos {
		urls[i] = fi.url
	}
	bytes, err := downloader.DownloadFilesSimultaneously(urls)
	for i, fileBytes := range bytes {
		if fileBytes == nil {
			continue
		}
		fullPath := path.Join(directory, infos[i].filename)
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
