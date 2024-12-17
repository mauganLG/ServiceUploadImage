package image

import (
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"os"
)

// ProcessImage perfom the read of the image, rescale it if necessary
// and write the image on the disk on the filePath
func ProcessImage(filePath string, multipFile multipart.File) error {

	bytesFile, err := io.ReadAll(multipFile)
	if err != nil {
		return fmt.Errorf("ReadAll -> %s", err)
	}

	data, err := Rescale(context.Background(), bytesFile)
	if err != nil {
		return fmt.Errorf("Rescale -> %s", err)
	}
	file, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		return fmt.Errorf("OpenFile -> %s", err)
	}
	defer file.Close()

	if _, err := file.Write(data); err != nil {
		return fmt.Errorf("file write -> %s", err)
	}

	return nil

}
