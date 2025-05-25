package queuevalues

var RankedQueueValue = map[int]string{
	420: "RANKED_SOLO_5x5",
	440: "RANKED_FLEX_5x5",
}

// Queues that are going to be stored.
var TreatedQueues = []int{400, 420, 430, 440, 450, 490, 700, 720, 900, 1020, 1300, 1400, 1700, 1710, 1900}

// Queues that have defined positions.
// Need to verify again to see if they really have.
// Needed to verify if the team_position value is valid or not. Sometimes could be "".
var QueuesWithPositions = []int{400, 420, 430, 440, 490, 700, 900, 1020, 1400, 1900}
