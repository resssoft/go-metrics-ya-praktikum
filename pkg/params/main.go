package params

import "time"
import "os"

func DurationByEnv(d time.Duration, evnName string) time.Duration {
	if os.Getenv(evnName) != "" {
		envVal := os.Getenv(evnName)
		parsedDuration, err := time.ParseDuration(envVal)
		if err == nil {
			d = parsedDuration
		} else {
			envVal += "s"
			parsedDuration, err := time.ParseDuration(envVal)
			if err == nil {
				d = parsedDuration
			}
		}
	}
	return d
}

func StrByEnv(s, evnName string) string {
	if os.Getenv(evnName) != "" {
		s = os.Getenv(evnName)
	}
	return s
}

func BoolByEnv(s bool, evnName string) bool {
	if os.Getenv(evnName) != "" {
		s = os.Getenv(evnName) == "true"
	}
	return s
}
