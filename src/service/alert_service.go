package service

import (
	"fmt"
	"os"
	"time"

	"github.com/h3llmy/system-monitoring/src/response"
	"github.com/h3llmy/system-monitoring/src/utils/mail"
)

type AlertService struct {
	Mailer         *mail.Mail
	lastCpuSent    time.Time
	lastMemorySent time.Time
	lastTempSent   time.Time
	alertCooldown  time.Duration
}

// NewAlertService creates a new AlertService instance that is responsible for sending email alerts.
// It takes a mailer dependency which is used to send emails.
// The returned AlertService has a cooldown period of 1 minute per alert type,
// meaning that at most one email will be sent per minute per alert type.
func NewAlertService(mailer *mail.Mail) *AlertService {
	return &AlertService{
		Mailer:        mailer,
		alertCooldown: 5 * time.Minute, // send at most once per minute per alert type
	}
}

// StartMonitoringAlert starts a goroutine that sends email alerts based on system metrics.
// It starts three goroutines: CpuAlert, MemoryAlert, and TemperatureAlert.
// Each goroutine sends email alerts based on their respective system metrics,
// with a cooldown period of 1 minute per alert type, meaning that at most
// one email will be sent per minute per alert type.
func (s *AlertService) StartMonitoringAlert() error {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		go s.CpuAlert()
		go s.MemoryAlert()
		go s.TemperatureAlert()
	}

	return nil
}

// CpuAlert is a goroutine that sends email alerts based on system CPU usage.
// It sends an email alert if the average CPU usage over the last maxHistory metrics
// is higher than 80%. The email alert is sent at most once per minute.
// The email alert is sent to the email address specified by the NOTIFICATION_EMAIL environment
// variable. The email alert has the subject "CPU Alert" and the body "The CPU usage is high".
func (s *AlertService) CpuAlert() error {
	if len(*systemHistory.Matrics) < maxHistory {
		return nil
	}

	var totalCpuUsage float64
	for _, m := range *systemHistory.Matrics {
		totalCpuUsage += *m.CPU
	}

	avgCpuUsage := totalCpuUsage / float64(maxHistory)
	if avgCpuUsage > 80 && time.Since(s.lastCpuSent) > s.alertCooldown {
		go s.Mailer.SendMail(os.Getenv("NOTIFICATION_EMAIL"), "CPU Alert", "The CPU usage is high")
		s.lastCpuSent = time.Now()
	}
	return nil
}

// MemoryAlert is a goroutine that sends email alerts based on system memory usage.
// It sends an email alert if the average memory usage over the last maxHistory metrics
// is higher than 80%. The email alert is sent at most once per minute.
// The email alert is sent to the email address specified by the NOTIFICATION_EMAIL environment
// variable. The email alert has the subject "Memory Alert" and the body "The memory usage is high: <percentage>%".
// The percentage is calculated by dividing the total memory used over the last maxHistory metrics
// by the total memory available on the system, and then multiplying by 100.
func (s *AlertService) MemoryAlert() error {
	if len(*systemHistory.Matrics) < maxHistory {
		return nil
	}

	var totalMemoryUsage int64
	for _, m := range *systemHistory.Matrics {
		totalMemoryUsage += m.Memory.Used
	}

	latestTotalMemory := (*systemHistory.Matrics)[len(*systemHistory.Matrics)-1].Memory.Total

	avgMemoryUsed := float64(totalMemoryUsage) / float64(maxHistory)
	avgMemoryUsagePercent := (avgMemoryUsed / float64(latestTotalMemory)) * 100

	if avgMemoryUsagePercent > 80 && time.Since(s.lastMemorySent) > s.alertCooldown {
		go s.Mailer.SendMail(
			os.Getenv("NOTIFICATION_EMAIL"),
			"Memory Alert",
			fmt.Sprintf("The memory usage is high: %.2f%%", avgMemoryUsagePercent),
		)
		s.lastMemorySent = time.Now()
	}

	return nil
}

// TemperatureAlert is a goroutine that sends email alerts based on system temperatures.
// It sends an email alert if any of the temperatures of the CPU, GPU, or CORE
// are higher than their respective high temperatures. The email alert is sent at most
// once per minute. The email alert is sent to the email address specified by the
// NOTIFICATION_EMAIL environment variable. The email alert has the subject
// "Temperature Alert" and the body contains the names of the sensors with high
// temperatures.
func (s *AlertService) TemperatureAlert() error {
	if systemHistory.Temprature == nil {
		return nil
	}

	var highTemps []string

	checkTemps := func(prefix string, sensors []*response.CoreTemperatureStats) {
		if sensors == nil {
			return
		}
		for _, t := range sensors {
			if t != nil && t.Temp > t.High {
				highTemps = append(highTemps, fmt.Sprintf("%s %s", prefix, t.Name))
			}
		}
	}

	checkTemps("CPU", systemHistory.Temprature.CPU)
	checkTemps("GPU", systemHistory.Temprature.GPU)
	checkTemps("CORE", systemHistory.Temprature.Core)

	if len(highTemps) > 0 && time.Since(s.lastTempSent) > s.alertCooldown {
		msg := "The following temperatures are high:\n"
		for _, name := range highTemps {
			msg += fmt.Sprintf("- %s\n", name)
		}

		go s.Mailer.SendMail(
			os.Getenv("NOTIFICATION_EMAIL"),
			"Temperature Alert",
			msg,
		)

		s.lastTempSent = time.Now()
	}

	return nil
}
