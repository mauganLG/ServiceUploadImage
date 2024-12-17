# ServiceUploadImage


This a service that accept an `JPEG` image on the path by POST requests on URL `/upload`

- The size of the image is limited to `8192 Kilobytes`
- If the limit is exceeded the service  respond with `413 Request entity too large`.
- The images are expected to be in `JPEG` format in the multipart body-part named `image`.
- If it’s not a JPEG image then the client  receive a *`400 Bad Request`* error.
- After the image passed step the server generates the ID for the image in `UUIDv4` format.
- While the server process the image, the client receive a `200 Ok` response with the ID for the image as soon as the upload 
   and the image-ID is to returned in the response body in JSON format in the following form: `{"image_id":"<uuid-v4>"}`.
- It perform image downscaling
- And then it save the resulting image on the local filesystem. The path to the image file is in the format  `<base path>/<uuid>.jpg` (`base path` need to be supplied in a environment variable named `BASE_PATH`).
- Because the image processing can be quite memory- and CPU-intensive, no more than N images is processed in parallel
   (with N, the number configurable at start-up time but default to number of processors).
- If there are already N images being processed when a new request comes in and no slot becomes available within
   *100 ms* – the client receive a `429 Too Many Requests` HTTP-error and his upload is not be accepted for processing.
 
