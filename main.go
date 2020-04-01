package main

import (
	"fmt"
	"goffmpeg/media"
	"goffmpeg/transcoder"
)

func main() {
	file := "/root/Downloads/c.avi"

	ffvideo := media.FFVideo{}
	ffvideo.Init(file)
	_, err := ffvideo.GetInfo()
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(ffvideo)

	tran := transcoder.VideoTranscoder{}
	err = tran.ToMP4AsH264(file, "/root/Downloads/o.mp4", transcoder.VideoQuality_720)
	fmt.Println(err)
}
