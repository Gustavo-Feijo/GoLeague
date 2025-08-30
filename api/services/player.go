package services

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"goleague/api/cache"
	"goleague/api/dto"
	"goleague/api/filters"
	"goleague/api/repositories"
	"goleague/pkg/messages"
	"goleague/pkg/redis"
	"strconv"
	"strings"
	"time"

	pb "goleague/pkg/grpc"

	tiervalues "goleague/pkg/riotvalues/tier"

	"google.golang.org/grpc"
	"google.golang.org/grpc/status"
	"gorm.io/gorm"
)

// PlayerService service with the  repositories and the gRPC client in case we need to force fetch something (Unlikely).
type PlayerService struct {
	db         *gorm.DB
	grpcClient *grpc.ClientConn
	matchCache *cache.MatchCache
	redis      *redis.RedisClient

	MatchRepository  repositories.MatchRepository
	PlayerRepository repositories.PlayerRepository
}

type PlayerServiceDeps struct {
	DB         *gorm.DB
	GrpcClient *grpc.ClientConn
	MatchCache *cache.MatchCache
	Redis      *redis.RedisClient
}

// NewPlayerService creates a service for handling player services.
func NewPlayerService(deps *PlayerServiceDeps) (*PlayerService, error) {
	// Create the repository.
	repo, err := repositories.NewPlayerRepository(deps.DB)
	if err != nil {
		return nil, errors.New("failed to start the player repository")
	}

	matchRepo, err := repositories.NewMatchRepository(deps.DB)
	if err != nil {
		return nil, errors.New("failed to start the match repository")
	}

	return &PlayerService{
		db:               deps.DB,
		grpcClient:       deps.GrpcClient,
		matchCache:       deps.MatchCache,
		MatchRepository:  matchRepo,
		PlayerRepository: repo,
		redis:            deps.Redis,
	}, nil
}

// createPlayerRateLimitKey generates a consistent hash-based key for rate limiting
func (ps *PlayerService) createPlayerRateLimitKey(gameName, gameTag, region, prefix string) string {
	keyData := fmt.Sprintf("%s|%s|%s",
		strings.ToLower(gameName),
		strings.ToLower(gameTag),
		strings.ToLower(region))

	hasher := sha256.New()
	hasher.Write([]byte(keyData))
	keyHash := hex.EncodeToString(hasher.Sum(nil)) // Use full hash for safety

	return fmt.Sprintf("%s:%s", prefix, keyHash)
}

// checkRateLimit checks if a rate limit is active and returns TTL if blocked.
func (ps *PlayerService) checkRateLimit(ctx context.Context, rateLimitKey string, lockDuration time.Duration) error {
	lockAcquired, err := ps.redis.SetNX(ctx, rateLimitKey, "processing", lockDuration).Result()
	if err != nil {
		return fmt.Errorf("couldn't check rate limits on redis: %w", err)
	}

	if !lockAcquired {
		ttl, err := ps.redis.TTL(ctx, rateLimitKey).Result()
		if err != nil {
			return fmt.Errorf(messages.OperationInProgress)
		}

		switch {
		case ttl == -2:
			// Key doesn't exist (race condition)
			return fmt.Errorf("request conflict detected, please retry")
		case ttl == -1:
			// Key exists but no expiration
			return fmt.Errorf(messages.OperationInProgress)
		case ttl > 0:
			return fmt.Errorf("operation already in progress, try again in %d seconds", int(ttl.Seconds()))
		default:
			return fmt.Errorf(messages.OperationInProgress)
		}
	}

	return nil
}

// GetPlayerSearch returns the result of a given search.
func (ps *PlayerService) GetPlayerSearch(filters *filters.PlayerSearchFilter) ([]*dto.PlayerSearch, error) {
	players, err := ps.PlayerRepository.SearchPlayer(filters)
	if err != nil {
		return nil, err
	}

	playerDto := make([]*dto.PlayerSearch, len(players))
	for key, player := range players {
		playerDto[key] = &dto.PlayerSearch{
			Id:            player.ID,
			Name:          player.RiotIdGameName,
			ProfileIcon:   player.ProfileIcon,
			Puuid:         player.Puuid,
			Region:        string(player.Region),
			SummonerLevel: player.SummonerLevel,
			Tag:           player.RiotIdTagline,
		}
	}

	return playerDto, nil
}

