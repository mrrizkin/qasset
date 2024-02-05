# QAsset

QAsset is a simple and quick file uploader designed to streamline the process of uploading files to a server.

## Table of Contents
- [Introduction](#introduction)
- [Commands](#commands)
- [Installation](#installation)
- [Usage](#usage)

## Introduction

QAsset simplifies the file uploading process to a server, providing a quick and efficient solution for users.

## Commands

The Makefile provides the following commands:

- `make info`: Display information about available commands.
- `make install`: Install dependencies and clean build artifacts.
- `make all`: Install dependencies, build, and move artifacts.
- `make serve`: Run the QAsset server.
- `make build`: Build the QAsset project.
- `make clean`: Clean build artifacts.
- `make move`: Move build artifacts and install as a service.

## Installation

To quickly install QAsset and its dependencies, run:

```bash
make install
```

This will tidy the Go modules and clean any existing build artifacts.

## Usage

After installation, use the following commands:

```bash
make build  # Build QAsset
make serve  # Run QAsset
make clean  # Clean build artifacts
```

For a complete build, including installation and running as a service:

```bash
make all
```

After installation, QAsset provides various endpoints to handle file uploads, deletions, and more. Here are the main endpoints:

- `GET /`: Displays a welcome message.
- `POST /upload_image`: Handles the upload of images.
- `POST /upload_file`: Handles the upload of files.
- `DELETE /delete_image`: Deletes uploaded images.
- `DELETE /delete_file`: Deletes uploaded files.
- `GET /robots.txt`: Returns the content for robots.txt.
- `GET /metrics`: Provides metrics for monitoring (requires a monitoring tool).

Additionally, the wildcard endpoint `GET /*` is handled by a generic handler with caching.

To interact with these endpoints, you can use tools like cURL or use your preferred programming language's HTTP library. For example:

```bash
# Upload an image
curl -X POST http://your-qasset-server/upload_image -F "file=@/path/to/your/image.jpg"

# Upload a file
curl -X POST http://your-qasset-server/upload_file -F "file=@/path/to/your/file.txt"

# Delete an image
curl -X DELETE http://your-qasset-server/delete_image?filename=image.jpg

# Delete a file
curl -X DELETE http://your-qasset-server/delete_file?filename=file.txt
```
