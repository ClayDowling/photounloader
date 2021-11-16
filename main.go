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

// Copier receives a copy request and copies from source to destination.
func Copier() {
	var req CopyRequest
	for {
		req = <-copychannel
		src, err := os.Open(req.Source)
		if err != nil {
			log.Printf("%v", err)
		}

		pathname := filepath.Dir(req.Destination)
		info, err := os.Stat(pathname)
		if err != nil {
			log.Printf("Creating directory %s", pathname)
			os.MkdirAll(pathname, 0x755)
		} else if !info.IsDir() {
			log.Printf("%s not a directory, cannot copy", pathname)
		}

		dst, err := os.Create(req.Destination)
		if err != nil {
			log.Printf("%v", err)
		}
		io.Copy(dst, src)
		dst.Close()
		src.Close()

		wg.Done()
	}
}

func findPictures(path string, d fs.DirEntry, err error) error {

	if d.Type().IsRegular() {
		if filepath.Ext(path) == ".DNG" {
			info, err := d.Info()
			if err != nil {
				return err
			}
			destfolder, err := filepath.Abs(filepath.Join(destFolder, info.ModTime().Format("2006-01-02")))
			if err != nil {
				return err
			}
			destfile := filepath.Join(destfolder, d.Name())

			fmt.Printf("%s -> %s\n", path, destfile)

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

	flag.StringVar(&destFolder, "dest", "", "Destinatin folder to copy images to")
	flag.Parse()

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

	filepath.WalkDir(cameraDrive, findPictures)

	wg.Wait()

}
