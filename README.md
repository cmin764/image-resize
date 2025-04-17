# Image Resizing Service

A Go-based service that resizes images from provided URLs and serves them through a REST API. The service supports both synchronous and asynchronous processing modes.

## Features

- Resize images from any public URL
- Support for both synchronous and asynchronous processing
- LRU caching for processed images
- Configurable timeouts and processing delays
- Browser-like User-Agent for better compatibility

## Prerequisites

- Go 1.16 or later
- Air (for development with hot reload)
- HTTPie (for testing, or use curl)

## Installation

1. Clone the repository:
```bash
git clone https://github.com/cmin764/interview-fm-backend.git
cd interview-fm-backend
```

2. Install Air for development:
```bash
go install github.com/cosmtrek/air@latest
```

3. Install project dependencies:
```bash
go mod download
```

## Configuration

Create a `.env` file in the project root by copying the template:
```bash
cp .env.template .env
```

Then edit the `.env` file with the following settings:

```env
# Image processing timeout in seconds (how long to wait for an image to be processed)
IMAGE_PROCESSING_TIMEOUT=1

# Simulated processing duration in seconds (for testing)
IMAGE_PROCESSING_DURATION=3
```

## Running the Server

### Development Mode (with hot reload)
```bash
# Try this first
air

# If the above doesn't work, use the full path
~/go/bin/air
```

### Production Mode
```bash
go run .
```

The server will start on `http://localhost:8080`

## API Usage

### Resize Images

Send a POST request to `/v1/resize` with a JSON body containing the image URLs and desired dimensions:

```bash
# Synchronous mode (default)
http POST ":8080/v1/resize" @req.json

# Asynchronous mode
http POST ":8080/v1/resize?async=true" @req.json
```

Example request body (`req.json`):
```json
{
  "urls": [
    "https://example.com/image1.jpg",
    "https://example.com/image2.jpg"
  ],
  "width": 200,
  "height": 0  // 0 maintains aspect ratio
}
```

### Access Resized Images

After processing, access the resized images at:
```
http://localhost:8080/v1/image/{image-id}.jpeg
```

## Response Handling

### Synchronous Mode
- Returns 201 Created when new images are processed
- Returns 200 OK when all images are served from cache
- Returns 202 Accepted in async mode

### Asynchronous Mode
- Returns 202 Accepted immediately
- Check individual image URLs for status:
  - 102 Processing: Image is still being processed
  - 200 OK: Image is ready
  - 404 Not Found: Image processing failed

## Example Workflow

1. Send resize request:
```bash
http POST ":8080/v1/resize?async=true" @req.json
```

2. Check image status:
```bash
http http://localhost:8080/v1/image/{image-id}.jpeg
```

3. If status is 102, wait and retry until you get 200 OK

## ToDo and Improvements

See [docs/improvements.md](docs/improvements.md) for a list of potential improvements, including:
- Adding comprehensive test coverage
- Performance optimizations
- Security enhancements
- Additional features
- Infrastructure improvements
