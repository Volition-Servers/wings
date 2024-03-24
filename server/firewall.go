package server

import (
	"fmt"
	"os/exec"
	"time"
)

// Add Rules
func (s *Server) AddRules() {
	time.Sleep(2 * time.Second)

	// Get Local IP
	localIp, err := s.Environment.GetLocalIP(s.Context())
	if err != nil {
		if err.Error() != "serverNotRunningError" {
			s.PublishConsoleOutputFromDaemon("Failed to add firewall rules to the server. Please restart the server...")
			return
		}
	}

	// Get Rules
	rules, err := s.client.GetFirewallRules(s.ctx, s.Config().Uuid)
	if err != nil {
		s.PublishConsoleOutputFromDaemon("Failed to add firewall rules to the server. Please restart the server...")
		return
	}

	addError := false

	for _, item := range rules {
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
		s.PublishConsoleOutputFromDaemon("Failed to add firewall rules to the server. Please restart the server...")
	} else {
		s.PublishConsoleOutputFromDaemon("Firewall rules successfully added to the server.")
	}
}
