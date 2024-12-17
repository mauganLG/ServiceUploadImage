package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"sync"
)

// ResponseFile define the format of Response sended by the server
// comnined with the http status received
type ResponseFile struct {
	Status   string    `json:"status"`
	Response *Response `json:"response"`
}

// return a string representation in json format of the struture
func (r *ResponseFile) PrintJsonFormat() string {
	jsonData, err := json.MarshalIndent(r, "", "  ")
	if err != nil {
		return ""
	}

	return string(jsonData)
}

// Response define the format of Response sended by the server
type Response struct {
	Error   string `json:"error,omitempty"`
	ImageID string `json:"image_id,omitempty"`
}

// Add the elements needed to the image to be enoded in multipart/form-data
func makeMultipartRequest(data []byte) (*bytes.Buffer, string, error) {

	var buf bytes.Buffer

	mw := multipart.NewWriter(&buf)

	w, err := mw.CreateFormFile(`image`, `image.jpg`)
	if err != nil {
		return nil, "", err
	}

	_, err = w.Write(data)
	if err != nil {
		return nil, "", err
	}

	err = mw.Close()
	if err != nil {
		return nil, "", err
	}

	return &buf, mw.FormDataContentType(), nil
}

// testFiles send the file defined in filePath and send to the localhost server
// port 8888 on the upload path and return a ResponseFile
// or an error
func testFiles(filePath string) (*ResponseFile, error) {
	inputData, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read %s file\n", filePath)

	}

	buf, contentType, err := makeMultipartRequest(inputData)
	if err != nil {
		return nil, fmt.Errorf("Error : %s\n", err.Error())

	}

	req, err := http.NewRequest("POST", "http://localhost:8888/upload", buf)
	if err != nil {
		return nil, fmt.Errorf("Error : %s\n", err.Error())

	}

	req.Header.Set("Content-Type", contentType)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("Error sending request: %s\n", err)

	}
	defer resp.Body.Close()

	response := &Response{}

	respBody, _ := io.ReadAll(resp.Body)
	json.Unmarshal(respBody, response)

	rpFile := &ResponseFile{
		Status:   resp.Status,
		Response: response,
	}

	return rpFile, nil
}

func main() {

	fileSmall := "testdata/testimage_small.jpg"

	fmt.Printf("FILE : %s\n", fileSmall)
	r, err := testFiles(fileSmall)
	if err != nil {
		fmt.Fprintf(os.Stderr, err.Error())
	}
	fmt.Printf(r.PrintJsonFormat())

	fileSmallPng := "testdata/testimage_small.png"
	fmt.Printf("FILE : %s\n", fileSmallPng)
	r, err = testFiles(fileSmallPng)
	if err != nil {
		fmt.Fprintf(os.Stderr, err.Error())
	}
	fmt.Printf(r.PrintJsonFormat())

	fileBig := "testdata/testimage_big.jpg"
	fmt.Printf("FILE : %s\n", fileBig)
	r, err = testFiles(fileBig)
	if err != nil {
		fmt.Fprintf(os.Stderr, err.Error())
	}
	fmt.Printf(r.PrintJsonFormat())

	fileText := "testdata/test_file.txt"
	fmt.Printf("FILE : %s\n", fileText)
	r, err = testFiles(fileText)
	if err != nil {
		fmt.Fprintf(os.Stderr, err.Error())
	}
	fmt.Printf(r.PrintJsonFormat())

	fmt.Printf("FILE Range : %s\n", fileSmall)

	nbTask := 1000000

	resultRf := make(chan *ResponseFile, nbTask)

	var wg sync.WaitGroup
	for range nbTask {
		wg.Add(1)
		go func() {
			defer wg.Done()
			r, err := testFiles(fileSmall)
			if err != nil {
				fmt.Fprintf(os.Stderr, err.Error())
			}
			resultRf <- r
		}()
	}

	go func() {
		wg.Wait()
		close(resultRf)
	}()

	for r := range resultRf {
		fmt.Printf(r.PrintJsonFormat())
	}
}
