package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/ndiecodes/downloader/internal"
)

func main() {
	startTime := time.Now()
	url := flag.String("u", "", "url to downloadable content")

	name := flag.String("n", "", "rename downloaded file")

	downloadDir := flag.String("d", "Downloads", "save file to another directory. eg Videos or Documents")

	flag.Parse()
	if *url == "" {
		flag.Usage()
		os.Exit(1)
	}

	dl := internal.Download{
		Url:           *url,
		DownloadDir:   *downloadDir,
		TargetPath:    *name,
		TotalSections: 10,
	}

	err := dl.Do()
	if err != nil {
		log.Fatal(err)
	}

	if err != nil {
		log.Fatalf("An error occured while downloading the file: %s \n", err)
	}
	fmt.Printf("Download completed in %v seconds\n", time.Since(startTime).Seconds())
}
