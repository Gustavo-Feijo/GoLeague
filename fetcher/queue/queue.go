package queue

var regions = map[string][]string{
	"AMERICAS": {"BR1", "LA1", "LA2", "NA1"},
	"EUROPE":   {"EUN1", "EUW1", "TR1", "ME1", "RU"},
	"ASIA":     {"KR", "JP1"},
	"SEA":      {"OC1", "SG2", "TW2", "VN2"},
}

type QueueManager struct {
	rateLimiter *RateLimiter
}

func startQueue() {
}

func runMainRegion(region string) {
}

func runSubRegion(region string) {
}
