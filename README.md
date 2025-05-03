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

```console
git clone https://github.com/cmin764/image-resize.git
cd image-resize
```

2. Install Air for development:

```console
go install github.com/air-verse/air@latest
```

3. Install project dependencies:
  
```console
go mod download
```

## Configuration

Create a `.env` file in the project root by copying the template:

```console
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

```console
# Try this first
air

# If the above doesn't work, use the full path
~/go/bin/air
```

### Production Mode

```console
go run .
```

The server will start on `http://localhost:8080`

## API Usage

### Resize Images

Send a POST request to `/v1/resize` with a JSON body containing the image URLs and desired dimensions:

```console
# Synchronous mode (default)
http POST ":8080/v1/resize" @data/req.json

# Asynchronous mode
http POST ":8080/v1/resize?async=true" @data/req.json
```

Example request [body](./data/req.json):

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

```console
http://localhost:8080/v1/image/{image-id}.jpeg
```

## Response Handling

### Synchronous Mode

- Returns:
  - `200 OK` when all images are served straight from the cache
  - `201 Created` when new images are processed and saved to cache

### Asynchronous Mode

- Returns `202 Accepted` immediately (later processing)

### Resized image retrieval

Check individual image URLs for status.

- `102 Processing`: Image is still being processed
- `200 OK`: Image is ready and retrieved
- `404 Not Found`: Image processing failed and not found in the cache

## Example Workflow

1. Send resize request:

```console
http POST ":8080/v1/resize?async=true" @data/req.json
```

2. Check image status:

```console
http http://localhost:8080/v1/image/AQPdOTElKSNH_piyJSCR4p1yMMRC1omE-mClICKUoks=.jpeg
```

> If status is `102`, wait and retry until you get `200` (or a `404` if the process failed in the meantime)

## Improvements

See [improvements](docs/improvements.md) for a list of potential fixes, including:

- Adding comprehensive test coverage
- Performance optimizations
- Security enhancements
- Additional features
- Infrastructure improvements
