package services

import apirepositories "goleague/api/repositories"

type TierlistService interface{}

func NewTierlistService(repo apirepositories.TierlistRepository) TierlistService {
	return nil
}
