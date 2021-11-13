package naruse

import "runtime"

const (
	version = "0.1.0"
	usage   = "A relay for VMess."
)

func Version() string {
	return version
}

func Usage() string {
	return usage
}

func Info() string {
	return "Naruse " + Version() + " (" + Usage() + ") (" + runtime.Version() + " " +
		runtime.GOOS + "/" + runtime.GOARCH + ")"
}
