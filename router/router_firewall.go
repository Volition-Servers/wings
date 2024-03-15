package router

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/pterodactyl/wings/router/middleware"
	"os/exec"

	"net/http"
)

type Rule struct {
	Ip       string `json:"ip"`
	Port     int    `json:"port"`
	Priority int    `json:"priority"`
	Type     string `json:"type"`
}

// Add Rule
func addFirewallRule(c *gin.Context) {
	// Get the requested server
	s := middleware.ExtractServer(c)

	// Get data
	var data struct {
		RULES []Rule `json:"rules"`
	}

	// Validate data
	if err := c.BindJSON(&data); err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"success": false, "error": "Invalid data provided."})
		return
	}

	// Get Local IP
	localIp, err := s.Environment.GetLocalIP(s.Context())
	if err != nil {
		if err.Error() == "serverNotRunningError" {
			c.JSON(http.StatusOK, gin.H{"success": true})
			return
		}

		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"success": false, "error": "Failed to fetch the local ip."})
		return
	}

	// Remove old rules
	err = exec.Command("/bin/sh", "-c", fmt.Sprintf("iptables-save | sed -r '/DOCKER.*comment.*%s/s/-A/iptables -D/e'", s.Config().Uuid)).Run()
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"success": false, "error": "Failed to remove old rules."})
		return
	}

	addError := false

	for _, item := range data.RULES {
		ruleType := ""

		if item.Type == "allow" {
			ruleType = "ACCEPT"
		} else {
			ruleType = "DROP"
		}

		// Add TCP Rule
		err = exec.Command("/bin/sh", "-c", fmt.Sprintf("iptables -I DOCKER 1 -d %s/32 ! -i pterodactyl0 -o pterodactyl0 -p tcp -m tcp --dport %d -s %s -m comment --comment '%s' -j %s", localIp, item.Port, item.Ip, s.Config().Uuid, ruleType)).Run()
		if err != nil {
			addError = true
		}

		// Add UDP Rule
		err = exec.Command("/bin/sh", "-c", fmt.Sprintf("iptables -I DOCKER 1 -d %s/32 ! -i pterodactyl0 -o pterodactyl0 -p udp -m udp --dport %d -s %s -m comment --comment '%s' -j %s", localIp, item.Port, item.Ip, s.Config().Uuid, ruleType)).Run()
		if err != nil {
			addError = true
		}
	}

	if addError {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"success": false, "error": "Failed to assign new rules."})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true})
}

func removeFirewallRule(c *gin.Context) {
	// Get the requested server
	s := middleware.ExtractServer(c)

	// Get data
	var data struct {
		RULE Rule `json:"rule"`
	}

	// Validate data
	if err := c.BindJSON(&data); err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"success": false, "error": "Invalid data provided."})
		return
	}

	ruleType := ""

	if data.RULE.Type == "allow" {
		ruleType = "ACCEPT"
	} else {
		ruleType = "DROP"
	}

	// Get Local IP
	localIp, err := s.Environment.GetLocalIP(s.Context())
	if err != nil {
		if err.Error() == "serverNotRunningError" {
			c.JSON(http.StatusOK, gin.H{"success": true})
			return
		}

		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"success": false, "error": "Failed to fetch the local ip."})
		return
	}

	removeError := false

	// Remove TCP Rule
	err = exec.Command("/bin/sh", "-c", fmt.Sprintf("iptables -D DOCKER -d %s/32 ! -i pterodactyl0 -o pterodactyl0 -p tcp -m tcp --dport %d -s %s -m comment --comment '%s' -j %s", localIp, data.RULE.Port, data.RULE.Ip, s.Config().Uuid, ruleType)).Run()
	if err != nil {
		removeError = true
	}

	// Remove UDP Rule
	err = exec.Command("/bin/sh", "-c", fmt.Sprintf("iptables -D DOCKER -d %s/32 ! -i pterodactyl0 -o pterodactyl0 -p udp -m udp --dport %d -s %s -m comment --comment '%s' -j %s", localIp, data.RULE.Port, data.RULE.Ip, s.Config().Uuid, ruleType)).Run()
	if err != nil {
		removeError = true
	}

	if removeError {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"success": false, "error": "Failed to remove the rule."})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true})
}
