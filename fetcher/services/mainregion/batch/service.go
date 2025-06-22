package batchservice

import (
	"errors"
	"goleague/fetcher/repositories"
	"goleague/pkg/database/models"
	"log"
	"sync"

	"gorm.io/gorm"
)

// Constants to improve maintainability.
const (
	EventTypeBuildingKill       = "BUILDING_KILL"
	EventTypeTurretPlateDestroy = "TURRET_PLATE_DESTROYED"
	EventTypeChampionKill       = "CHAMPION_KILL"
	EventTypeFeatUpdate         = "FEAT_UPDATE"
	EventTypeItemDestroyed      = "ITEM_DESTROYED"
	EventTypeItemPurchased      = "ITEM_PURCHASED"
	EventTypeItemSold           = "ITEM_SOLD"
	EventTypeLevelUp            = "LEVEL_UP"
	EventTypeSkillLevelUp       = "SKILL_LEVEL_UP"
	EventTypeWardKill           = "WARD_KILL"
	EventTypeWardPlaced         = "WARD_PLACED"
	EventTypeEliteMonsterKill   = "ELITE_MONSTER_KILL"
)

// BatchCollector is used for handling events insertion.
type BatchCollector struct {
	batches map[string][]any
	db      *gorm.DB
	mu      sync.RWMutex
}

// NewBatchCollector creates the batch collector.
func NewBatchCollector(db *gorm.DB) *BatchCollector {
	return &BatchCollector{
		batches: make(map[string][]any),
		db:      db,
	}
}

// Add a event to the collector or create the slice.
func (bc *BatchCollector) Add(eventType string, event any) {
	bc.mu.Lock()
	defer bc.mu.Unlock()

	// Create the slice if doesn't exist.
	if _, exists := bc.batches[eventType]; !exists {
		bc.batches[eventType] = make([]any, 0)
	}

	bc.batches[eventType] = append(bc.batches[eventType], event)
}

// ProcessBatches process the current stored event batches.
func (bc *BatchCollector) ProcessBatches() error {
	bc.mu.Lock()
	defer bc.mu.Unlock()

	if len(bc.batches) == 0 {
		return nil
	}

	var errs []error

	// Handle each event type and conversion to the respective model.
	for eventType, events := range bc.batches {
		switch eventType {
		case EventTypeBuildingKill, EventTypeTurretPlateDestroy:
			processBatchEvents[models.EventKillStruct](bc.db, events, eventType, &errs)

		case EventTypeChampionKill:
			processBatchEvents[models.EventPlayerKill](bc.db, events, eventType, &errs)

		case EventTypeFeatUpdate:
			processBatchEvents[models.EventFeatUpdate](bc.db, events, eventType, &errs)

		case EventTypeItemDestroyed, EventTypeItemPurchased, EventTypeItemSold:
			processBatchEvents[models.EventItem](bc.db, events, eventType, &errs)

		case EventTypeLevelUp:
			processBatchEvents[models.EventLevelUp](bc.db, events, eventType, &errs)

		case EventTypeSkillLevelUp:
			processBatchEvents[models.EventSkillLevelUp](bc.db, events, eventType, &errs)

		case EventTypeWardKill, EventTypeWardPlaced:
			processBatchEvents[models.EventWard](bc.db, events, eventType, &errs)

		case EventTypeEliteMonsterKill:
			processBatchEvents[models.EventMonsterKill](bc.db, events, eventType, &errs)
		}
	}

	// Return the errors if any was found.
	if len(errs) > 0 {
		return errors.Join(errs...)
	}

	return nil
}

// processBatchEvents returns the typed model list for inserting into the database.
func processBatchEvents[T any](db *gorm.DB, events []any, eventType string, errs *[]error) {
	modelList := make([]*T, 0, len(events))
	invalidCount := 0

	for _, event := range events {
		if model, ok := event.(*T); ok && model != nil {
			modelList = append(modelList, model)
		} else {
			invalidCount++
		}
	}

	if invalidCount > 0 {
		log.Printf("Warning: %d invalid events found for %s", invalidCount, eventType)
	}
	if err := repositories.CreateEventBatch(db, modelList); err != nil {
		log.Printf("Error inserting event %s: %v", eventType, err)
		*errs = append(*errs, err)
	}
}
