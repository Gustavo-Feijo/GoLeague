CREATE TABLE
    player_fetch_priorities (
        player_id BIGINT PRIMARY KEY,
        fetch_priority INTEGER NOT NULL,
        last_calculated TIMESTAMP NOT NULL DEFAULT NOW (),
        region VARCHAR(10) NOT NULL
    );

DROP INDEX IF EXISTS index_player_queue_time;

CREATE INDEX IF NOT EXISTS idx_rating_entries_player_fetch ON rating_entries (player_id, fetch_time DESC) INCLUDE (numeric_score);

CREATE INDEX IF NOT EXISTS idx_match_stats_player ON match_stats (player_id, match_id);