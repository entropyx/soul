package env

import "os"

const (
	ModeTest       = "test"
	ModeDev        = "dev"
	ModeProduction = "production"
	ModeStaging    = "staging"
	ModeDebug      = "debug"
)

var Mode string

func init() {
	mode := os.Getenv("ENV")
	if mode == "" {
		mode = ModeTest
	}
	Mode = mode
}
