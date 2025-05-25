package main

import (
	"context"
	"goleague/fetcher/assets"
	regionmanager "goleague/fetcher/regionmanager"
	pb "goleague/pkg/grpc"
	"goleague/pkg/models/champion"
	"goleague/pkg/models/image"
	"goleague/pkg/models/item"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Server definition.
type server struct {
	pb.UnimplementedAssetsServiceServer
	regionManager *regionmanager.RegionManager
}

// Util function to get the PB definition for the gold.
func goldToPB(gold item.Gold) pb.Gold {
	return pb.Gold{
		Base:        int32(gold.Base),
		Total:       int32(gold.Total),
		Sell:        int32(gold.Sell),
		Purchasable: bool(gold.Purchasable),
	}
}

// Util function to get the PB definition from the image.
func imageToPB(image image.Image) pb.Image {
	return pb.Image{
		FullImage: image.Full,
		Sprite:    image.Sprite,
		X:         int32(image.X),
		Y:         int32(image.Y),
		W:         int32(image.W),
		H:         int32(image.H),
	}
}

// Util function to get the PB definition from the spell.
func spellToPB(spell champion.Spell) *pb.Spell {
	spellImage := imageToPB(spell.Image)
	return &pb.Spell{
		Id:          spell.ID,
		Name:        spell.Name,
		Description: spell.Description,
		Cooldown:    &spell.Cooldown,
		Cost:        &spell.Cost,
		Image:       &spellImage,
		ChampionId:  spell.ChampionID,
	}
}

// Revalidate the entire champion cache and return a champion if a valid id is provided.
func (s *server) RevalidateChampionCache(context context.Context, srv *pb.ChampionId) (*pb.Champion, error) {
	// Get the champion.
	champion, err := assets.RevalidateChampionCache("en_US", srv.Id)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "can't get the champion from the cache: %v", err)
	}

	// Get the protobuff definition for the image.
	championImage := imageToPB(champion.Image)

	// Get the spells.
	championSpells := make([]*pb.Spell, len(champion.Spells))
	for i, spell := range champion.Spells {
		championSpells[i] = spellToPB(spell)
	}

	championPassive := spellToPB(champion.Passive)

	return &pb.Champion{
		Id:      champion.ID,
		Key:     champion.NameKey,
		Name:    champion.Name,
		Title:   champion.Title,
		Image:   &championImage,
		Spells:  championSpells,
		Passive: championPassive,
	}, nil
}

// Revalidate a item cache.
func (s *server) RevalidateItemCache(ctx context.Context, srv *pb.ItemId) (*pb.Item, error) {
	// Get the champion.
	item, err := assets.RevalidateItemCache("en_US", srv.Id)
	if err != nil || item == nil {
		return nil, status.Errorf(codes.Internal, "can't get the champion from the cache: %v", err)
	}

	// Get the protobuff definition for the image.
	itemImage := imageToPB(item.Image)
	itemGold := goldToPB(item.Gold)

	return &pb.Item{
		Id:          item.ID,
		Name:        item.Name,
		Description: item.Description,
		Plaintext:   item.Plaintext,
		Image:       &itemImage,
		Gold:        &itemGold,
	}, nil
}
