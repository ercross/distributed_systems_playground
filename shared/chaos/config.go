package chaos

import "strings"

type Config struct {
	ServiceName string

	LatencyEnabled bool
	MinLatencyMs   int
	MaxLatencyMs   int

	FailureEnabled bool
	FailureRate    float64

	TimeoutEnabled bool
	TimeoutMs      int
	TimeoutRate    float64

	CrashEnabled bool
	CrashRate    float64

	PacketLossEnabled bool
	PacketLossRate    float64
}

func NewConfig(serviceName string) *Config {
	prefix := strings.ToUpper(serviceName) + "_"

	return &Config{
		ServiceName:       serviceName,
		LatencyEnabled:    envBool(prefix+"LATENCY_ENABLED", false),
		MinLatencyMs:      envInt(prefix+"LATENCY_MIN_MS", 50),
		MaxLatencyMs:      envInt(prefix+"LATENCY_MAX_MS", 500),
		FailureEnabled:    envBool(prefix+"FAILURE_ENABLED", false),
		FailureRate:       envFloat(prefix+"FAILURE_RATE", 0.3),
		TimeoutEnabled:    envBool(prefix+"TIMEOUT_ENABLED", false),
		TimeoutMs:         envInt(prefix+"TIMEOUT_MS", 3000),
		TimeoutRate:       envFloat(prefix+"TIMEOUT_RATE", 0.1),
		CrashEnabled:      envBool(prefix+"CRASH_ENABLED", false),
		CrashRate:         envFloat(prefix+"CRASH_RATE", 0.01),
		PacketLossEnabled: envBool(prefix+"PACKET_LOSS_ENABLED", false),
		PacketLossRate:    envFloat(prefix+"PACKET_LOSS_RATE", 0.01),
	}
}
