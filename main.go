package main

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/shirou/gopsutil/v3/disk"
)

var cameraDrive string

func findPictures(path string, d fs.DirEntry, err error) error {

	if d.Type().IsRegular() {
		if filepath.Ext(path) == ".DNG" {
			info, err := d.Info()
			if err != nil {
				fmt.Errorf("%s: %v", path, err)
			}
			fmt.Printf("%s %v\n", path, info.ModTime())
		}
	}

	return nil
}

func main() {
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
	filepath.WalkDir(cameraDrive, findPictures)

}
