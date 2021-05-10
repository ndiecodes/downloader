package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"sync"
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
		TargerPath:    "taoma.mp4",
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
	var wg sync.WaitGroup
	for i, s := range sections {
		wg.Add(1)
		i := i
		s := s
		go func() {
			defer wg.Done()
			err = d.downloadSections(i, s)
			if err != nil {
				panic(err)
			}
		}()
	}

	wg.Wait()

	err = d.mergeFiles(sections)
	if err != nil {
		return err
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

func (d Download) mergeFiles(sections [][2]int) error {
	f, err := os.OpenFile(d.TargerPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, os.ModePerm)
	if err != nil {
		return err
	}
	defer f.Close()

	for i := range sections {
		b, err := ioutil.ReadFile(fmt.Sprintf("section-%v.tmp", i))

		if err != nil {
			return err
		}

		n, err := f.Write(b)

		if err != nil {
			return err
		}

		fmt.Printf("%v bytes merged\n", n)

	}

	return nil

}
