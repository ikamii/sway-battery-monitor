package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

// Holds application configuration
type Config struct {
	Threshold      int
	CheckInterval  int
	NotifyCommand  string
	BatteryPath    string
	LastNotifiedAt time.Time
	CooldownPeriod time.Duration
}

func main() {
	// Parse command line arguments
	threshold := flag.Int("threshold", 20, "Battery percentage threshold")
	checkInterval := flag.Int("interval", 60, "Check interval in seconds")
	batteryPath := flag.String("battery", "/sys/class/power_supply/BAT0", "Path to battery informantion")
	cooldown := flag.Int("cooldown", 300, "Cooldown periiod in seconds between notifications")
	flag.Parse()

	// Initialize configuration
	config := &Config{
		Threshold:      *threshold,
		CheckInterval:  *checkInterval,
		NotifyCommand:  "zenity",
		BatteryPath:    *batteryPath,
		CooldownPeriod: time.Duration(*cooldown) * time.Second,
	}

	log.Printf("Starting battery monitor with threshold: %d%%, interval: %ds\n", config.Threshold, config.CheckInterval)

	// Start monitoring
	for {
		batteryPercent, charging, err := getBatteryInfo(config.BatteryPath)
		if err != nil {
			log.Printf("Error reading battery info: %v", err)
		} else {
			log.Printf("Battery at %d%%, charging: %v", batteryPercent, charging)

			// Check if notification should be sent
			if batteryPercent <= config.Threshold && !charging {
				if time.Since(config.LastNotifiedAt) >= config.CooldownPeriod {
					sendNotification(config.NotifyCommand, batteryPercent, *batteryPath)
					config.LastNotifiedAt = time.Now()
				}
			}
		}

		// Wait for the next check
		time.Sleep(time.Duration(config.CheckInterval) * time.Second)
	}
}

// Reads battery status from the system
func getBatteryInfo(batteryPath string) (int, bool, error) {
	// Read battery capacity (percentage)
	capacityFile := batteryPath + "/capacity"
	capacityBytes, err := os.ReadFile(capacityFile)
	if err != nil {
		return 0, false, fmt.Errorf("Failed to read battery capacity: %v", err)
	}

	capacity, err := strconv.Atoi(strings.TrimSpace(string(capacityBytes)))
	if err != nil {
		return 0, false, fmt.Errorf("Invalid battery capacity value: %v", err)
	}

	// Read battery status
	statusFile := batteryPath + "/status"
	statusBytes, err := os.ReadFile(statusFile)
	if err != nil {
		return capacity, false, fmt.Errorf("Failed to read battery status %v", err)
	}

	status := strings.TrimSpace(string(statusBytes))
	charging := status == "Charging" || status == "Full"

	return capacity, charging, nil
}

// Sennds a desktop notification
func sendNotification(command string, batteryPercent int, batteryPath string) {
	title := "Low Battery"
	message := fmt.Sprintf("Battery level is at %d%%", batteryPercent)

	cmd := exec.Command(command, "--warning", "--text", message, "--width=300", "--height=100", "--title", title)
	cmd.Env = append(os.Environ(), "GTK_THEME=Adwaita:dark")
	err := cmd.Start()
	if err != nil {
		log.Printf("Failed to send notification %v", err)
		return
	}

	// Check for charging status to dismiss notiification
	for {
		_, charging, err := getBatteryInfo(batteryPath)
		if err != nil {
			log.Printf("Error reading batter info: %v", err)
			break
		}

		if charging {
			cmd.Process.Kill()
			break
		}

		time.Sleep(2 * time.Second)
	}
}
