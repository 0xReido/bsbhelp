package abbc

import (
	"image"

	wr "github.com/mroth/weightedrand"
)

type Trait struct {
	TraitType  string
	TraitValue string
	Weight     float64
}

type TraitData struct {
	TraitType        string
	TraitValue       string
	TraitProbability int
	TraitImage       image.Image
}

func GetRandomTrait(traits map[string]TraitData) (string, error) {
	choices := []wr.Choice{}
	for traitValue, traitData := range traits {
		choices = append(choices, wr.Choice{
			Item:   traitValue,
			Weight: uint(traitData.TraitProbability),
		})
	}
	chooser, err := wr.NewChooser(choices...)
	if err != nil {
		return "", err
	}
	result := chooser.Pick().(string)
	return result, nil
}
