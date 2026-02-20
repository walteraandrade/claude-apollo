package notifier

import (
	"fmt"
	"os/exec"
)

type Notifier interface {
	Notify(subject, body string) error
}

type Desktop struct{}

func (d *Desktop) Notify(subject, body string) error {
	path, err := exec.LookPath("notify-send")
	if err != nil {
		return fmt.Errorf("notify-send not found: %w", err)
	}
	return exec.Command(path, "--app-name=Apollo", subject, body).Run()
}

type Fallback struct{}

func (f *Fallback) Notify(subject, body string) error {
	return nil
}

func New() Notifier {
	if _, err := exec.LookPath("notify-send"); err == nil {
		return &Desktop{}
	}
	return &Fallback{}
}
