package dto

// Result of a tierlist fetch.
type TierlistResult struct {
	BanCount     int
	BanRate      float64
	ChampionId   int
	PickCount    int
	PickRate     float64
	TeamPosition string
	WinRate      float64
}

// The fulltierlist we will return.
type FullTierlist struct {
	Champion     map[string]any // Will be a simplified version of champion.Champion, without spell data.
	BanCount     int
	Banrate      float64
	PickCount    int
	PickRate     float64
	TeamPosition string
	WinRate      float64
}
