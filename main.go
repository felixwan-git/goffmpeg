package main

import (
	"fmt"

	"github.com/felixwan-git/goffmpeg/media"
	"github.com/felixwan-git/goffmpeg/transcoder"
)

func main() {
	file := "/root/Downloads/d.wmv"

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
	err = tran.ToM3U8("/root/Downloads/o.mp4", "/root/Downloads/o/o.m3u8")
	fmt.Println(err)
}
