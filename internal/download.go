package internal

import (
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/user"
	"path"
	"strconv"
	"sync"
)

type Download struct {
	Url           string
	DownloadDir   string
	TargetPath    string
	tmpDir        string
	tmpFiles      []string //Slice length must match Total Sections
	TotalSections int
}

func (d *Download) setSavePath(path string) {
	downloadsDir, err := d.getDownloadPath()
	if err != nil {
		log.Fatal(err)
	}

	d.TargetPath = downloadsDir + path
}

func (d *Download) setTempDir(path string) {
	d.tmpDir = path
}

func (d *Download) setTmpFilesArray() {
	d.tmpFiles = make([]string, d.TotalSections)
}

func (d *Download) Do() error {

	dir, err := ioutil.TempDir("", "downloader")
	if err != nil {
		log.Fatal(err)
	}

	d.setTempDir(dir)

	defer os.RemoveAll(dir)

	d.setTmpFilesArray()

	fmt.Println(("Making connection------------"))
	r, err := d.getNewRequest("HEAD")
	if err != nil {
		return err
	}
	resp, err := http.DefaultClient.Do(r)
	if err != nil {
		return err
	}

	if resp.Header.Get("Accept-Ranges") != "bytes" {
		d.downloadLoneFile()
		return nil
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

	sections := d.computeSections(size)

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

func (d *Download) computeSections(size int) [][2]int {

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

	return sections

}

func (d *Download) downloadLoneFile() error {

	fmt.Println("Downloading Lone File.......")

	r, err := d.getNewRequest("GET")
	if err != nil {
		return err
	}

	resp, err := http.DefaultClient.Do(r)
	if err != nil {
		return err
	}

	f, err := os.Create(d.TargetPath)

	if err != nil {
		return err
	}

	_, err = io.Copy(f, resp.Body)

	if err != nil {
		return err
	}

	defer f.Close()

	return nil
}

func (d *Download) getNewRequest(method string) (*http.Request, error) {
	r, err := http.NewRequest(method, d.Url, nil)
	if d.TargetPath == "" {
		d.setSavePath(path.Base(r.URL.Path))
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

	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	file, err := ioutil.TempFile(d.tmpDir, "section.*.tmp")
	if err != nil {
		log.Fatal(err)
	}

	d.tmpFiles[i] = file.Name()

	err = ioutil.WriteFile(file.Name(), b, os.ModePerm)

	if err != nil {
		return err
	}

	return nil

}

func (d *Download) getDownloadPath() (string, error) {
	currentUser, err := user.Current()
	if err != nil {
		return "", err
	}
	downloads := fmt.Sprintf("%s/%s/", currentUser.HomeDir, d.DownloadDir)
	return downloads, nil
}

func (d *Download) mergeFiles(sections [][2]int) error {

	fmt.Println("Merging Temp Files--------------")
	f, err := os.Create(d.TargetPath)
	if err != nil {
		return err
	}
	defer f.Close()
	for i := range sections {
		b, err := ioutil.ReadFile(d.tmpFiles[i])

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
