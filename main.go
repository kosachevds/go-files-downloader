package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"path"
	"strings"
	"time"
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
		downloadFromFileAsync(InputFile, Separator, 2)
		done <- true
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

func downloadFromFile(filename, separator string) {
	infos, err := readInfos(filename, separator)
	if err != nil {
		panic(err)
	}
	downloadAll(infos, Directory)
}

func downloadFromFileAsync(filename, separator string, maxSimultaneous int) {
	infos, err := readInfos(filename, separator)
	if err != nil {
		panic(err)
	}
	downloadAllAsync(infos, Directory, maxSimultaneous)
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
	url := strInfo[lastSpace+1:]
	filename := strInfo[:lastSpace]
	return &FileInfo{filename, url}, nil
}

func downloadFile(info *FileInfo, directory string) error {
	if _, err := os.Stat(directory); os.IsNotExist(err) {
		os.MkdirAll(directory, os.ModePerm)
	}
	var newFileName = path.Join(directory, info.filename)
	if _, err := os.Stat(newFileName); err == nil {
		return nil
	}
	if len(info.url) == 0 {
		return fmt.Errorf("empty url")
	}
	resp, err := http.Get(info.url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	fout, err := os.Create(newFileName)
	if err != nil {
		return err
	}
	defer fout.Close()

	_, err = io.Copy(fout, resp.Body)
	return err
}

func downloadAllAsync(files []FileInfo, directory string, maxSimultaneous int) {
	infosQueue := make(chan *FileInfo, maxSimultaneous)
	errors := make(chan error, maxSimultaneous)

	go func() {
		for fi := range infosQueue {
			go func(fi *FileInfo) {
				err := downloadFile(fi, directory)
				if err != nil {
					err = fmt.Errorf("%v error: %v\n", fi.filename, err)
				}
				errors <- err
			}(fi)
		}
		close(errors)
	}()
	for i, fi := range files {
		if i >= maxSimultaneous {
			err := <-errors
			if err != nil {
				appendToFile(EmptyUrlsFile, fmt.Sprint(err))
			}
		}
		infosQueue <- &fi
	}
	close(infosQueue)
	for err := range errors {
		if err != nil {
			appendToFile(EmptyUrlsFile, fmt.Sprint(err))
		}
	}
}

func downloadAll(files []FileInfo, directory string) {
	for i := range files {
		err := downloadFile(&files[i], directory)
		if err != nil {
			appendToFile(EmptyUrlsFile, fmt.Sprintf("%v error: %v\n", files[i].filename, err))
			// err = fmt.Errorf("Downloading %v error: %v", files[i].filename, err)
			// panic(err)
		}
	}
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
