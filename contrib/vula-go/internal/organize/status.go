package organize

import (
	"fmt"
	"strings"
	"time"

	"codeberg.org/vula/vula/contrib/vula-go/internal/core"
)

var unitNames = []string{
	"vula-publish.service",
	"vula-discover.service",
	"vula-organize.service",
}

func Status(systemd core.Systemd, meta core.DbusMeta) (StyledOutput, error) {
	output := StyledOutput{}

	for _, unitName := range unitNames {
		status, text := summarizeUnit(systemd, unitName)
		statusLine := formatOutputLine(status, text)
		output.AppendStyledOutput(&statusLine)
		output.AddNewline()
	}

	hasOwner, err := meta.NameHasOwner("local.vula.organize")
	if err != nil {
		return nil, err
	}

	if hasOwner {
		statusLine := formatOutputLine("active", "local.vula.organize dbus service")
		output.AppendStyledOutput(&statusLine)
	}

	return output, nil
}

func formatOutputLine(status string, text string) StyledOutput {
	output := StyledOutput{}
	var color Color
	switch strings.TrimSpace(status) {
	case "active":
		color = GreenColor
	case "inactive":
		fallthrough
	case "activatable":
		color = YellowColor
	default:
		color = RedColor
	}

	output.Add(Text("["))
	output.Add(Colored(centerText(status, 8), color))
	output.Add(Text("] "))
	output.Add(Text(text))
	return output
}

// summarizeUnit returns the status and the text to be printed.
func summarizeUnit(s core.Systemd, unitName string) (string, string) {
	// Note: setting the status to "error" deviates from the python version of vula. Instead,
	// the reason for the error is included in the text which seems more reasonable. Also,
	// no duration is calculated for non-active units as it wouldn't make sense.
	unit, err := s.GetUnit(unitName)
	if err != nil {
		if strings.Contains(err.Error(), "not loaded") {
			return "disabled", unitName
		}
		return "error", fmt.Sprintf("%s (%s)", unitName, err)
	}

	unitState, err := unit.GetActiveState()
	if err != nil {
		return "error", fmt.Sprintf("%s (%s)", unitName, err)
	}

	if unitState != "active" {
		return unitState, unitName
	}

	timeStamp, err := unit.GetStateChangeTimeStamp()
	if err != nil {
		return unitState, fmt.Sprintf("%-35s (couldn't get duration)", unitName)
	}

	duration := time.Since(timeStamp).Round(time.Second)
	return unitState, fmt.Sprintf("%-35s (%s)", unitName, formatDuration(duration))
}

func formatDuration(d time.Duration) string {
	totalSeconds := int(d.Seconds())
	hours := totalSeconds / 3600
	remaining := totalSeconds % 3600
	minutes := remaining / 60
	seconds := remaining % 60
	return fmt.Sprintf("%d:%02d:%02d", hours, minutes, seconds)
}

func centerText(s string, width int) string {
	if len(s) >= width {
		return fmt.Sprintf("%-*s", width, s)
	}
	if width <= 0 {
		return ""
	}
	pad := width - len(s)
	left := pad / 2
	right := pad - left
	return strings.Repeat(" ", left) + s + strings.Repeat(" ", right)
}
