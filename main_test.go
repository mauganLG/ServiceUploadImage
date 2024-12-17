package main

import (
	"bytes"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"sync"
	"testing"
)

func makeMultipartRequest(t *testing.T, data []byte) (*bytes.Buffer, string) {
	t.Helper()
	var buf bytes.Buffer

	mw := multipart.NewWriter(&buf)
	if err := mw.SetBoundary(`xYzZY`); err != nil {
		t.Fatal(err)
	}

	w, err := mw.CreateFormFile(`image`, `image.jpg`)
	if err != nil {
		t.Fatal(err)
	}

	_, err = w.Write(data)
	if err != nil {
		t.Fatal(err)
	}

	err = mw.Close()
	if err != nil {
		t.Fatal(err)
	}

	return &buf, mw.FormDataContentType()
}

// TestMakeHandler is a very basic test that does not cover all of the specification and should
// only serve as a starting point.
func TestMakeHandlerGood(t *testing.T) {

	fsDirectory, err := os.MkdirTemp(``, `test-make-handler`)
	if err != nil {
		t.Fatalf(`failed to create temporary directory`)
	}

	defer os.RemoveAll(fsDirectory)

	handler := MakeHandler(fsDirectory, 1)

	t.Run(`small image`, func(t *testing.T) {
		inputData, err := os.ReadFile(`testdata/testimage_small.jpg`)
		if err != nil {
			t.Fatal(`failed to read testdata file`)
		}

		buf, contentType := makeMultipartRequest(t, inputData)
		req := httptest.NewRequest(`POST`, `/upload`, buf)
		req.Header.Set(`Content-Type`, contentType)
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf(`unexpected status code %v`, w.Code)
		}
	})
}

func TestMakeHandlerTooBig(t *testing.T) {

	fsDirectory, err := os.MkdirTemp(``, `test-make-handler`)
	if err != nil {
		t.Fatalf(`failed to create temporary directory`)
	}

	defer os.RemoveAll(fsDirectory)

	handler := MakeHandler(fsDirectory, 1)
	t.Run(`too big image`, func(t *testing.T) {
		inputData, err := os.ReadFile(`testdata/testimage_big.jpg`)
		if err != nil {
			t.Fatal(`failed to read testdata file`)
		}

		buf, contentType := makeMultipartRequest(t, inputData)
		req := httptest.NewRequest(`POST`, `/upload`, buf)
		req.Header.Set(`Content-Type`, contentType)
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)

		if w.Code != http.StatusRequestEntityTooLarge {
			t.Fatalf(`unexpected status code %v`, w.Code)
		}
	})

}

func TestMakeHandlerNotAnJpeg(t *testing.T) {

	fsDirectory, err := os.MkdirTemp(``, `test-make-handler`)
	if err != nil {
		t.Fatalf(`failed to create temporary directory`)
	}

	defer os.RemoveAll(fsDirectory)

	handler := MakeHandler(fsDirectory, 1)
	t.Run(`not JPEG format`, func(t *testing.T) {
		inputData, err := os.ReadFile(`testdata/testimage_small.png`)
		if err != nil {
			t.Fatal(`failed to read testdata file`)
		}

		buf, contentType := makeMultipartRequest(t, inputData)
		req := httptest.NewRequest(`POST`, `/upload`, buf)
		req.Header.Set(`Content-Type`, contentType)
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Fatalf(`unexpected status code %v`, w.Code)
		}
	})
}
func TestMakeHandlerNotAnImage(t *testing.T) {

	fsDirectory, err := os.MkdirTemp(``, `test-make-handler`)
	if err != nil {
		t.Fatalf(`failed to create temporary directory`)
	}

	defer os.RemoveAll(fsDirectory)

	handler := MakeHandler(fsDirectory, 1)
	t.Run(`not an image`, func(t *testing.T) {
		inputData, err := os.ReadFile(`testdata/test_file.txt`)
		if err != nil {
			t.Fatal(`failed to read testdata file`)
		}

		buf, contentType := makeMultipartRequest(t, inputData)
		req := httptest.NewRequest(`POST`, `/upload`, buf)
		req.Header.Set(`Content-Type`, contentType)
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Fatalf(`unexpected status code %v`, w.Code)
		}
	})
}

func TestMakeHandlerManyRequestGood(t *testing.T) {

	fsDirectory, err := os.MkdirTemp(``, `test-make-handler`)
	if err != nil {
		t.Fatalf(`failed to create temporary directory`)
	}

	defer os.RemoveAll(fsDirectory)

	handler := MakeHandler(fsDirectory, 4)
	t.Run(`too many request`, func(t *testing.T) {

		nbTask := 3

		var wg sync.WaitGroup

		for i := range nbTask {
			wg.Add(1)
			go func(i int) {
				defer wg.Done()

				inputData, err := os.ReadFile(`testdata/testimage_small.jpg`)
				if err != nil {
					t.Errorf(`failed to read testdata file`)
				}

				buf, contentType := makeMultipartRequest(t, inputData)
				req := httptest.NewRequest(`POST`, `/upload`, buf)
				req.Header.Set(`Content-Type`, contentType)

				w := httptest.NewRecorder()
				handler.ServeHTTP(w, req)

				if w.Code != http.StatusOK {
					t.Errorf("Request %d failed: unexpected status code  %v", i, w.Code)
				}
			}(i)
		}

		wg.Wait()
	})
}
func TestMakeHandlerManyRequest(t *testing.T) {

	fsDirectory, err := os.MkdirTemp(``, `test-make-handler`)
	if err != nil {
		t.Fatalf(`failed to create temporary directory`)
	}

	defer os.RemoveAll(fsDirectory)

	handler := MakeHandler(fsDirectory, 1)
	t.Run(`too many request`, func(t *testing.T) {

		nbTask := 100

		var wg sync.WaitGroup

		for i := range nbTask {
			wg.Add(1)
			go func(i int) {
				defer wg.Done()

				inputData, err := os.ReadFile(`testdata/testimage_small.jpg`)
				if err != nil {
					t.Errorf(`failed to read testdata file`)
				}

				buf, contentType := makeMultipartRequest(t, inputData)
				req := httptest.NewRequest(`POST`, `/upload`, buf)
				req.Header.Set(`Content-Type`, contentType)

				w := httptest.NewRecorder()
				handler.ServeHTTP(w, req)

				switch w.Code {
				case http.StatusOK:
					t.Logf("Request %d succeeded with status 200 OK", i)
				case http.StatusTooManyRequests:
					t.Logf("Request %d received status 429 Too Many Requests", i)
				default:
					t.Errorf("Request %d failed: expected status 200 or 429, got %v", i, w.Code)
				}

			}(i)
		}

		wg.Wait()

	})
}
