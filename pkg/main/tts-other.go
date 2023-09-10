//go:build !windows

package main

func durationAsTTS(duration time.Duration) string {
	//no-op
	return ""
}
func ttsSay(text string) error {
	//no-op
	return err
}
