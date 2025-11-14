-- Create enum types if they do not exist.
DO $$ 
		BEGIN
		    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'queue_type') THEN
		        CREATE TYPE queue_type AS ENUM ('RANKED_SOLO_5x5', 'RANKED_FLEX_SR');
		    END IF;

		    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'tier_type') THEN
		        CREATE TYPE tier_type AS ENUM ('IRON', 'BRONZE', 'SILVER', 'GOLD', 'PLATINUM', 'EMERALD', 'DIAMOND', 'MASTER', 'GRANDMASTER', 'CHALLENGER');
		    END IF;

		    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'rank_type') THEN
		        CREATE TYPE rank_type AS ENUM ('IV', 'III', 'II', 'I');
		END IF;
END $$;

-- Player information table.
-- Fetcher on the fetcher, stores basic player info.
CREATE TABLE player_infos (
	id bigserial NOT NULL,
	profile_icon int8 NULL,
	puuid bpchar(78) NULL,
	riot_id_game_name varchar(100) NULL,
	riot_id_tagline varchar(5) NULL,
	summoner_level int8 NULL,
	region varchar(5) NULL,
	unfetched_match bool DEFAULT true NULL,
	last_match_fetch timestamptz DEFAULT CURRENT_TIMESTAMP NULL,
	updated_at timestamptz NULL,
	CONSTRAINT player_infos_pkey PRIMARY KEY (id)
);

-- Indexes for player_infos table.
CREATE INDEX idx_name_tag ON player_infos USING btree (riot_id_game_name, riot_id_tagline);
CREATE INDEX idx_player_infos_last_match_fetch ON player_infos USING btree (last_match_fetch);
CREATE INDEX idx_player_infos_puuid ON player_infos USING btree (puuid);
CREATE UNIQUE INDEX idx_player_region ON player_infos USING btree (puuid, region);
CREATE INDEX idx_player_search_all ON player_infos USING btree (region, riot_id_game_name text_pattern_ops, riot_id_tagline text_pattern_ops) WHERE ((riot_id_game_name)::text <> ''::text);
CREATE INDEX idx_unfetched_region ON player_infos USING btree (region, unfetched_match);

CREATE TABLE rating_entries (
	id bigserial NOT NULL,
	player_id int8 NULL,
	queue "queue_type" NULL,
	tier "tier_type" NULL,
	"rank" "rank_type" NULL,
	numeric_score int8 NULL,
	league_points int8 NULL,
	wins int8 NULL,
	losses int8 NULL,
	region varchar(5) NULL,
	fetch_time timestamptz NULL,
	CONSTRAINT rating_entries_pkey PRIMARY KEY (id),
	CONSTRAINT fk_rating_entries_player FOREIGN KEY (player_id) REFERENCES player_infos(id)
);
CREATE INDEX idx_player_queue_time ON rating_entries USING btree (queue);
CREATE INDEX index_player_queue_time ON rating_entries USING btree (player_id, fetch_time);

-- Function to update unfetched_match flag on player_infos after inserting a new rating entry.
-- Rating entries only are created when something on the ratings changed, mostly due to new matches.
CREATE OR REPLACE FUNCTION update_unfetched_match()
      RETURNS TRIGGER AS $$
      BEGIN
        UPDATE player_infos
        SET unfetched_match = true 
        WHERE id = NEW.player_id;

        RETURN NEW;
      END;
      $$ LANGUAGE plpgsql;
      
CREATE TRIGGER update_fetch_match_rating_insert
	AFTER INSERT ON rating_entries
	FOR EACH ROW
		EXECUTE FUNCTION update_unfetched_match();


-- Match information table.
-- Stores match metadata.
CREATE TABLE match_infos (
	id bigserial NOT NULL,
	game_version varchar(20) NULL,
	match_id varchar(20) NULL,
	match_start timestamptz NULL,
	match_duration int8 NULL,
	match_winner int8 NULL,
	match_surrender bool NULL,
	match_remake bool NULL,
	average_rating numeric NULL,
	frame_interval int8 NULL,
	fully_fetched bool NULL,
	queue_id int8 NULL,
	created_at timestamptz NULL,
	CONSTRAINT match_infos_pkey PRIMARY KEY (id)
);

