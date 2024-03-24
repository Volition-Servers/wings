package router

import (
	"archive/zip"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/pterodactyl/wings/config"
	"github.com/pterodactyl/wings/router/middleware"
	"io"
	"mime"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"
)

// Install mod
func installMod(c *gin.Context) {
	// Get the requested server
	s := middleware.ExtractServer(c)

	// Get the download url
	var data struct {
		DOWNLOAD_URL    string `json:"download_url"`
		FOLDER          string `json:"folder"`
		DECOMPRESS_TYPE string `string:"decompress_type"`
	}

	// Validate parameters
	if err := c.BindJSON(&data); err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"success": false, "error": "Invalid parameters."})
		return
	}

	if data.DOWNLOAD_URL == "" || data.DECOMPRESS_TYPE == "" {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"success": false, "error": "Invalid download url or decompress type."})
		return
	}

	if data.DECOMPRESS_TYPE == "unzip" || data.DECOMPRESS_TYPE == "none" {
	} else {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"success": false, "error": "Invalid decompress type."})
		return
	}

	// Create base folder
	if err := s.Filesystem().CreateDirectory(data.FOLDER, "/"); err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"success": false, "error": "Failed to create the plugins directory."})
		return
	}

	// Download mod
	filename := path.Join(config.Get().System.Data, "/", c.Param("server"), "/", data.FOLDER, "/")

	if data.DECOMPRESS_TYPE == "none" {
		filename = path.Join(config.Get().System.Data, "/", c.Param("server"), "/", data.FOLDER, "/")
	}

	downloadedFileName, err := DownloadMod(filename, data.DOWNLOAD_URL, data.DECOMPRESS_TYPE)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"success": false, "error": "Failed to download the mod."})
		return
	}

	// Unzip the mod
	if data.DECOMPRESS_TYPE == "unzip" {
		_, downloadErr := UnzipMod(path.Join(filename, downloadedFileName), filename)
		if downloadErr != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"success": false, "error": "Failed to unzip the mod."})
			return
		}

		deleteErr := s.Filesystem().Delete(path.Join("/", data.FOLDER, "mod.zip"))
		if deleteErr != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"success": false, "error": "Failed to delete the mod.zip."})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{"success": true})
}

// Uninstall Mod
func uninstallMod(c *gin.Context) {
	// Get the requested server
	s := middleware.ExtractServer(c)

	// Get the download url
	var data struct {
		BASE        string   `json:"base"`
		REMOVES     []string `json:"removes"`
		REMOVE_BASE bool     `string:"remove_base"`
	}

	// Validate parameters
	if err := c.BindJSON(&data); err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"success": false, "error": "Invalid parameters."})
		return
	}

	// Delete files
	if data.REMOVE_BASE {
		files, err := s.Filesystem().ListDirectory(path.Join("/", data.BASE, "/"))
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"success": false, "error": "Failed to load files."})
			return
		}

		for _, file := range files {
			err := s.Filesystem().Delete(path.Join("/", data.BASE, "/", file.Name()))
			if err != nil {
				c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"success": false, "error": "Failed to delete all files."})
				return
			}
		}
	} else {
		successDelete := false

		for _, file := range data.REMOVES {
			err := s.Filesystem().Delete(path.Join("/", data.BASE, "/", file))
			if err != nil {
				successDelete = true
			}
		}

		if successDelete {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"success": false, "error": "Failed to delete mod files."})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{"success": true})
}

// Download file from URL
func DownloadMod(filepath string, url string, decompress_type string) (string, error) {
	// Get the data
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}

	defer resp.Body.Close()

	filename := "mod.zip"

	// Get the file name
	if decompress_type == "none" {
		_, params, err := mime.ParseMediaType(resp.Header.Get("Content-Disposition"))
		if err != nil {
			regExpString := `^[a-zA-Z0-9](?:[a-zA-Z0-9 ._-]*[a-zA-Z0-9])?\.[a-zA-Z0-9_-]+$`
			splittedUrl := strings.Split(url, "/")
			lastUrlElement := splittedUrl[len(splittedUrl) - 1]

			regex, regexErr := regexp.Compile(regExpString)
			if regexErr != nil {
				return "", err
			}

			if !regex.MatchString(lastUrlElement) {
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
	} else {
		filepath = path.Join(filepath, "mod.zip")
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

// Unzip the downloaded file
func UnzipMod(src string, dest string) ([]string, error) {
	var filenames []string

	r, err := zip.OpenReader(src)
	if err != nil {
		return filenames, err
	}

	defer r.Close()

	for _, f := range r.File {
		// Store filename/path for returning and using later on
		fpath := filepath.Join(dest, f.Name)

		// Check for ZipSlip. More Info: http://bit.ly/2MsjAWE
		if !strings.HasPrefix(fpath, filepath.Clean(dest)+string(os.PathSeparator)) {
			return filenames, fmt.Errorf("%s: illegal file path", fpath)
		}

		filenames = append(filenames, fpath)

		if f.FileInfo().IsDir() {
			// Make Folder
			os.MkdirAll(fpath, os.ModePerm)
			continue
		}

		// Make File
		if err = os.MkdirAll(filepath.Dir(fpath), os.ModePerm); err != nil {
			return filenames, err
		}

		outFile, err := os.OpenFile(fpath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		if err != nil {
			return filenames, err
		}

		rc, err := f.Open()
		if err != nil {
			return filenames, err
		}

		_, err = io.Copy(outFile, rc)

		// Close the file without defer to close before next iteration of loop
		outFile.Close()
		rc.Close()

		if err != nil {
			return filenames, err
		}
	}

	return filenames, nil
}
