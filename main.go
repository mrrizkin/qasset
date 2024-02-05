package main

import (
	"flag"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/log"
	"github.com/gofiber/fiber/v2/middleware/cache"
	"github.com/gofiber/fiber/v2/middleware/healthcheck"
	"github.com/gofiber/fiber/v2/middleware/helmet"
	"github.com/gofiber/fiber/v2/middleware/idempotency"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/monitor"
	"github.com/gofiber/fiber/v2/middleware/recover"
)

var (
	err error
	cwd string

	prefork   = flag.Bool("prefork", false, "Enable prefork")
	port      = flag.String("port", "3000", "Port to listen on")
	imageDir  = "storage/images"
	fileDir   = "storage/files"
	pathRegex = regexp.MustCompile(`^[a-zA-Z0-9_\-/\\]+$`)
)

func init() {
	// make sure the storage directory exists
	os.MkdirAll("storage/images", 0755)
	os.MkdirAll("storage/files", 0755)
}

func main() {
	flag.Parse()

	app := fiber.New(fiber.Config{
		Prefork: *prefork,
	})

	app.Use(logger.New())
	app.Use(helmet.New())
	app.Use(idempotency.New())
	app.Use(healthcheck.New())
	app.Use(recover.New())

	app.Get("/", welcomeHandler)
	app.Post("/upload_image", uploadImageHandler)
	app.Post("/upload_file", uploadFileHandler)
	app.Delete("/delete_image", deleteImageHandler)
	app.Delete("/delete_file", deleteFileHandler)
	app.Get("/robots.txt", robotsTxtHandler)
	app.Get("/metrics", monitor.New(monitor.Config{Title: "QAsset Metrics Page"}))

	app.Get("/*", cache.New(), genericHandler)

	log.Fatal(app.Listen(fmt.Sprintf(":%s", *port)))
}

func welcomeHandler(c *fiber.Ctx) error {
	c.Set("Content-Type", "text/html")

	return c.SendString(
		"<h1>Welcome to QAsset</h1>" +
			"<p>QAsset is a simple and easy to use asset management system.</p>" +
			"<p>It allows you to upload and manage images and files.</p>" +
			"<p>For more information, please visit <a href='https://github.com/nugrhrizki/qasset'><strong>QAsset</strong></a>.</p>",
	)
}

type UploadImageRequest struct {
	Path string `form:"path"`
}

func uploadImageHandler(c *fiber.Ctx) error {
	var req UploadImageRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "invalid request",
		})
	}

	image, err := c.FormFile("image")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "image is required",
		})
	}

	allowedExtensions := []string{".png", ".jpg", ".gif", ".jpeg"}
	ext := filepath.Ext(image.Filename)
	if !contains(allowedExtensions, ext) {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "only images are allowed",
		})
	}

	filename := qassetSignature(image.Filename)
	customPath := req.Path
	if customPath == "" {
		customPath = "images"
	}

	if customPath != "images" {
		if !pathRegex.MatchString(customPath) {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"success": false,
				"message": "invalid path",
			})
		}

		if err := os.MkdirAll(filepath.Join(imageDir, customPath), 0755); err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"success": false,
				"message": "unable to upload image",
				"debug":   err.Error(),
			})
		}
	}

	if err := c.SaveFile(image, filepath.Join(imageDir, customPath, filename)); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "unable to upload image",
			"debug":   err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success":  true,
		"message":  "image has been uploaded successfully",
		"filename": filename,
		"url":      fmt.Sprintf("http://%s/%s/%s", c.Hostname(), customPath, filename),
	})
}

type UploadFileRequest struct {
	Path string `form:"path"`
}

func uploadFileHandler(c *fiber.Ctx) error {
	var req UploadFileRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "invalid request",
		})
	}

	file, err := c.FormFile("file")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "file is required",
		})
	}

	filename := qassetSignature(file.Filename)
	customPath := req.Path
	if customPath == "" {
		customPath = "files"
	}

	if customPath != "files" {
		if !pathRegex.MatchString(customPath) {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"success": false,
				"message": "invalid path",
			})
		}

		if err := os.MkdirAll(filepath.Join(fileDir, customPath), 0755); err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"success": false,
				"message": "unable to upload file",
				"debug":   err.Error(),
			})
		}
	}

	if err := c.SaveFile(file, filepath.Join(fileDir, customPath, filename)); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "unable to upload file",
			"debug":   err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success":  true,
		"message":  "file has been uploaded successfully",
		"filename": filename,
		"url":      fmt.Sprintf("http://%s/%s/%s", c.Hostname(), customPath, filename),
	})
}

type DeleteImageRequest struct {
	Path     string `form:"path"`
	Filename string `form:"filename"`
}

func deleteImageHandler(c *fiber.Ctx) error {
	var req DeleteImageRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "invalid request",
		})
	}

	if req.Path == "" || req.Filename == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "path and filename are required",
		})
	}

	if err := os.Remove(filepath.Join(imageDir, req.Path, req.Filename)); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "unable to delete image",
			"debug":   err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "image has been deleted successfully",
	})
}

type DeleteFileRequest struct {
	Path     string `form:"path"`
	Filename string `form:"filename"`
}

func deleteFileHandler(c *fiber.Ctx) error {
	var req DeleteFileRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "invalid request",
		})
	}

	if req.Path == "" || req.Filename == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "path and filename are required",
		})
	}

	if err := os.Remove(filepath.Join(fileDir, req.Path, req.Filename)); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "unable to delete file",
			"debug":   err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "file has been deleted successfully",
	})
}

func robotsTxtHandler(c *fiber.Ctx) error {
	return c.SendString("User-agent: *\nDisallow: /")
}

func genericHandler(c *fiber.Ctx) error {
	path := c.Path()

	if _, err := os.Stat(filepath.Join(imageDir, path)); err == nil {
		return c.SendFile(filepath.Join(imageDir, path))
	}

	if _, err := os.Stat(filepath.Join(fileDir, path)); err == nil {
		return c.SendFile(filepath.Join(fileDir, path))
	}

	return c.SendFile("resources/asset/broken.webp")
}

func contains(arr []string, str string) bool {
	for _, a := range arr {
		if a == str {
			return true
		}
	}
	return false
}

func qassetSignature(str string) string {
	current := strconv.Itoa(int(time.Now().Unix()))
	randStr := randomString(7)
	return "cdn-" + current + randStr + "-" + str
}

func randomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[rand.Intn(len(charset))]
	}
	return string(b)
}