-- Indexes for match_infos table.
CREATE INDEX idx_match_infos_average_rating ON match_infos USING btree (average_rating);
CREATE UNIQUE INDEX idx_match_infos_match_id ON match_infos USING btree (match_id);
CREATE INDEX idx_match_infos_queue_id ON match_infos USING btree (queue_id);

-- Match statistics table.
-- Stores per-player statistics for each match.
CREATE TABLE match_stats (
	id bigserial NOT NULL,
	match_id int8 NOT NULL,
	player_id int8 NOT NULL,
	assists int8 NULL,
	all_in_pings int8 NULL,
	assist_me_ping int8 NULL,
	baron_kills int8 NULL,
	basic_pings int8 NULL,
	champion_level int8 NULL,
	champion_id int8 NULL,
	ability_uses int8 NULL,
	control_wards_placed int8 NULL,
	skillshots_dodged int8 NULL,
	skillshots_hit int8 NULL,
	command_pings int8 NULL,
	danger_pings int8 NULL,
	deaths int8 NULL,
	enemy_missing_pings int8 NULL,
	enemy_vision_pings int8 NULL,
	get_back_pings int8 NULL,
	gold_earned int8 NULL,
	gold_spent int8 NULL,
	hold_pings int8 NULL,
	item0 int8 NULL,
	item1 int8 NULL,
	item2 int8 NULL,
	item3 int8 NULL,
	item4 int8 NULL,
	item5 int8 NULL,
	kills int8 NULL,
	magic_damage_dealt_to_champions int8 NULL,
	magic_damage_taken int8 NULL,
	need_vision_pings int8 NULL,
	neutral_minions_killed int8 NULL,
	on_my_way_pings int8 NULL,
	participant_id int8 NULL,
	physical_damage_dealt_to_champions int8 NULL,
	physical_damage_taken int8 NULL,
	push_pings int8 NULL,
	retreat_pings int8 NULL,
	longest_time_spent_living int8 NULL,
	magic_damage_dealt int8 NULL,
	team_id int8 NULL,
	team_position text NULL,
	time_c_cing_others int8 NULL,
	total_damage_dealt_to_champions int8 NULL,
	total_minions_killed int8 NULL,
	total_time_spent_dead int8 NULL,
	true_damage_dealt_to_champions int8 NULL,
	vision_cleared_pings int8 NULL,
	vision_score int8 NULL,
	wards_killed int8 NULL,
	wards_placed int8 NULL,
	win bool NULL,
	CONSTRAINT match_stats_pkey PRIMARY KEY (id),
	CONSTRAINT fk_match_stats_match FOREIGN KEY (match_id) REFERENCES match_infos(id),
	CONSTRAINT fk_match_stats_player FOREIGN KEY (player_id) REFERENCES player_infos(id)
);

-- Match statistics indexes.
CREATE UNIQUE INDEX idx_match_player ON match_stats USING btree (match_id, player_id);

-- Fallback table if unable to revalidate cache with hot data. 
CREATE TABLE cache_backups (
	cache_key text NOT NULL,
	cache_value jsonb NULL,
	CONSTRAINT cache_backups_pkey PRIMARY KEY (cache_key)
);

-- Match bans table.
CREATE TABLE match_bans (
	match_id int8 NOT NULL,
	pick_turn int8 NOT NULL,
	champion_id int8 NULL,
	CONSTRAINT match_bans_pkey PRIMARY KEY (match_id, pick_turn)
);


-- Event table for feats of strenght updates.
CREATE TABLE event_feat_updates (
	match_id int8 NULL,
	"timestamp" int8 NULL,
	feat_type int8 NULL,
	feat_value int8 NULL,
	team_id int8 NULL,
	CONSTRAINT fk_event_feat_updates_match_info FOREIGN KEY (match_id) REFERENCES match_infos(id)
);
CREATE INDEX idx_event_feat_updates_match_id ON event_feat_updates USING btree (match_id);

-- Event table for items.
CREATE TABLE event_items (
	match_id int8 NOT NULL,
	participant_id int8 NULL,
	"timestamp" int8 NOT NULL,
	item_id int8 NULL,
	after_id int8 NULL,
	"action" text NULL,
	CONSTRAINT fk_event_items_match_info FOREIGN KEY (match_id) REFERENCES match_infos(id)
);
CREATE INDEX idx_event_items_timestamp ON event_items USING btree ("timestamp");
CREATE INDEX idx_match_participant ON event_items USING btree (match_id, participant_id);

