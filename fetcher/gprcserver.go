package main

import (
	"context"
	"goleague/fetcher/assets"
	pb "goleague/pkg/grpc"
	"goleague/pkg/models/champion"
	"goleague/pkg/models/image"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Server definition.
type server struct {
	pb.UnimplementedAssetsServiceServer
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

func (s *server) RevalidateItemCache(context.Context, *pb.ItemId) (*pb.Item, error) {
	return nil, status.Errorf(codes.Unimplemented, "method RevalidateItemCache not implemented")
}
