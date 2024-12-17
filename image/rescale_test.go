package image_test

import (
	"bytes"
	"context"
	"image/jpeg"
	"os"
	"testing"

	"github.com/tutti-ch/backend-coding-task-template/image"
)

func TestRescale(t *testing.T) {
	t.Run(`rescale big image`, func(t *testing.T) {
		inputData, err := os.ReadFile(`../testdata/testimage_big.jpg`)
		if err != nil {
			t.Fatal(`failed to read testdata file`)
		}

		data, err := image.Rescale(context.Background(), inputData)
		if err != nil {
			t.Fatal(`rescale failed`)
		}

		if len(data) > 1_000_000 {
			t.Fatal(`output is larget than expected, probably not rescaled`)
		}

		imageConfig, err := jpeg.DecodeConfig(bytes.NewReader(data))
		if err != nil {
			t.Fatal(`reading rescaled image failed`)
		}
		if imageConfig.Width != 1920 {
			t.Fatal(`image width not as expected`, imageConfig.Width, `!=`, 1920)
		}
		if imageConfig.Height != 1280 {
			t.Fatal(`image height not as expected`, imageConfig.Height, `!=`, 1280)
		}
	})

	t.Run(`rescale small image`, func(t *testing.T) {
		inputData, err := os.ReadFile(`../testdata/testimage_small.jpg`)
		if err != nil {
			t.Fatal(`failed to read testdata file`)
		}

		data, err := image.Rescale(context.Background(), inputData)
		if err != nil {
			t.Fatal(`rescale failed`)
		}

		if !bytes.Equal(inputData, data) {
			t.Fatal(`output should be equal to input`)
		}
	})
}
