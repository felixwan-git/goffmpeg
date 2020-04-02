package media

import (
	"fmt"
	"log"
	"os"
	"path"
	"strconv"
	"strings"

	"github.com/felixwan-git/goffmpeg/file"
	"github.com/felixwan-git/goffmpeg/utils"
)

type FFVideo struct {
	FilePath                  string
	Metadata                  MetadataInfo
	Audio                     AudioInfo
	Quality                   QualityInfo
	Duration, Bitrate, Format string
}

type MetadataInfo struct {
	MajorBrand, MinorVersion, CompatibleBrands, Encoder, Comment, CopyRight string
}

type AudioInfo struct {
	Format, Quality, Speed string
}

type QualityInfo struct {
	Width, Height, DARWidthScale, DARHeightScale int
	SAR, Speed, FPS, TBR, TBN, TBC               string
}

func (ffvideo *FFVideo) Init(filePath string) error {
	exist := file.Exists(filePath)
	if !exist {
		return fmt.Errorf("File not found %s", filePath)
	}

	ffvideo.FilePath = filePath
	return nil
}

func (ffvideo *FFVideo) GetInfo() (*FFVideo, error) {
	output, err, stderr := utils.ExecCommand(utils.FFMpegCommand, "-i", ffvideo.FilePath)
	if err != nil && !strings.Contains(stderr, "At least one output file must be specified") {
		return ffvideo, fmt.Errorf("Get video info failed, %s", err)
	}
	if output == "" {
		output = stderr
	}

	err = ffvideo.parse(output)
	if err != nil {
		return ffvideo, fmt.Errorf("Parse video info failed,%s output[%s]", err, output)
	}
	return ffvideo, nil
}

func (ffvideo *FFVideo) CutVideo(segmentFile, start, end string, args ...string) error {
	var err error
	dir := path.Dir(segmentFile)
	if !file.Exists(dir) {
		err = os.MkdirAll(dir, os.ModePerm)
	}
	if err != nil {
		return fmt.Errorf("segment file path invalid(%s), Path(%s)", err, segmentFile)
	}

	commands := []string{"-i", ffvideo.FilePath, "-ss", start, "-to", end}
	commands = append(commands, args...)
	commands = append(commands, segmentFile)
	output, err, stderr := utils.ExecCommand(utils.FFMpegCommand, commands...)
	if err != nil {
		return fmt.Errorf("cut video failed, FilePath[%s] segmentFile[%s],err[%s], out[%s], stderr[%s]", ffvideo.FilePath, segmentFile, err, output, stderr)
	}
	log.Printf("cut video done. output[%s] stdoutput[%s]", output, stderr)
	return nil
}

func (ffvideo *FFVideo) parse(data string) error {
	if data == "" {
		return fmt.Errorf("Data is null,parse failed")
	}

	data = "\n" + data
	data = data[strings.Index(data, "\nInput "):]
	ffvideo.parseSelf(data)
	ffvideo.parseMetadata(data)
	ffvideo.parseAudio(data)
	ffvideo.parseQuality(data)
	return nil
}

func (ffvideo *FFVideo) parseSelf(data string) {
	data = strings.Split(data, "  Duration:")[1]
	data = strings.ReplaceAll(data, " ", "")
	fmt.Println(data)
	ffvideo.Duration = data[:strings.Index(data, ",")]
	ffvideo.Bitrate = data[strings.Index(data, "bitrate:")+8 : strings.Index(data, "\n")]
	ffvideo.Format = data
	if strings.Contains(data, "Video:") {
		ffvideo.Format = data[strings.Index(data, "Video:")+6:]
		ffvideo.Format = ffvideo.Format[:strings.Index(ffvideo.Format, ",")]
	}
}

func (ffvideo *FFVideo) parseMetadata(data string) {
	start := strings.Index(data, "\n  Metadata:")
	num := strings.Index(data, "\n  Duration:")
	if start < 0 || num < 0 {
		return
	}

	ffvideo.Metadata = MetadataInfo{}
	data = strings.Trim(data[start:num], " ")
	array := strings.Split(data, "\n")
	for _, item := range array {
		if strings.Contains(item, "major_brand") {
			ffvideo.Metadata.MajorBrand = strings.Split(item, ":")[1]
		}

		if strings.Contains(item, "minor_version") {
			ffvideo.Metadata.MinorVersion = strings.Split(item, ":")[1]
		}

		if strings.Contains(item, "compatible_brands") {
			ffvideo.Metadata.CompatibleBrands = strings.Split(item, ":")[1]
		}

		if strings.Contains(item, "encoder") {
			ffvideo.Metadata.Encoder = strings.Split(item, ":")[1]
		}

		if strings.Contains(item, "comment") {
			ffvideo.Metadata.Comment = strings.Split(item, ":")[1]
		}

		if strings.Contains(item, "copyright") {
			ffvideo.Metadata.CopyRight = strings.Split(item, ":")[1]
		}
	}
}

func (ffvideo *FFVideo) parseAudio(data string) {
	if !strings.Contains(data, " Audio:") {
		return
	}

	ffvideo.Audio = AudioInfo{}
	data = strings.Split(data, " Audio:")[1]
	data = strings.Trim(data[0:strings.Index(data, "\n")], " ")
	array := strings.Split(data, ",")
	ffvideo.Audio.Format = array[0]
	ffvideo.Audio.Quality = array[1]
	if strings.Contains(data, "kb/s") {
		ffvideo.Audio.Speed = array[4]
	}
}

func (ffvideo *FFVideo) parseQuality(data string) {
	if !strings.Contains(data, " Video:") {
		return
	}

	ffvideo.Quality = QualityInfo{}
	data = strings.Split(data, " Video:")[1]
	data = strings.ReplaceAll(data[0:strings.Index(data, "\n")], " ", "")
	array := strings.Split(data, ",")
	str := array[2]
	if !strings.Contains(str, "x") && (strings.Contains(str, ")") || strings.Contains(str, "(")) {
		data = strings.ReplaceAll(data, ","+str, str)
		array = strings.Split(data, ",")
	}

	for _, item := range array {
		width, err := strconv.Atoi(strings.Split(item, "x")[0])
		if strings.Contains(item, "x") && err == nil {
			ffvideo.Quality.Width = width
			ffvideo.Quality.Height, _ = strconv.Atoi(strings.Split(strings.Split(item, "[")[0], "x")[1])
		}
		if strings.Contains(item, "SAR") {
			ffvideo.Quality.SAR = item[strings.Index(item, "SAR")+3 : strings.Index(item, "DAR")]
			item = strings.ReplaceAll(item, "]", "")
			dar := item[strings.Index(item, "DAR")+3:]
			darArray := strings.Split(dar, ":")
			ffvideo.Quality.DARWidthScale, _ = strconv.Atoi(darArray[0])
			ffvideo.Quality.DARHeightScale, _ = strconv.Atoi(darArray[1])
		}
		if strings.Contains(item, "kb/s") {
			ffvideo.Quality.Speed = item
		}
		if strings.Contains(item, "fps") {
			ffvideo.Quality.FPS = strings.Split(strings.Trim(item, " "), "f")[0]
		}
		if strings.Contains(item, "tbr") {
			ffvideo.Quality.TBR = strings.Split(strings.Trim(item, " "), "t")[0]
		}
		if strings.Contains(item, "tbn") {
			ffvideo.Quality.TBN = strings.Split(strings.Trim(item, " "), "t")[0]
		}
		if strings.Contains(item, "tbc") {
			ffvideo.Quality.TBC = strings.Split(strings.Trim(item, " "), "t")[0]
		}
	}
}
