package main

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"
)

type Download struct {
	Url           string
	TargerPath    string
	TotalSections int
}

func main() {
	startTime := time.Now()
	d := Download{
		Url:           "https://d.mandela.h.sabishare.com/dl/qGgRVKqeR89/cc8bf714619a570ad086fe1ce2d6e2bf69e9d2f3f12435d79cbab12461b05ced/Taaooma_-_Paul_The_Apprentice_(Part_2)_(NetNaija.com).mp4",
		TargerPath:    "Taoma.mp4",
		TotalSections: 10,
	}

	err := d.Do()
	if err != nil {
		log.Fatalf("An error occured while downloading the file: %s \n", err)
	}
	fmt.Printf("Download completed in %v seconds\n", time.Now().Sub(startTime).Seconds())
}

func (d Download) Do() error {
	fmt.Println(("Making connection"))
	r, err := d.getNewRequest("HEAD")
	if err != nil {
		return err
	}
	resp, err := http.DefaultClient.Do(r)
	if err != nil {
		return err
	}
	fmt.Printf("Got %v\n", resp.StatusCode)

	if resp.StatusCode > 299 {
		return errors.New(fmt.Sprintf("Can't process, response is %v", resp.StatusCode))
	}

	size, err := strconv.Atoi(resp.Header.Get("Content-Length"))
	if err != nil {
		return err
	}

	fmt.Printf("size is %v bytes\n", size)

	var sections = make([][2]int, d.TotalSections)

	return nil

}

func (d Download) getNewRequest(method string) (*http.Request, error) {
	r, err := http.NewRequest(method, d.Url, nil)

	if err != nil {
		return nil, err
	}
	r.Header.Set("User-Agent", "Downloader")
	return r, nil
}
