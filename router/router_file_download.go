package router

import (
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/pterodactyl/wings/config"
	"io"
	"mime"
	"net/http"
	"os"
	"path"
	"regexp"
	"strings"
)

// Download File
func fileDownloadFromUrl(c *gin.Context) {
	// Get the download url
	var data struct {
		URL string `json:"url"`
		Path string `json:"path"`
	}

	// Validate the download url
	if err := c.BindJSON(&data); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"success": false, "error": "Invalid download url."})
		return
	}

	if data.URL == "" {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"success": false, "error": "Invalid download url."})
		return
	}

	// Download plugin
	_, err := DownloadFileFromUrl(fmt.Sprintf("%s/%s%s", config.Get().System.Data, c.Param("server"), data.Path), data.URL)
	if err != nil {
		fmt.Println(err)
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"success": false, "error": err})
		return
	}

	c.Status(http.StatusNoContent)
}

// Download file from URL
func DownloadFileFromUrl(filepath string, url string) (string, error) {
	// Get the data
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}

	defer resp.Body.Close()

	filename := ""

	_, params, err := mime.ParseMediaType(resp.Header.Get("Content-Disposition"))
	if err != nil {
		regExpString := `^[a-zA-Z0-9](?:[a-zA-Z0-9 ._+%-]*[a-zA-Z0-9])?\.[a-zA-Z0-9_+%-]+$`
		splittedUrl := strings.Split(url, "/")
		lastUrlElement := splittedUrl[len(splittedUrl) - 1]

		regex, regexErr := regexp.Compile(regExpString)
		if regexErr != nil {
			return "", err
		}

		if !regex.MatchString(lastUrlElement) {
			fmt.Println(lastUrlElement)
			return "", err
		}

		filename = lastUrlElement
	} else {
		filename = params["filename"]
	}

	// Modify download path
	filepath = path.Join(filepath, filename)

	if filename == "" {
		return "", errors.New("invalid name")
	}

	// Create the file
	out, err := os.Create(filepath)
	if err != nil {
		return "", err
	}

	defer out.Close()

	// Write the body to file
	_, err = io.Copy(out, resp.Body)
	return filename, err
}
