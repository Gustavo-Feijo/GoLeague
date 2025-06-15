package dto

// PlayerSearch is the type of a player search result.
type PlayerSearch struct {
	Id            uint
	Name          string
	ProfileIcon   int
	Puuid         string
	Region        string
	SummonerLevel int
	Tag           string
}

type PlayerId struct {
	ID uint
}
