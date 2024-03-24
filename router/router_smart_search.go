package router

import (
	"bufio"
	"github.com/gin-gonic/gin"
	"github.com/pterodactyl/wings/config"
	"github.com/pterodactyl/wings/router/middleware"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
	"sync"
)

type ResponseItem struct {
	File  string `json:"file"`
	Lines []int  `json:"lines"`
}

func smartContains(s []string, e string) bool {
	for _, a := range s {
		if a == strings.ToLower(e) {
			return true
		}
	}
	return false
}

// Handle the smart file search
func smartSearch(c *gin.Context) {
	// Get the requested server
	s := middleware.ExtractServer(c)

	// Get the data
	var data struct {
		PATH  string `json:"path"`
		QUERY string `json:"query"`
	}

	// Validate parameters
	if err := c.BindJSON(&data); err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"success": false, "error": "Invalid parameters"})
		return
	}

	// Make safe path
	_, p, closeFd, err := s.Filesystem().UnixFS().SafePath(data.PATH)
	defer closeFd()

	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"success": false, "error": "Invalid server path"})
		return
	}

	var results []ResponseItem
	var wg sync.WaitGroup
	searchPath := path.Join(config.Get().System.Data, s.Config().Uuid, p)
	var ignoredExtensions = []string{".jar", ".dll", ".zip", ".tar.gz", ".7z", ".exe", ".so", ".rar", ".mp3", ".mp4", ".mov", ".png", ".avi", ".jpg", ".jpeg", ".gif", ".pdf", ".svg", ".webp", ".webm", ".luac"}

	filepath.Walk(searchPath, func(folder string, file os.FileInfo, err error) error {
		if !file.IsDir() {
			// Ignore the extensions
			if smartContains(ignoredExtensions, filepath.Ext(path.Join(folder, file.Name()))) {
				return nil
			}

			// Ignore files above 10 MB
			if file.Size() > 10485760 {
				return nil
			}

			wg.Add(1)

			// Make the result chain to fetch the response
			result := make(chan ResponseItem)
			go readFile(&wg, path.Join(config.Get().System.Data, s.Config().Uuid), folder, data.QUERY, result)
			value := <-result

			// Add the file when found match
			if len(value.Lines) > 0 {
				results = append(results, value)
			}
		}
		return nil
	})

	wg.Wait()

	c.JSON(http.StatusOK, gin.H{"success": true, "result": results})
}

// Read the file and search for query
func readFile(wg *sync.WaitGroup, serverPath string, folder string, query string, result chan ResponseItem) {
	defer wg.Done()

	file, err := os.Open(folder)
	defer file.Close()

	if err != nil {
		return
	}

	var lines []int

	scanner := bufio.NewScanner(file)
	for i := 1; scanner.Scan(); i++ {
		if strings.Contains(strings.ToLower(scanner.Text()), strings.ToLower(query)) {
			lines = append(lines, i)
		}
	}

	result <- ResponseItem{File: strings.ReplaceAll(folder, serverPath, ""), Lines: lines}
}
