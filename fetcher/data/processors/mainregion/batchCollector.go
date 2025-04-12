package mainregion_processor

import (
	"errors"
	"goleague/pkg/database/models"
	"log"
	"sync"
)

// Batch collector used for handling events insertion
type batchCollector struct {
	batches map[string][]interface{}
	mu      sync.Mutex
}

// Create the batch collector.
func createBatchCollector() *batchCollector {
	return &batchCollector{
		batches: make(map[string][]interface{}),
	}
}

// Add a event to the collector or create the slice.
func (bc *batchCollector) Add(eventType string, event interface{}) {
	bc.mu.Lock()
	defer bc.mu.Unlock()

	// Create the slice if doesn't exist.
	if _, exists := bc.batches[eventType]; !exists {
		bc.batches[eventType] = make([]any, 0)
	}

	bc.batches[eventType] = append(bc.batches[eventType], event)
}

// Process the current stored event batches.
func (bc *batchCollector) processBatches(timelineService models.TimelineService) error {
	bc.mu.Lock()
	defer bc.mu.Unlock()

	var errs []error
	db := timelineService.GetDb()

	// Handle each event type and conversion to the respective model.
	// Could handle it directly by changing the batch structure to have pre defined model slices.
	// However that would not give flexibility.
	for eventType, events := range bc.batches {
		switch eventType {
		case "BUILDING_KILL", "TURRET_PLATE_DESTROYED":
			modelList := make([]*models.EventKillStruct, 0, len(events))
			for _, event := range events {
				if model, ok := event.(*models.EventKillStruct); ok && model != nil {
					modelList = append(modelList, model)
				}
			}
			if err := models.CreateEventBatch(db, modelList); err != nil {
				log.Printf("Error inserting struct kills: %v", err)
				errs = append(errs, err)
			}

		case "CHAMPION_KILL":
			modelList := make([]*models.EventPlayerKill, 0, len(events))
			for _, event := range events {
				if model, ok := event.(*models.EventPlayerKill); ok && model != nil {
					modelList = append(modelList, model)
				}
			}
			if err := models.CreateEventBatch(db, modelList); err != nil {
				log.Printf("Error inserting player kills: %v", err)
				errs = append(errs, err)
			}

		case "FEAT_UPDATE":
			modelList := make([]*models.EventFeatUpdate, 0, len(events))
			for _, event := range events {
				if model, ok := event.(*models.EventFeatUpdate); ok && model != nil {
					modelList = append(modelList, model)
				}
			}
			if err := models.CreateEventBatch(db, modelList); err != nil {
				log.Printf("Error inserting feat updates: %v", err)
				errs = append(errs, err)
			}

		case "ITEM_DESTROYED", "ITEM_PURCHASED", "ITEM_SOLD":
			modelList := make([]*models.EventItem, 0, len(events))
			for _, event := range events {
				if model, ok := event.(*models.EventItem); ok && model != nil {
					modelList = append(modelList, model)
				}
			}
			if err := models.CreateEventBatch(db, modelList); err != nil {
				log.Printf("Error inserting item events: %v", err)
				errs = append(errs, err)
			}

		case "LEVEL_UP":
			modelList := make([]*models.EventLevelUp, 0, len(events))
			for _, event := range events {
				if model, ok := event.(*models.EventLevelUp); ok && model != nil {
					modelList = append(modelList, model)
				}
			}
			if err := models.CreateEventBatch(db, modelList); err != nil {
				log.Printf("Error inserting level up event: %v", err)
				errs = append(errs, err)
			}

		case "SKILL_LEVEL_UP":
			modelList := make([]*models.EventSkillLevelUp, 0, len(events))
			for _, event := range events {
				if model, ok := event.(*models.EventSkillLevelUp); ok && model != nil {
					modelList = append(modelList, model)
				}
			}
			if err := models.CreateEventBatch(db, modelList); err != nil {
				log.Printf("Error inserting skill level up event:%v", err)
				errs = append(errs, err)
			}

		case "WARD_KILL", "WARD_PLACED":
			modelList := make([]*models.EventWard, 0, len(events))
			for _, event := range events {
				if model, ok := event.(*models.EventWard); ok && model != nil {
					modelList = append(modelList, model)
				}
			}
			if err := models.CreateEventBatch(db, modelList); err != nil {
				log.Printf("Error inserting ward event:%v", err)
				errs = append(errs, err)
			}

		case "ELITE_MONSTER_KILL":
			modelList := make([]*models.EventMonsterKill, 0, len(events))
			for _, event := range events {
				if model, ok := event.(*models.EventMonsterKill); ok && model != nil {
					modelList = append(modelList, model)
				}
			}
			if err := models.CreateEventBatch(db, modelList); err != nil {
				log.Printf("Error inserting ward event:%v", err)
				errs = append(errs, err)
			}
		}
	}

	// Return the errors if any was found.
	if len(errs) > 0 {
		return errors.Join(errs...)
	}

	return nil
}
