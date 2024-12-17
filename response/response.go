package response

import (
	"encoding/json"
	"net/http"
)

// Response define the format of Response to be send to the client
type Response struct {
	Error   string `json:"error,omitempty"`
	ImageID string `json:"image_id,omitempty"`
}

// Ok write an OK status and send the uuid or the image
func OK(writer http.ResponseWriter, ImageId string) {
	writer.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(writer).Encode(Response{
		Error:   "",
		ImageID: ImageId,
	})
}

// ErrorHanlder write an status code passed and a message
func ErrorHandler(writer http.ResponseWriter, message string, statusCode int) {
	writer.WriteHeader(statusCode)
	_ = json.NewEncoder(writer).Encode(Response{
		Error:   message,
		ImageID: "",
	})
}
