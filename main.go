package main

import (
	"flag"
	"fmt"
	"io"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"sync"
	"strings"

	"github.com/shirou/gopsutil/v3/disk"
)

var cameraDrive string
var destFolder string
var wg sync.WaitGroup

type CopyRequest struct {
	Source      string
	Destination string
}

var copychannel chan CopyRequest

func shouldCopy(req CopyRequest) bool {
	srcinfo, err := os.Stat(req.Source)
	if err != nil { // Do not copy if there is no source file
		return false
	}
	dstinfo, err := os.Stat(req.Destination)
	if err != nil { // Definite copy if there is no destination file
		return true
	}
	return srcinfo.Size() != dstinfo.Size()
}

func isImage(extension string) bool {
	normalized := strings.ToLower(extension)
	switch(normalized) {
	case ".jpeg":
		return true
	case ".jpg":
		return true
	case ".dng":
		return true
	default:
		return false
	}
	return false
}

// Copier receives a copy request and copies from source to destination.
func Copier() {
	var req CopyRequest
	for {
		req = <-copychannel

		pathname := filepath.Dir(req.Destination)
		info, err := os.Stat(pathname)
		if err != nil {
			log.Printf("Creating directory %s", pathname)
			os.MkdirAll(pathname, 0x755)
		} else if !info.IsDir() {
			log.Printf("%s not a directory, cannot copy", pathname)
			continue
		}

		if !shouldCopy(req) {
			wg.Done()
			continue
		}

		src, err := os.Open(req.Source)
		if err != nil {
			log.Printf("%v", err)
		}
		dst, err := os.Create(req.Destination)
		if err != nil {
			log.Printf("%v", err)
		}
		io.Copy(dst, src)
		dst.Close()
		src.Close()
		log.Printf("%s -> %s\n", req.Source, req.Destination)

		wg.Done()
	}
}

func findPictures(path string, d fs.DirEntry, err error) error {

	if d.Type().IsRegular() {
		if isImage(filepath.Ext(path)) {
			info, err := d.Info()
			if err != nil {
				return err
			}
			destfolder, err := filepath.Abs(filepath.Join(destFolder, info.ModTime().Format("2006-01-02")))
			if err != nil {
				return err
			}
			destfile := filepath.Join(destfolder, d.Name())

			wg.Add(1)

			var request CopyRequest
			request.Source = path
			request.Destination = destfile

			copychannel <- request
		}
	}

	return nil
}

func main() {

	var showHelp bool
	var showVersion bool
	const VERSION = "1.0.1"

	flag.StringVar(&destFolder, "dest", "", "Destination folder to copy images to")
	flag.BoolVar(&showHelp, "help", false, "Display help message")
	flag.BoolVar(&showVersion, "version", false, "Display version")
	flag.Parse()

	if showHelp {
		flag.PrintDefaults()
		os.Exit(1)
	}
	if showVersion {
		fmt.Printf("Photo Unloader version %s\n", VERSION)
		os.Exit(1)
	}

	partitions, err := disk.Partitions(false)
	if err != nil {
		fmt.Errorf("scanning partitions: %v", err)
	}

	for _, p := range partitions {
		dcimcandidate, err := filepath.Abs(filepath.Join(p.Mountpoint, "DCIM"))
		if err != nil {
			fmt.Errorf("%s: %v", p.Mountpoint, err)
		}
		fi, err := os.Stat(dcimcandidate)
		if err != nil {
			fmt.Errorf("%s: %v", dcimcandidate, err)
		}
		if fi != nil && fi.IsDir() {
			cameraDrive = dcimcandidate
		}
	}

	fmt.Printf("Camera folder %s", cameraDrive)

	copychannel = make(chan CopyRequest)
	go Copier()
	go Copier()
	go Copier()
	go Copier()

	filepath.WalkDir(cameraDrive, findPictures)

	wg.Wait()

}
