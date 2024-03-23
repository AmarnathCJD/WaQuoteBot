package main

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"image"
	"image/draw"
	"os"
	"os/exec"
	"path"
	"strconv"

	"github.com/kolesa-team/go-webp/decoder"
	"github.com/kolesa-team/go-webp/encoder"
	"github.com/kolesa-team/go-webp/webp"
)

func WebpImagePad(inputData []byte, wPad, hPad int, updateId int64) ([]byte, error) {
	webpDecoder, err := decoder.NewDecoder(bytes.NewBuffer(inputData), &decoder.Options{NoFancyUpsampling: true})
	if err != nil {
		return nil, fmt.Errorf("failed to create a webp decoder: %s", err)
	}

	inputImage, err := webpDecoder.Decode()
	if err != nil {
		return nil, fmt.Errorf("failed to decode webp image: %s", err)
	}

	var (
		wOffset = wPad / 2
		hOffset = hPad / 2
	)

	outputWidth := inputImage.Bounds().Dx() + wPad
	outputHeight := inputImage.Bounds().Dy() + hPad

	outputImage := image.NewRGBA(image.Rect(0, 0, outputWidth, outputHeight))
	draw.Draw(outputImage, image.Rect(wOffset, hOffset, outputWidth-wOffset, outputHeight-hOffset), inputImage, image.Point{}, draw.Src)

	encoderOptions, err := encoder.NewLossyEncoderOptions(encoder.PresetDefault, 100)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize encoder options: %s", err)
	}

	var outputBuffer bytes.Buffer
	if err = webp.Encode(&outputBuffer, outputImage, encoderOptions); err != nil {
		return nil, fmt.Errorf("failed to encode into webp: %s", err)
	}

	if outputData, err := WebpWriteExifData(outputBuffer.Bytes(), updateId); err == nil {
		return outputData, nil
	}

	return outputBuffer.Bytes(), nil
}

func WebpWriteExifData(inputData []byte, updateId int64) ([]byte, error) {
	var (
		startingBytes = []byte{0x49, 0x49, 0x2A, 0x00, 0x08, 0x00, 0x00, 0x00, 0x01, 0x00, 0x41, 0x57, 0x07, 0x00}
		endingBytes   = []byte{0x16, 0x00, 0x00, 0x00}
		b             bytes.Buffer

		currUpdateId = strconv.FormatInt(updateId, 10)
		currPath     = path.Join("downloads", currUpdateId)
		inputPath    = path.Join(currPath, "input_exif.webm")
		outputPath   = path.Join(currPath, "output_exif.webp")
		exifDataPath = path.Join(currPath, "raw.exif")
	)

	if _, err := b.Write(startingBytes); err != nil {
		return nil, err
	}

	jsonData := map[string]interface{}{
		"sticker-pack-id":        "wabot.roseloverx.",
		"sticker-pack-name":      "Valeri - Quoted Sticker",
		"sticker-pack-publisher": "Valeri",
		"emojis":                 []string{"😀"},
	}
	jsonBytes, err := json.Marshal(jsonData)
	if err != nil {
		return nil, err
	}

	jsonLength := (uint32)(len(jsonBytes))
	lenBuffer := make([]byte, 4)
	binary.LittleEndian.PutUint32(lenBuffer, jsonLength)

	if _, err := b.Write(lenBuffer); err != nil {
		return nil, err
	}
	if _, err := b.Write(endingBytes); err != nil {
		return nil, err
	}
	if _, err := b.Write(jsonBytes); err != nil {
		return nil, err
	}

	if err := os.MkdirAll(currPath, os.ModePerm); err != nil {
		return nil, err
	}
	defer os.RemoveAll(currPath)

	if err := os.WriteFile(inputPath, inputData, os.ModePerm); err != nil {
		return nil, err
	}
	if err := os.WriteFile(exifDataPath, b.Bytes(), os.ModePerm); err != nil {
		return nil, err
	}

	cmd := exec.Command("webpmux",
		"-set", "exif",
		exifDataPath, inputPath,
		"-o", outputPath,
	)

	if err := cmd.Run(); err != nil {
		fmt.Println(err)
		return nil, err
	}

	return os.ReadFile(outputPath)
}

func ind(x string) *string {
	return &x
}

func inb(x bool) *bool {
	return &x
}

func ini(x int) *uint32 {
	y := uint32(x)
	return &y
}
