package pkg

import (
	"fmt"
	"strings"
	"time"

	"github.com/kiwikid/openscadgen/pkg/models"
	"github.com/pterm/pterm"
)

type ProgressReporter interface {
	Update(msg string)
	Done()
	Error(err error)
	// Simplified progress bar methods
	Construct(instances []models.InstanceConfig, nonSkippedInstances int)
	StartInstance(instanceId string, name string, instanceIndex int, instanceCount int)
	//ProgressInstance(instanceId string, progress int)
	FinishInstance()
}

type NoopProgress struct{}

func (n *NoopProgress) Update(msg string)                                                    {}
func (n *NoopProgress) Done()                                                                {}
func (n *NoopProgress) Error(err error)                                                      {}
func (n *NoopProgress) Construct(instances []models.InstanceConfig, nonSkippedInstances int) {}
func (n *NoopProgress) StartInstance(instanceId string, name string, instanceIndex int, instanceCount int) {
}
func (n *NoopProgress) ProgressInstance(instanceId string, progress int) {}
func (n *NoopProgress) FinishInstance()                                  {}

type ChanProgress struct {
	Updates         chan<- string
	instances       map[string]models.InstanceConfig
	currentInstance string
}

func (c *ChanProgress) Update(msg string) { c.Updates <- msg }
func (c *ChanProgress) Done()             { c.Updates <- "done 2" }
func (c *ChanProgress) Error(err error)   { c.Updates <- "error: " + err.Error() }
func (c *ChanProgress) Construct(instances []models.InstanceConfig) {
	c.instances = make(map[string]models.InstanceConfig)
	for _, instance := range instances {
		c.instances[instance.AutoName] = instance
	}
	c.currentInstance = instances[0].AutoName
	c.Updates <- fmt.Sprintf("Constructed progress for %d instances", len(instances))
}

func (c *ChanProgress) StartInstance(instanceId string, name string, instanceIndex int, instanceCount int) {
	instance, exists := c.instances[instanceId]
	if !exists {
		//	c.Updates <- fmt.Sprintf("Starting: %s", instanceId)
		return
	}
	c.currentInstance = instance.AutoName

	// Check if instance has camera configurations
	if len(instance.ExportImages) > 0 {
		cameraNames := make([]string, len(instance.ExportImages))
		for i, img := range instance.ExportImages {
			cameraNames[i] = img.CameraName
		}
		//c.Updates <- fmt.Sprintf("Starting: %s - cameras: %v", instance.AutoName, cameraNames)
	} else {
		//c.Updates <- fmt.Sprintf("Starting: %s", instance.AutoName)
	}
}

/*
	func (c *ChanProgress) ProgressInstance(instanceId string, progress int) {
		instance, exists := c.instances[instanceId]
		if exists && len(instance.ExportImages) > 0 {
			cameraNames := make([]string, len(instance.ExportImages))
			for i, img := range instance.ExportImages {
				cameraNames[i] = img.CameraName
			}
			c.Updates <- fmt.Sprintf("Progress: %s - cameras: %v (%d%%)", instanceId, cameraNames, progress)
		} else {
			c.Updates <- fmt.Sprintf("Progress: %s (%d%%)", instanceId, progress)
		}
	}
*/
func (c *ChanProgress) FinishInstance() {
	c.Updates <- "Instance complete" + c.currentInstance
}

// TerminalProgressReporter provides simplified terminal progress reporting
type TerminalProgressReporter struct {
	isQuiet            bool
	currentInstance    string
	currentCamera      string
	completedInstances int
	totalInstances     int
	instances          map[string]models.InstanceConfig
	// Progress bars
	totalProgress *pterm.ProgressbarPrinter
	//	currentProgress *pterm.ProgressbarPrinter
	// Timer for current instance progress
	instanceStartTime time.Time
	stopTimer         chan struct{}
}

func NewTerminalProgressReporter(config *models.Config) *TerminalProgressReporter {
	reporter := &TerminalProgressReporter{
		isQuiet:   config.Quiet,
		stopTimer: make(chan struct{}),
	}

	if !config.Quiet {
		// Create total progress bar
		reporter.totalProgress, _ = pterm.DefaultProgressbar.WithTotal(0).
			WithTitle("\nTotal Progress").
			WithShowElapsedTime(true).
			WithShowCount(true).
			WithBarStyle(pterm.NewStyle(pterm.FgGreen)).
			WithTitleStyle(pterm.NewStyle(pterm.FgCyan, pterm.Bold)).
			Start("Total Progress\n")

		// Create current instance progress bar
		/*reporter.currentProgress, _ = pterm.DefaultProgressbar.WithTotal(100).
		WithTitle("Current Instance").
		WithShowElapsedTime(true).
		WithShowCount(true).
		WithBarStyle(pterm.NewStyle(pterm.FgBlue)).
		WithTitleStyle(pterm.NewStyle(pterm.FgYellow, pterm.Bold)).
		Start("Current Instance")*/
	}

	return reporter
}

