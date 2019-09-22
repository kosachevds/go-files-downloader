package downloader

import (
	"fmt"
	"io/ioutil"
	"strings"
)

type FileInfo struct {
	Filename string
	URL      string
}

func ReadInfos(filename string, separator string) ([]FileInfo, error) {
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
		info, err := ParseInfo(info)
		if err != nil {
			continue
		}
		result[resultCount] = *info
		resultCount++
	}
	return result[:resultCount], nil
}

func ParseInfo(strInfo string) (*FileInfo, error) {
	lastSpace := strings.LastIndex(strInfo, " ")
	if lastSpace < 0 {
		return nil, fmt.Errorf("wrong downloader.FileInfo format")
	}
	url := strings.TrimSpace(strInfo[lastSpace+1:])
	// TODO: trim
	filename := strings.TrimSpace(strInfo[:lastSpace])
	return &FileInfo{filename, url}, nil
}
