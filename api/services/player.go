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
	"strconv"
	"strings"
	"time"

	pb "goleague/pkg/grpc"
	"goleague/pkg/redis"
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
			return fmt.Errorf("operation already in progress, please wait")
		}

		switch {
		case ttl == -2:
			// Key doesn't exist (race condition)
			return fmt.Errorf("request conflict detected, please retry")
		case ttl == -1:
			// Key exists but no expiration
			return fmt.Errorf("operation already in progress, please wait")
		case ttl > 0:
			return fmt.Errorf("operation already in progress, try again in %d seconds", int(ttl.Seconds()))
		default:
			return fmt.Errorf("operation already in progress, please wait")
		}
	}

	return nil
}

// GetPlayerSearch returns the result of a given search.
func (ps *PlayerService) GetPlayerSearch(filters map[string]any) ([]*dto.PlayerSearch, error) {
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
func (ps *PlayerService) GetPlayerMatchHistory(filters map[string]any) (dto.MatchPreviewList, error) {
	// Convert to string.
	// Received through path params.
	name := filters["gameName"].(string)
	tag := filters["gameTag"].(string)
	region := filters["region"].(string)

	playerId, err := ps.PlayerRepository.GetPlayerIdByNameTagRegion(name, tag, region)
	if err != nil {
		return nil, fmt.Errorf("couldn't find the playerId: %w", err)
	}

	filters["playerId"] = playerId
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

// GetPlayerStats returns the player stats for a given player.
func (ps *PlayerService) GetPlayerStats(filters map[string]any) (dto.FullPlayerStats, error) {
	// Convert to string.
	// Received through path params.
	name := filters["gameName"].(string)
	tag := filters["gameTag"].(string)
	region := filters["region"].(string)

	playerId, err := ps.PlayerRepository.GetPlayerIdByNameTagRegion(name, tag, region)
	if err != nil {
		return nil, fmt.Errorf("couldn't find the playerId: %w", err)
	}

	filters["playerId"] = playerId
	playerStats, err := ps.PlayerRepository.GetPlayerStats(filters)
	if err != nil {
		return nil, fmt.Errorf("couldn't get the player stats: %w", err)
	}

	if len(playerStats) == 0 {
		return nil, nil
	}

	playerStatsDto := make(dto.FullPlayerStats)

	for _, stats := range playerStats {
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
			continue
		}

		// Add the stats entries
		if champion != "ALL" {
			playerStatsDto[queue].ChampionData[champion] = entry
		}

		if lane != "ALL" {
			playerStatsDto[queue].LaneData[lane] = entry
		}
	}

	return playerStatsDto, nil
}

// ForceFetchPlayer makes a gRPC requets to the fetcher to forcefully get data from a Player.
func (ps *PlayerService) ForceFetchPlayer(filters filters.PlayerForceFetchParams) (*pb.Summoner, error) {
	rateLimitKey := ps.createPlayerRateLimitKey(filters.GameName, filters.GameTag, filters.Region, "force_fetch_player")
	redisCtx, cancelRedis := context.WithTimeout(context.Background(), time.Second)
	defer cancelRedis()
	if err := ps.checkRateLimit(redisCtx, rateLimitKey, time.Minute*5); err != nil {
		return nil, err
	}

	client := pb.NewServiceClient(ps.grpcClient)

	request := &pb.SummonerRequest{
		GameName: filters.GameName,
		TagLine:  filters.GameTag,
		Region:   filters.Region,
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	// Make the request
	resp, err := client.GetSummonerData(ctx, request)
	if err != nil {
		st, ok := status.FromError(err)
		if ok {
			return nil, fmt.Errorf("couldn't force fetch the player: %w", errors.New(st.Message()))
		}
		return nil, fmt.Errorf("couldn't force fetch the player: %w", err)
	}

	return resp, nil
}

// ForceFetchPlayer makes a gRPC requets to the fetcher to forcefully get data from a Player.
func (ps *PlayerService) ForceFetchPlayerMatchHistory(filters filters.PlayerForceFetchMatchHistoryParams) (*pb.MatchHistoryFetchNotification, error) {
	rateLimitKey := ps.createPlayerRateLimitKey(filters.GameName, filters.GameTag, filters.Region, "force_fetch_player_matches")
	redisCtx, cancelRedis := context.WithTimeout(context.Background(), time.Second)
	defer cancelRedis()
	if err := ps.checkRateLimit(redisCtx, rateLimitKey, time.Minute*5); err != nil {
		return nil, err
	}

	client := pb.NewServiceClient(ps.grpcClient)

	request := &pb.SummonerRequest{
		GameName: filters.GameName,
		TagLine:  filters.GameTag,
		Region:   filters.Region,
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	// Make the request
	resp, err := client.FetchMatchHistory(ctx, request)
	if err != nil {
		st, ok := status.FromError(err)
		if ok {
			return nil, fmt.Errorf("couldn't force fetch the player match history: %w", st.Err())
		}
		return nil, fmt.Errorf("couldn't force fetch the player match history: %w", err)
	}

	return resp, nil
}

// formatMatchPreviews return the formatted dto for the matches.
func formatMatchPreviews(rawPreviews []repositories.RawMatchPreview) dto.MatchPreviewList {
	fullPreview := make(dto.MatchPreviewList)

	// Range through each raw preview and format it.
	for _, r := range rawPreviews {

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

		items := make([]int, 0, 6)

		// Add non-null items to the array
		if r.Item0 != nil && *r.Item0 != 0 {
			items = append(items, *r.Item0)
		}
		if r.Item1 != nil && *r.Item1 != 0 {
			items = append(items, *r.Item1)
		}
		if r.Item2 != nil && *r.Item2 != 0 {
			items = append(items, *r.Item2)
		}
		if r.Item3 != nil && *r.Item3 != 0 {
			items = append(items, *r.Item3)
		}
		if r.Item4 != nil && *r.Item4 != 0 {
			items = append(items, *r.Item4)
		}
		if r.Item5 != nil && *r.Item5 != 0 {
			items = append(items, *r.Item5)
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

	return fullPreview
}

func handleCachedMatches(cachedMatches []dto.MatchPreview, matchesDto dto.MatchPreviewList) {
	for _, match := range cachedMatches {
		m := match
		matchesDto[match.Metadata.MatchId] = &m
	}
}
