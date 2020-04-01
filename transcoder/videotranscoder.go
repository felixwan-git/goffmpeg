package transcoder

import (
	"fmt"
	"log"
	"os"
	"path"
	"strconv"
	"strings"

	"github.com/felixwan-git/goffmpeg/file"
	"github.com/felixwan-git/goffmpeg/media"
	"github.com/felixwan-git/goffmpeg/utils"
)

type VideoTranscoder struct {
}

type VideoQuality int

const (
	VideoQuality_Origin VideoQuality = iota
	VideoQuality_720
	VideoQuality_1080
	VideoQuality_2K
	VideoQuality_4K
	VideoQuality_8K
)

type Size struct {
	Width, Height int
}

func (videoTran *VideoTranscoder) ToMP4AsH264(inputFile, outputFile string, quality VideoQuality) error {
	ffvideo := new(media.FFVideo)
	ffvideo.Init(inputFile)
	_, err := ffvideo.GetInfo()
	if err != nil {
		return fmt.Errorf("Convert to mp4 failed, get video info failed, %s", err)
	}

	size, err := zoomSize(quality, ffvideo)
	if err != nil {
		return fmt.Errorf("zoom size error(%s)", err)
	}
	return videoTran.ToMP4AsH264ForArgs(inputFile, outputFile, "-vf", "scale=-2:"+strconv.Itoa(size.Height))
}

func (videoTran *VideoTranscoder) ToMP4AsH264ForArgs(inputFile, outputFile string, argnuments ...string) error {
	if outputFile == "" || inputFile == "" || !file.Exists(inputFile) {
		return fmt.Errorf("file path invalid, inputFile: %s, outputFile: %s", inputFile, outputFile)
	}

	log.Println("to mp4 in video")
	dir := path.Dir(outputFile)
	if !file.Exists(dir) {
		os.MkdirAll(dir, os.ModePerm)
	}

	commands := []string{"-i", inputFile, "-y", "-c:v", "libx264", "-strict", "-2"}
	commands = append(commands, argnuments...)
	commands = append(commands, outputFile)
	output, err, stderr := utils.ExecCommand(utils.FFMpegCommand, commands...)
	if err != nil {
		return fmt.Errorf("convert to mp4 failed, inputFile[%s] outputFile[%s],err[%s], out[%s], stderr[%s]", inputFile, outputFile, err, output, stderr)
	}
	log.Printf("convert mp4 done. output[%s] stdoutput[%s]", output, stderr)
	return nil
}

func (videoTran *VideoTranscoder) ToM3U8(inputFile, outputFile string) error {
	return videoTran.ToM3U8ForSegment(inputFile, outputFile, 6)
}

func (videoTran *VideoTranscoder) ToM3U8ForSegment(inputFile, outputFile string, segmentSeconds int) error {
	if outputFile == "" || inputFile == "" || !file.Exists(inputFile) {
		return fmt.Errorf("file path invalid, inputFile: %s, outputFile: %s", inputFile, outputFile)
	}
	if ext := path.Ext(inputFile); ext != ".mp4" {
		fmt.Errorf("video format isn't mp4 ")
	}
	dir := path.Dir(outputFile)
	if !file.Exists(dir) {
		os.MkdirAll(dir, os.ModePerm)
	}
	fileName := strings.TrimSuffix(path.Base(outputFile), path.Ext(outputFile))
	tsSegmentName := dir + "/" + fileName + "-%03d.ts"
	tsTempName := dir + "/" + fileName + ".ts"
	if file.Exists(tsTempName) {
		os.Remove(tsTempName)
	}

	commands := []string{"-i", tsTempName, "-c", "copy", "-map", "0", "-f", "segment", "-segment_list", outputFile, "-segment_time", strconv.Itoa(segmentSeconds), tsSegmentName}
	err := videoTran.toTs(inputFile, tsTempName)
	if err != nil {
		return fmt.Errorf("to m3u8 failed(%s)", err)
	}
	log.Println("convert video ts done")

	log.Printf("convert video M3U8  data[%s]", commands)
	out, err, stderr := utils.ExecCommand(utils.FFMpegCommand, commands...)
	if err != nil {
		return fmt.Errorf("convert video m3u8 failed, out[%s], stderr[%s]", out, stderr)
	}
	log.Println("convert video M3U8 done")

	os.Remove(tsTempName)
	return nil
}

func (videoTran *VideoTranscoder) toTs(inputFile, tsTempName string) error {
	commands := []string{"-y", "-i", inputFile, "-vcodec", "copy", "-acodec", "copy", "-vbsf", "h264_mp4toannexb", tsTempName}
	log.Printf("convert video TS data[&s]", commands)
	out, err, stderr := utils.ExecCommand(utils.FFMpegCommand, commands...)
	log.Printf("convert video ts, out[%s], stderr[%s]", out, stderr)
	if err == nil {
		return nil
	}
	log.Printf("convert ts failed, try to h264")

	commands = []string{"-y", "-i", inputFile, "-vcodec", "copy", "-acodec", "copy", "-bsf:v", "h264_mp4toannexb", tsTempName}
	log.Printf("convert video TS  data[%s]", commands)
	out, err, stderr = utils.ExecCommand(utils.FFMpegCommand, commands...)
	log.Printf("convert video ts, out[%s], stderr[%s]", out, stderr)
	return nil
}

func zoomSize(quality VideoQuality, ffvideo *media.FFVideo) (Size, error) {
	if ffvideo.Quality == (media.QualityInfo{}) {
		return Size{}, fmt.Errorf("video quality info is nil")
	}
	if ffvideo.Quality.Width == 0 || ffvideo.Quality.Height == 0 {
		return Size{}, fmt.Errorf("video quality size is zero")
	}
	if VideoQuality_2K <= quality {
		return Size{}, fmt.Errorf("notsupport 2k level resize")
	}

	var heightQuality int
	switch quality {
	case VideoQuality_Origin:
		heightQuality = ffvideo.Quality.Height
		break
	case VideoQuality_720:
		heightQuality = 720
		break
	case VideoQuality_1080:
		heightQuality = 1080
		break
	}

	if ffvideo.Quality.Height <= heightQuality {
		return Size{ffvideo.Quality.Width, ffvideo.Quality.Height}, nil
	}

	width := ffvideo.Quality.Width * heightQuality / ffvideo.Quality.Height
	return Size{width, heightQuality}, nil
}