func (t *TerminalProgressReporter) Update(msg string) {
	if !t.isQuiet {
		pterm.Info.Println(msg)
	}
}

func (t *TerminalProgressReporter) Construct(instances []models.InstanceConfig, nonSkippedInstances int) {
	t.instances = make(map[string]models.InstanceConfig)
	for _, instance := range instances {
		t.instances[instance.AutoName] = instance
	}
	if !t.isQuiet {
		// Set the total for the total progress bar
		if t.totalProgress != nil {
			t.totalProgress.Total = nonSkippedInstances
			t.totalInstances = nonSkippedInstances
		}
		pterm.Info.Printf("Processing %d instances\n", nonSkippedInstances)
	}
}

func (t *TerminalProgressReporter) StartInstance(instanceId string, name string, instanceIndex int, instanceCount int) {
	t.currentInstance = instanceId

	// Get instance details if available
	instance, exists := t.instances[instanceId]
	if exists && len(instance.ExportImages) > 0 {
		cameraNames := make([]string, len(instance.ExportImages))
		for i, img := range instance.ExportImages {
			cameraNames[i] = img.CameraName
		}
		t.currentCamera = strings.Join(cameraNames, ", ")
	} else {
		t.currentCamera = ""
	}

	if !t.isQuiet {
		// Stop previous timer if running
		select {
		case t.stopTimer <- struct{}{}:
		default:
		}

		// Reset current progress bar for new instance
		/*	if t.currentProgress != nil {
			t.currentProgress.Current = 0
			t.currentProgress.Total = 100
			if t.currentCamera != "" {
				t.currentProgress.UpdateTitle(fmt.Sprintf("Current Instance: %s - %s", instanceId, t.currentCamera))
			} else {
				t.currentProgress.UpdateTitle(fmt.Sprintf("Current Instance: %s", instanceId))
			}
		}*/

		// Start timer for time-based progress
		t.instanceStartTime = time.Now()
		go t.updateInstanceProgress()

		/*if t.currentCamera != "" {
			pterm.Info.Printf("Starting: %s - %s\n", instanceId, t.currentCamera)
		} else {
			pterm.Info.Printf("Starting: %s\n", instanceId)
		}*/
	}
}

func (t *TerminalProgressReporter) ProgressInstance(instanceId string, progress int) {
	if !t.isQuiet {
		// For time-based progress, we don't update the bar here
		// The timer goroutine handles the progress updates
		if t.currentCamera != "" {
			pterm.Info.Printf("Progress: %s - %s (%d%%)\n", instanceId, t.currentCamera, progress)
		} else {
			pterm.Info.Printf("Progress: %s (%d%%)\n", instanceId, progress)
		}
	}
}

func (t *TerminalProgressReporter) FinishInstance() {
	t.completedInstances++
	if !t.isQuiet {
		// Stop the timer
		select {
		case t.stopTimer <- struct{}{}:
		default:
		}

		// Update total progress bar
		if t.totalProgress != nil {
			t.totalProgress.Increment()
		}

		//percentage := int(float64(t.completedInstances) / float64(t.totalInstances) * 100)
		//pterm.Info.Printf("Completed. - %s (%d/%d - %d%%)\n", t.currentInstance, t.completedInstances, t.totalInstances, percentage)
	}
}

func (t *TerminalProgressReporter) Done() {
	if !t.isQuiet {
		// Stop the timer
		select {
		case t.stopTimer <- struct{}{}:
		default:
		}

		pterm.Info.Printf("Processing complete: %d/%d instances\n", t.completedInstances, t.totalInstances)
	}
}

func (t *TerminalProgressReporter) Error(err error) {
	if !t.isQuiet {
		pterm.Error.Println(err.Error())
	}
}

func (t *TerminalProgressReporter) updateInstanceProgress() {
	if t.isQuiet {
		return
	}

	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			// Calculate progress based on elapsed time
			elapsed := time.Since(t.instanceStartTime)
			estimatedTotal := 5 * time.Second // Assume 5 seconds per instance

			progress := int((float64(elapsed) / float64(estimatedTotal)) * 100)
			if progress > 95 { // Cap at 95% until explicitly completed
				progress = 95
			}

			// Update progress bar
			/*targetProgress := int(float64(progress) / 100.0 * float64(t.currentProgress.Total))
			toIncrement := targetProgress - t.currentProgress.Current
			if toIncrement > 0 {
				for i := 0; i < toIncrement; i++ {
					t.currentProgress.Increment()
			}*/
		case <-t.stopTimer:
			return
		}
	}
}
