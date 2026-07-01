//go:build !linux && !darwin

package deps

func installHint(binary string) string {
	return "install " + binary + " using your system package manager"
}
