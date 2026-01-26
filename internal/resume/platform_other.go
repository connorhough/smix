//go:build !darwin && !linux

package resume

func init() {
	CurrentPlatform = nil
}
