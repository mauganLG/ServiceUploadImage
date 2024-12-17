package image

import (
	"bytes"
	"context"
	"fmt"
	"image"
	"image/jpeg"
)

// Rescale performs image downscaling (preserving Width/Height ratio) to a predefined width of 1920 pixels.
// If the image is already smaller than this then the scaling does not happen – the image is returned as-is.
// This is a naive implementation it
// - completely ignores exif metadata
// - doesn't care about jpeg colorspace or rotation
// - uses golang image/jpeg which is slow and incomplete
// - uses an extremely primitive scaling algorithm
func Rescale(_ context.Context, input []byte) ([]byte, error) {
	const maxTargetWidth = 1920

	imageConfig, err := jpeg.DecodeConfig(bytes.NewReader(input))
	if err != nil {
		return nil, fmt.Errorf("failed to decode image config: %w", err)
	}

	if imageConfig.Width <= maxTargetWidth {
		// Image is already small, do not modify it
		return input, nil
	}

	imInput, err := jpeg.Decode(bytes.NewReader(input))
	if err != nil {
		return nil, fmt.Errorf("failed to decode image: %w", err)
	}

	scaleFactor := maxTargetWidth / float64(imageConfig.Width)
	targetHeight := int(float64(imageConfig.Height) * scaleFactor)
	imOutput := image.NewRGBA(image.Rect(0, 0, maxTargetWidth, targetHeight))

	for x := 0; x < imOutput.Bounds().Dx(); x++ {
		for y := 0; y < imOutput.Bounds().Dy(); y++ {
			xSource := int(float64(x) * (1 / scaleFactor))
			ySource := int(float64(y) * (1 / scaleFactor))
			imOutput.Set(x, y, imInput.At(xSource, ySource))
		}
	}

	var outputBuf bytes.Buffer
	err = jpeg.Encode(&outputBuf, imOutput, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to encode output image: %w", err)
	}

	return outputBuf.Bytes(), nil
}
