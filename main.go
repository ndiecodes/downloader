package main

import (
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path"
	"strconv"
	"sync"
	"time"
)

type Download struct {
	Url           string
	TargerPath    string
	tmpDir        string
	tempFiles     []string
	TotalSections int
}

func main() {
	startTime := time.Now()
	url := flag.String("u", "", "url to downloadable content")

	name := flag.String("n", "", "rename downloaded file")

	flag.Parse()
	if *url == "" {
		flag.Usage()
		os.Exit(1)
	}

	d := Download{
		Url:           *url,
		TargerPath:    *name,
		tmpDir:        "",
		tempFiles:     make([]string, 10),
		TotalSections: 10,
	}

	dir, err := ioutil.TempDir("", "downloader")
	if err != nil {
		log.Fatal(err)
	}

	d.setTempDir(dir)

	defer os.RemoveAll(dir)

	err = d.Do()
	if err != nil {
		log.Fatalf("An error occured while downloading the file: %s \n", err)
	}
	fmt.Printf("Download completed in %v seconds\n", time.Now().Sub(startTime).Seconds())
}

func (d *Download) setPath(path string) {
	d.TargerPath = path
}

func (d *Download) setTempDir(path string) {
	d.tmpDir = path
}

func (d *Download) Do() error {
	fmt.Println(("Making connection------------"))
	r, err := d.getNewRequest("HEAD")
	if err != nil {
		return err
	}
	resp, err := http.DefaultClient.Do(r)
	if err != nil {
		return err
	}
	fmt.Printf("Status: %v\n", resp.StatusCode)

	if resp.StatusCode > 299 {
		return errors.New(fmt.Sprintf("Can't process, response is %v", resp.StatusCode))
	}

	size, err := strconv.Atoi(resp.Header.Get("Content-Length"))
	if err != nil {
		return err
	}

	fmt.Printf("Total size: %v bytes\n", size)

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

func (d *Download) getNewRequest(method string) (*http.Request, error) {
	r, err := http.NewRequest(method, d.Url, nil)
	if d.TargerPath == "" {
		d.setPath(path.Base(r.URL.Path))
	}

	if err != nil {
		return nil, err
	}
	r.Header.Set("User-Agent", "Downloader")
	return r, nil
}

func (d *Download) downloadSections(i int, s [2]int) error {
	r, err := d.getNewRequest("GET")
	if err != nil {
		return err
	}

	r.Header.Set("Range", fmt.Sprintf("bytes=%v-%v", s[0], s[1]))

	resp, err := http.DefaultClient.Do(r)
	if err != nil {
		return err
	}

	fmt.Printf("Downloaded %v bytes for section %v: %v\n", resp.Header.Get("Content-Length"), i, s)

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	file, err := ioutil.TempFile(d.tmpDir, "section.*.tmp")
	if err != nil {
		log.Fatal(err)
	}

	d.tempFiles[i] = file.Name()

	err = ioutil.WriteFile(file.Name(), b, os.ModePerm)

	if err != nil {
		return err
	}

	return nil

}

func (d *Download) mergeFiles(sections [][2]int) error {

	fmt.Println("Merging Temp Files--------------")
	f, err := os.OpenFile(d.TargerPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, os.ModePerm)
	if err != nil {
		return err
	}
	defer f.Close()
	for i := range sections {
		b, err := ioutil.ReadFile(d.tempFiles[i])

		if err != nil {
			return err
		}

		_, err = f.Write(b)

		if err != nil {
			return err
		}

	}
	return nil

}