// GetPlayerMatchHistory returns a player match list based on filters.
func (ps *PlayerService) GetPlayerMatchHistory(filters *filters.PlayerMatchHistoryFilter) (dto.MatchPreviewList, error) {
	// Convert to string.
	// Received through path params.
	name := filters.GameName
	tag := filters.GameTag
	region := filters.Region

	playerId, err := ps.PlayerRepository.GetPlayerIdByNameTagRegion(name, tag, region)
	if err != nil {
		return nil, fmt.Errorf(messages.CouldNotFindId+": %w", "player", err)
	}

	filters.PlayerId = &playerId
	matchesIds, err := ps.PlayerRepository.GetPlayerMatchHistoryIds(filters)
	if err != nil {
		return nil, fmt.Errorf("couldn't get the match ids: %w", err)
	}

	if len(matchesIds) == 0 {
		return nil, nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	// Get all the cached matches previews.
	// Match previews shouldn't change, using cache to reduce load into the database.
	cachedMatches, missingMatches, err := ps.matchCache.GetMatchesPreviewByMatchIds(ctx, matchesIds)
	if err == nil {
		// All matches in cache.
		if len(missingMatches) == 0 {
			matchPreviews := make(dto.MatchPreviewList)
			handleCachedMatches(cachedMatches, matchPreviews)
			return matchPreviews, nil
		}

		matchesIds = missingMatches
	}

	// Get the non cached matches from the database.
	matchPreviews, err := ps.MatchRepository.GetMatchPreviews(matchesIds)
	if err != nil {
		return nil, fmt.Errorf("couldn't get the match history for the player: %w", err)
	}

	formatedPreviews := formatMatchPreviews(matchPreviews)

	// Some matches came from cache, others from db.
	if len(missingMatches) > 0 {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		for _, match := range formatedPreviews {
			ps.matchCache.SetMatchPreview(ctx, *match)
		}
		handleCachedMatches(cachedMatches, formatedPreviews)
	}

	return formatedPreviews, nil
}

func (ps *PlayerService) GetPlayerInfo(filters *filters.PlayerInfoFilter) (*dto.FullPlayerInfo, error) {
	name := filters.GameName
	tag := filters.GameTag
	region := filters.Region

	playerId, err := ps.PlayerRepository.GetPlayerIdByNameTagRegion(name, tag, region)
	if err != nil {
		return nil, fmt.Errorf(messages.CouldNotFindId+": %w", "player", err)
	}

	playerInfo, err := ps.PlayerRepository.GetPlayerById(playerId)
	if err != nil {
		return nil, fmt.Errorf("couldn't get the player info: %w", err)
	}

	playerRatings, err := ps.PlayerRepository.GetPlayerRatingsById(playerId)
	if err != nil {
		return nil, fmt.Errorf("couldn't get the player rating: %w", err)
	}

	fullPlayerInfo := dto.FullPlayerInfo{
		Id:            playerInfo.ID,
		Name:          playerInfo.RiotIdGameName,
		ProfileIcon:   playerInfo.ProfileIcon,
		Puuid:         playerInfo.Puuid,
		Region:        string(playerInfo.Region),
		SummonerLevel: playerInfo.SummonerLevel,
		Tag:           playerInfo.RiotIdTagline,
	}

	if len(playerRatings) > 0 {
		ratings := make([]dto.RatingInfo, len(playerRatings))

		for key, rating := range playerRatings {
			ratings[key] = dto.RatingInfo{
				LeaguePoints: rating.LeaguePoints,
				Losses:       rating.Losses,
				Queue:        rating.Queue,
				Rank:         rating.Rank,
				Region:       string(rating.Region),
				Tier:         rating.Tier,
				Wins:         rating.Wins,
			}
		}

		fullPlayerInfo.Rating = ratings
	}

	return &fullPlayerInfo, nil
}

// GetPlayerStats returns the player stats for a given player.
func (ps *PlayerService) GetPlayerStats(filters *filters.PlayerStatsFilter) (dto.FullPlayerStats, error) {
	name := filters.GameName
	tag := filters.GameTag
	region := filters.Region

	playerId, err := ps.PlayerRepository.GetPlayerIdByNameTagRegion(name, tag, region)
	if err != nil {
		return nil, fmt.Errorf(messages.CouldNotFindId+": %w", "player", err)
	}

	filters.PlayerId = &playerId
	playerStats, err := ps.PlayerRepository.GetPlayerStats(filters)
	if err != nil {
		return nil, fmt.Errorf("couldn't get the player stats: %w", err)
	}

	if len(playerStats) == 0 {
		return nil, nil
	}

	playerStatsDto := make(dto.FullPlayerStats)

	for _, stats := range playerStats {
		parsePlayerStats(playerStatsDto, stats)
	}

	return playerStatsDto, nil
}

// ForceFetchPlayer makes a gRPC requets to the fetcher to forcefully get data from a Player.
func (ps *PlayerService) ForceFetchPlayer(filters *filters.PlayerForceFetchFilter) (*pb.Summoner, error) {
	client := pb.NewServiceClient(ps.grpcClient)

	grpcCall := func(ctx context.Context, req *pb.SummonerRequest) (any, error) {
		return client.FetchSummonerData(ctx, req)
	}

	resp, err := ps.executeSummonerGRPCCall(filters.GameName, filters.GameTag, filters.Region, "force_fetch_player", grpcCall)
	if err != nil {
		return nil, err
	}

	return resp.(*pb.Summoner), nil
}

// ForceFetchPlayer makes a gRPC requets to the fetcher to forcefully get data from a Player.
func (ps *PlayerService) ForceFetchPlayerMatchHistory(filters *filters.PlayerForceFetchMatchListFilter) (*pb.MatchHistoryFetchNotification, error) {
	client := pb.NewServiceClient(ps.grpcClient)

	grpcCall := func(ctx context.Context, req *pb.SummonerRequest) (any, error) {
		return client.FetchMatchHistory(ctx, req)
	}

	resp, err := ps.executeSummonerGRPCCall(filters.GameName, filters.GameTag, filters.Region, "force_fetch_player_matches", grpcCall)
	if err != nil {
		return nil, err
	}

	return resp.(*pb.MatchHistoryFetchNotification), nil
}

// executeSummonerGRPCCall is a helper to execute any gRPC call for summoner requests.
func (ps *PlayerService) executeSummonerGRPCCall(
	gameName string,
	gameTag string,
	region string,
	operation string,
	grpcCall func(context.Context, *pb.SummonerRequest) (any, error),
) (any, error) {
	rateLimitKey := ps.createPlayerRateLimitKey(gameName, gameTag, region, operation)
	redisCtx, cancelRedis := context.WithTimeout(context.Background(), time.Second)
	defer cancelRedis()
	if err := ps.checkRateLimit(redisCtx, rateLimitKey, time.Minute*5); err != nil {
		return nil, err
	}

	request := &pb.SummonerRequest{
		GameName: gameName,
		TagLine:  gameTag,
		Region:   region,
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	resp, err := grpcCall(ctx, request)
	if err != nil {
		st, ok := status.FromError(err)
		if ok {
			return nil, fmt.Errorf("couldn't execute %s: %w", operation, errors.New(st.Message()))
		}
		return nil, fmt.Errorf("couldn't execute %s: %w", operation, err)
	}

	return resp, nil
}

// formatMatchPreviews return the formatted dto for the matches.
func formatMatchPreviews(rawPreviews []repositories.RawMatchPreview) dto.MatchPreviewList {
	fullPreview := make(dto.MatchPreviewList)

	// Range through each raw preview and format it.
	for _, r := range rawPreviews {
		parsePreviewData(fullPreview, r)
	}

	return fullPreview
}

// parsePlayerStats parse a raw player stats entry to insert into the DTO.
func parsePlayerStats(playerStatsDto dto.FullPlayerStats, stats repositories.RawPlayerStatsStruct) {
	var champion string
	var lane string
	var queue string

	// Handle value conversion.
	if stats.ChampionId == -1 {
		champion = "ALL"
	} else {
		champion = strconv.Itoa(stats.ChampionId)
	}

	if stats.TeamPosition == "" || stats.TeamPosition == "ALL" {
		lane = "ALL"
	} else {
		lane = stats.TeamPosition
	}

	if stats.QueueId == -1 {
		queue = "ALL"
	} else {
		queue = strconv.Itoa(stats.QueueId)
	}

	// Initialize the queue entry if it doesn't exist
	if playerStatsDto[queue] == nil {
		playerStatsDto[queue] = &dto.PlayerStatsQueue{
			Unfiltered:   nil,
			ChampionData: make(map[string]*dto.StatsEntry),
			LaneData:     make(map[string]*dto.StatsEntry),
		}
	}

	entry := &dto.StatsEntry{
		AverageAssists: stats.AverageAssists,
		AverageDeaths:  stats.AverageDeaths,
		AverageKills:   stats.AverageKills,
		CsPerMin:       stats.CsPerMin,
		KDA:            stats.KDA,
		Matches:        stats.Matches,
		WinRate:        stats.WinRate,
	}

	// Only add the entries with no champion or lane filter.
	if stats.AggregationLevel == "by_queue" || stats.AggregationLevel == "overall" {
		playerStatsDto[queue].Unfiltered = entry
		return
	}

	// Add the stats entries
	if champion != "ALL" {
		playerStatsDto[queue].ChampionData[champion] = entry
	}

	if lane != "ALL" {
		playerStatsDto[queue].LaneData[lane] = entry
	}
}

// parsePreviewData simply parse one raw preview entry to a full preview.
func parsePreviewData(fullPreview dto.MatchPreviewList, r repositories.RawMatchPreview) {
	// Initialize the full preview.
	if _, ok := fullPreview[r.MatchID]; !ok {
		fullPreview[r.MatchID] = &dto.MatchPreview{
			Metadata: &dto.MatchPreviewMetadata{
				AverageElo: tiervalues.CalculateInverseRank(int(r.AverageRating)),
				Date:       r.Date,
				Duration:   r.Duration,
				InternalId: r.InternalId,
				MatchId:    r.MatchID,
				QueueId:    r.QueueID,
			},
			Data: make([]*dto.MatchPreviewData, 0),
		}
	}

	rawItems := []*int{r.Item0, r.Item1, r.Item2, r.Item3, r.Item4, r.Item5}
	items := make([]int, 0, 6)
	for _, it := range rawItems {
		if it != nil && *it != 0 {
			items = append(items, *it)
		}
	}

	preview := &dto.MatchPreviewData{
		GameName:      r.RiotIDGameName,
		Tag:           r.RiotIDTagline,
		Region:        r.Region,
		Assists:       r.Assists,
		Kills:         r.Kills,
		Deaths:        r.Deaths,
		ChampionLevel: r.ChampionLevel,
		ChampionID:    r.ChampionID,
		TotalCs:       r.TotalMinionsKilled + r.NeutralMinionsKilled,
		Items:         items,
		Win:           r.Win,
		QueueID:       r.QueueID,
	}

	fullPreview[r.MatchID].Data = append(fullPreview[r.MatchID].Data, preview)
}

func handleCachedMatches(cachedMatches []dto.MatchPreview, matchesDto dto.MatchPreviewList) {
	for _, match := range cachedMatches {
		m := match
		matchesDto[match.Metadata.MatchId] = &m
	}
}
