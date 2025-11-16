CREATE VIEW all_events AS
    SELECT 
        match_id,
        timestamp,
        'feat_update' as event_type,
        NULL::bigint as participant_id,
        json_build_object(
            'feat_type', feat_type,
            'feat_value', feat_value,
            'team_id',team_id
        ) as data
    FROM event_feat_updates

    UNION ALL

    SELECT 
        match_id,
        timestamp,
        'item' as event_type,
        participant_id,
        json_build_object(
            'item_id', item_id,
            'after_id', after_id,
            'action', action
        ) as data
    FROM event_items

    UNION ALL

    SELECT 
        match_id,
        timestamp,
        'kill_struct' as event_type,
        participant_id,
        json_build_object(
            'building_type', building_type,
            'event_type', event_type,
            'lane_type', lane_type,
            'team_id',team_id,
            'tower_type', tower_type,
            'x',x,
            'y',y
        ) as data
    FROM event_kill_structs

    UNION ALL

    SELECT 
        match_id,
        timestamp,
        'level_up' as event_type,
        participant_id,
        json_build_object(
            'level', level
        ) as data
    FROM event_level_ups

    UNION ALL

    SELECT 
        match_id,
        timestamp,
        'monster_kill' as event_type,
        participant_id,
        json_build_object(
            'monster_type', monster_type,
            'team_id',killer_team ,
        	'x',x,
        	'y',y
        ) as data
    FROM event_monster_kills emk

    UNION ALL

    SELECT 
        match_id,
        timestamp,
        'player_kill' as event_type,
        participant_id,
        json_build_object(
        	'victim_participant_id',victim_participant_id ,
        	'x',x,
        	'y',y
        ) as data
    FROM event_player_kills 
    
    UNION ALL

    SELECT 
        match_id,
        timestamp,
        'skill_level_up' as event_type,
        participant_id,
        json_build_object(
            'level_up_type', level_up_type,
            'skill_slot', skill_slot
        ) as data
    FROM event_skill_level_ups

    UNION ALL

    SELECT
	    match_id,
	    timestamp,
	    'ward' AS event_type,
	    participant_id,
	    json_build_object(
            'event_type', event_type,
            'ward_type', ward_type
        ) AS DATA
    FROM event_wards;