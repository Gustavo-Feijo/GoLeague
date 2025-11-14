DROP TRIGGER IF EXISTS update_fetch_match_rating_insert ON rating_entries;

DROP FUNCTION IF EXISTS update_unfetched_match;

-- Drop tables in dependency order
DROP TABLE IF EXISTS participant_frames;

DROP TABLE IF EXISTS event_wards;

DROP TABLE IF EXISTS event_skill_level_ups;

DROP TABLE IF EXISTS event_player_kills;

DROP TABLE IF EXISTS event_monster_kills;

DROP TABLE IF EXISTS event_level_ups;

DROP TABLE IF EXISTS event_kill_structs;

DROP TABLE IF EXISTS event_items;

DROP TABLE IF EXISTS event_feat_updates;

DROP TABLE IF EXISTS match_bans;

DROP TABLE IF EXISTS cache_backups;

DROP TABLE IF EXISTS match_stats;

DROP TABLE IF EXISTS rating_entries;

DROP TABLE IF EXISTS match_infos;

DROP TABLE IF EXISTS player_infos;

-- Drop custom types if they exist
DROP TYPE IF EXISTS queue_type;

DROP TYPE IF EXISTS tier_type;

DROP TYPE IF EXISTS rank_type;