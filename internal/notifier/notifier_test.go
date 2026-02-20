package notifier

import "testing"

func TestNew(t *testing.T) {
	n := New()
	if n == nil {
		t.Fatal("New() returned nil")
	}
}

func TestFallbackNotify(t *testing.T) {
	f := &Fallback{}
	if err := f.Notify("test", "body"); err != nil {
		t.Fatal(err)
	}
}
