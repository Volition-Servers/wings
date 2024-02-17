package router

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/pterodactyl/wings/router/middleware"
	auth "gopkg.in/src-d/go-git.v4/plumbing/transport/http"
	"net/http"
	"os/exec"
)

func gitClone(c *gin.Context) {
	// Get the requested server
	s := middleware.ExtractServer(c)

	// Get the download url
	var data struct {
		PATH   string `json:"path"`
		URL    string `json:"url"`
		BRANCH string `json:"branch"`
		TOKEN  string `json:"token"`
	}

	// Validate parameters
	if err := c.BindJSON(&data); err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"success": false, "error": "Invalid parameters"})
		return
	}

	// Make safe path
	p, err := s.Filesystem().SafePath(data.PATH)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"success": false, "error": "Invalid server path"})
		return
	}

	// Clone the directory
	options := &git.CloneOptions{
		URL: data.URL,
	}

	// Add auth parameter if needed
	if len(data.TOKEN) != 0 {
		options.Auth = &auth.BasicAuth{
			Username: "pterodactyl",
			Password: data.TOKEN,
		}
	}

	if data.BRANCH != "" {
		options.ReferenceName = plumbing.NewBranchReferenceName(data.BRANCH)
		options.SingleBranch = true
	}

	_, cloneErr := git.PlainClone(p, false, options)

	// Check errors
	if cloneErr != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"success": false, "error": fmt.Sprintf("Failed to start the clone process: %s", cloneErr.Error())})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true})
}

func gitPull(c *gin.Context) {
	// Get the requested server
	s := middleware.ExtractServer(c)

	// Get the download url
	var data struct {
		PATH  string `json:"path"`
		TOKEN string `json:"token"`
		RESET bool   `json:"reset"`
	}

	// Validate parameters
	if err := c.BindJSON(&data); err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"success": false, "error": "Invalid parameters"})
		return
	}

	// Make safe path
	p, err := s.Filesystem().SafePath(data.PATH)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"success": false, "error": "Invalid server path"})
		return
	}

	// Open current git instance
	r, err := git.PlainOpen(p)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"success": false, "error": "Initialized git repository not found"})
		return
	}

	// Get the working directory for the repository
	w, err := r.Worktree()
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"success": false, "error": "Failed to open the worktree"})
		return
	}

	// Reset every local changes if needed
	if data.RESET {
		_ = w.Reset(&git.ResetOptions{
			Mode: git.HardReset,
		})
	}

	// Run stash if not reset
	if !data.RESET {
		_ = exec.Command("/bin/sh", "-c", fmt.Sprintf("cd %s && git add ./*", p)).Run()
		_ = exec.Command("/bin/sh", "-c", fmt.Sprintf("cd %s && git stash", p)).Run()
	}

	// Make options
	options := &git.PullOptions{
		RemoteName: "origin",
	}

	// Add auth parameter if needed
	if len(data.TOKEN) != 0 {
		options.Auth = &auth.BasicAuth{
			Username: "pterodactyl",
			Password: data.TOKEN,
		}
	}

	// Pull the repository
	err = w.Pull(options)

	// Move back the stashed changes
	if !data.RESET {
		_ = exec.Command("/bin/sh", "-c", fmt.Sprintf("cd %s && git stash pop", p)).Run()
	}

	// Check errors
	if err == git.NoErrAlreadyUpToDate {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"success": false, "error": "Already up-to-date"})
		return
	}

	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"success": false, "error": fmt.Sprintf("Failed to pull the repository: %s", err.Error())})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true})
}