-- Event table for structure kills (Tower, Inhibitor, Nexus, Plates).
CREATE TABLE event_kill_structs (
	match_id int8 NOT NULL,
	participant_id int8 NULL,
	"timestamp" int8 NOT NULL,
	team_id int8 NULL,
	event_type text NULL,
	building_type text NULL,
	lane_type text NULL,
	tower_type text NULL,
	x int8 NULL,
	y int8 NULL,
	CONSTRAINT fk_event_kill_structs_match_info FOREIGN KEY (match_id) REFERENCES match_infos(id)
);
CREATE INDEX idx_event_kill_structs_timestamp ON event_kill_structs USING btree ("timestamp");

-- Event table for champion level up.
CREATE TABLE event_level_ups (
	match_id int8 NOT NULL,
	participant_id int8 NULL,
	"timestamp" int8 NOT NULL,
	"level" int8 NULL,
	CONSTRAINT fk_event_level_ups_match_info FOREIGN KEY (match_id) REFERENCES match_infos(id)
);
CREATE INDEX idx_event_level_ups_timestamp ON event_level_ups USING btree ("timestamp");

-- Event table for monster kills (Dragons, Rift Herald, Atahkan, Baron).
CREATE TABLE event_monster_kills (
	match_id int8 NOT NULL,
	participant_id int8 NULL,
	"timestamp" int8 NOT NULL,
	killer_team int8 NULL,
	monster_type text NULL,
	x int8 NULL,
	y int8 NULL,
	CONSTRAINT fk_event_monster_kills_match_info FOREIGN KEY (match_id) REFERENCES match_infos(id)
);
CREATE INDEX idx_event_monster_kills_timestamp ON event_monster_kills USING btree ("timestamp");

-- Event table for player kills.
CREATE TABLE event_player_kills (
	match_id int8 NOT NULL,
	participant_id int8 NULL,
	"timestamp" int8 NOT NULL,
	victim_participant_id int8 NULL,
	x int8 NULL,
	y int8 NULL,
	CONSTRAINT fk_event_player_kills_match_info FOREIGN KEY (match_id) REFERENCES match_infos(id)
);
CREATE INDEX idx_event_player_kills_timestamp ON event_player_kills USING btree ("timestamp");

-- Event table for skill level ups.
CREATE TABLE event_skill_level_ups (
	match_id int8 NOT NULL,
	participant_id int8 NULL,
	"timestamp" int8 NOT NULL,
	level_up_type varchar(30) NULL,
	skill_slot int8 NULL,
	CONSTRAINT fk_event_skill_level_ups_match_info FOREIGN KEY (match_id) REFERENCES match_infos(id)
);
CREATE INDEX idx_event_skill_level_ups_timestamp ON event_skill_level_ups USING btree ("timestamp");

-- Event table for ward placements and destructions.
CREATE TABLE event_wards (
	match_id int8 NOT NULL,
	participant_id int8 NULL,
	"timestamp" int8 NOT NULL,
	event_type text NOT NULL,
	ward_type text NULL,
	CONSTRAINT fk_event_wards_match_info FOREIGN KEY (match_id) REFERENCES match_infos(id)
);
CREATE INDEX idx_event_wards_timestamp ON event_wards USING btree ("timestamp");

-- Participant frames table for storing per-frame data for each participant in a match.
CREATE TABLE participant_frames (
	match_stat_id int8 NOT NULL,
	frame_index int8 NOT NULL,
	current_gold int8 NULL,
	magic_damage_done int8 NULL,
	magic_damage_done_to_champions int8 NULL,
	magic_damage_taken int8 NULL,
	physical_damage_done int8 NULL,
	physical_damage_done_to_champions int8 NULL,
	physical_damage_taken int8 NULL,
	total_damage_done int8 NULL,
	total_damage_done_to_champions int8 NULL,
	total_damage_taken int8 NULL,
	true_damage_done int8 NULL,
	true_damage_done_to_champions int8 NULL,
	true_damage_taken int8 NULL,
	jungle_minions_killed int8 NULL,
	"level" int8 NULL,
	minions_killed int8 NULL,
	participant_id int8 NULL,
	total_gold int8 NULL,
	xp int8 NULL,
	CONSTRAINT participant_frames_pkey PRIMARY KEY (match_stat_id, frame_index),
	CONSTRAINT fk_participant_frames_match_stat FOREIGN KEY (match_stat_id) REFERENCES match_stats(id)
);

