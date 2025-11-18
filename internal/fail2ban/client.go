package fail2ban

import (
	"bufio"
	"fmt"
	"os/exec"
	"strings"
	"time"
)

type Client struct {
	clientPath string
	useSudo    bool
}

func NewClient(clientPath string, useSudo bool) *Client {
	return &Client{
		clientPath: clientPath,
		useSudo:    useSudo,
	}
}

func (c *Client) executeCommand(args ...string) (string, error) {
	var cmd *exec.Cmd
	
	if c.useSudo {
		// Use sudo to run fail2ban-client
		cmd = exec.Command("sudo", append([]string{c.clientPath}, args...)...)
	} else {
		cmd = exec.Command(c.clientPath, args...)
	}
	
	output, err := cmd.CombinedOutput()
	if err != nil {
		outputStr := strings.TrimSpace(string(output))
		
		if strings.Contains(outputStr, "Permission denied") || strings.Contains(outputStr, "you must be root") {
			return "", fmt.Errorf("permission denied: fail2ban requires root privileges. Either run the server as root, or set 'use_sudo: true' in config and configure passwordless sudo for fail2ban-client. Error: %s", outputStr)
		}
		
		return "", fmt.Errorf("fail2ban-client error: %w, output: %s", err, outputStr)
	}
	return strings.TrimSpace(string(output)), nil
}

// GetStatus returns the overall status of fail2ban
func (c *Client) GetStatus() (map[string]interface{}, error) {
	output, err := c.executeCommand("status")
	if err != nil {
		return nil, err
	}

	status := make(map[string]interface{})
	scanner := bufio.NewScanner(strings.NewReader(output))
	
	var currentSection string
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		if strings.HasPrefix(line, "Status for the jail:") {
			currentSection = "jails"
			if status["jails"] == nil {
				status["jails"] = []string{}
			}
			continue
		}

		if strings.Contains(line, ":") {
			parts := strings.SplitN(line, ":", 2)
			if len(parts) == 2 {
				key := strings.TrimSpace(parts[0])
				value := strings.TrimSpace(parts[1])
				status[key] = value
			}
		} else if currentSection == "jails" {
			jails := status["jails"].([]string)
			status["jails"] = append(jails, line)
		}
	}

	return status, nil
}

// GetJailStatus returns detailed status for a specific jail
func (c *Client) GetJailStatus(jailName string) (map[string]interface{}, error) {
	output, err := c.executeCommand("status", jailName)
	if err != nil {
		return nil, err
	}

	status := make(map[string]interface{})
	scanner := bufio.NewScanner(strings.NewReader(output))
	
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		if strings.Contains(line, ":") {
			parts := strings.SplitN(line, ":", 2)
			if len(parts) == 2 {
				key := strings.TrimSpace(parts[0])
				value := strings.TrimSpace(parts[1])
				
				// Try to parse numbers
				if strings.Contains(value, " ") {
					status[key] = value
				} else {
					status[key] = value
				}
			}
		}
	}

	return status, nil
}

// GetJails returns a list of all configured jails
func (c *Client) GetJails() ([]string, error) {
	output, err := c.executeCommand("status")
	if err != nil {
		return nil, err
	}

	var jails []string
	scanner := bufio.NewScanner(strings.NewReader(output))
	inJailSection := false

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		
		if strings.Contains(line, "Jail list:") || strings.Contains(line, "Status for the jail:") {
			inJailSection = true
			continue
		}

		if inJailSection && line != "" {
			// Jails are listed one per line or comma-separated
			if strings.Contains(line, ",") {
				parts := strings.Split(line, ",")
				for _, part := range parts {
					jail := strings.TrimSpace(part)
					if jail != "" {
						jails = append(jails, jail)
					}
				}
			} else {
				jails = append(jails, line)
			}
		}
	}

	return jails, nil
}

// GetBannedIPs returns a list of banned IPs for a jail
func (c *Client) GetBannedIPs(jailName string) ([]string, error) {
	output, err := c.executeCommand("get", jailName, "banned")
	if err != nil {
		return nil, err
	}

	var ips []string
	scanner := bufio.NewScanner(strings.NewReader(output))
	
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line != "" {
			ips = append(ips, line)
		}
	}

	return ips, nil
}

// BanIP bans an IP address in a specific jail
func (c *Client) BanIP(jailName, ip string) error {
	_, err := c.executeCommand("set", jailName, "banip", ip)
	return err
}

// UnbanIP unbans an IP address in a specific jail
func (c *Client) UnbanIP(jailName, ip string) error {
	_, err := c.executeCommand("set", jailName, "unbanip", ip)
	return err
}

// GetJailStats returns statistics for a jail
func (c *Client) GetJailStats(jailName string) (map[string]interface{}, error) {
	status, err := c.GetJailStatus(jailName)
	if err != nil {
		return nil, err
	}

	stats := make(map[string]interface{})
	
	// Extract relevant statistics from status
	if filterList, ok := status["Filter"].(string); ok {
		stats["filter"] = filterList
	}
	if currentlyBanned, ok := status["Currently banned"].(string); ok {
		stats["currently_banned"] = currentlyBanned
	}
	if totalBanned, ok := status["Total banned"].(string); ok {
		stats["total_banned"] = totalBanned
	}
	if bannedIPList, ok := status["Banned IP list"].(string); ok {
		stats["banned_ips"] = strings.Fields(bannedIPList)
	}

	return stats, nil
}

// GetOverallStats returns overall statistics
func (c *Client) GetOverallStats() (map[string]interface{}, error) {
	jails, err := c.GetJails()
	if err != nil {
		return nil, err
	}

	stats := make(map[string]interface{})
	stats["jail_count"] = len(jails)
	stats["jails"] = jails
	
	totalBanned := 0
	jailStats := make(map[string]interface{})
	
	for _, jail := range jails {
		jailStat, err := c.GetJailStats(jail)
		if err == nil {
			jailStats[jail] = jailStat
			if banned, ok := jailStat["currently_banned"].(string); ok {
				var count int
				fmt.Sscanf(banned, "%d", &count)
				totalBanned += count
			}
		}
	}
	
	stats["total_banned_ips"] = totalBanned
	stats["jail_details"] = jailStats
	stats["timestamp"] = time.Now().Unix()

	return stats, nil
}

// StartJail starts a jail
func (c *Client) StartJail(jailName string) error {
	_, err := c.executeCommand("start", jailName)
	return err
}

// StopJail stops a jail
func (c *Client) StopJail(jailName string) error {
	_, err := c.executeCommand("stop", jailName)
	return err
}

// RestartJail restarts a jail
func (c *Client) RestartJail(jailName string) error {
	_, err := c.executeCommand("restart", jailName)
	return err
}

// ReloadJail reloads a jail configuration
func (c *Client) ReloadJail(jailName string) error {
	_, err := c.executeCommand("reload", jailName)
	return err
}

