package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
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
		Url:           "https://www632.ff-02.com/token=XzpWnlP2dW-QPShMJLf56Q/1620630799/129.205.0.0/167/8/f8/acc82f1223337de4e2ec11d3d2722f88-480p.mp4",
		TargerPath:    "Taoma.mp4",
		TotalSections: 20,
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

	eachSize := size / d.TotalSections

	var sections = make([][2]int, d.TotalSections)

	for i := range sections {
		if i == 0 {
			sections[i][0] = 0
		} else {
			sections[i][0] = sections[i-1][1] + 1
		}
		if i < d.TotalSections-1 {
			sections[i][1] = sections[i][0] + eachSize
		} else {
			sections[i][1] = size - 1
		}
	}

	fmt.Println(sections)

	for i, s := range sections {
		err = d.downloadSections(i, s)
		if err != nil {
			return err
		}
	}

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

func (d Download) downloadSections(i int, s [2]int) error {
	r, err := d.getNewRequest("GET")
	if err != nil {
		return err
	}

	fmt.Printf("bytes=%v-%v\n", s[0], s[1])

	r.Header.Set("Range", fmt.Sprintf("bytes=%v-%v", s[0], s[1]))

	resp, err := http.DefaultClient.Do(r)
	if err != nil {
		return err
	}

	fmt.Println(resp.Header.Get("Content-Length"))

	fmt.Printf("Downloaded %v bytes for section %v: %v\n", resp.Header.Get("Content-Length"), i, s)

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(fmt.Sprintf("section-%v.tmp", i), b, os.ModePerm)

	if err != nil {
		return err
	}

	return nil

}
