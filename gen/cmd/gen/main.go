package main

import (
	"fmt"
	"image"
	"log"
	"math/rand"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"text/tabwriter"
	"time"

	"image/color"
	"image/draw"
	"image/png"

	"github.com/goccy/go-yaml"
	"github.com/mroth/weightedrand"
	"github.com/oliamb/cutter"
	"github.com/schollz/progressbar/v3"
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

func (g *Generator) GetRandomTrait(traitType string) (string, error) {
	result := g.TraitChoosers[traitType].Pick().(string)
	return result, nil
}

type Datum struct {
	Name   string `yaml:"name"`
	File   string `yaml:"file"`
	Chance int    `yaml:"chance"`
	Active bool   `yaml:"active"`
}

type YamlTraitData struct {
	Values []Datum `yaml:"values"`
}

type Data struct {
	Traits map[string]YamlTraitData `yaml:"traits"`
}

func GetTraits(wantedTraitType string) (map[string]TraitData, error) {

	dataFile, err := os.Open(filepath.FromSlash("./abbc.yml"))
	if err != nil {
		return nil, err
	}

	d := &Data{}
	decoder := yaml.NewDecoder(dataFile)
	err = decoder.Decode(d)
	if err != nil {
		return nil, err
	}

	totalProbability := 0
	traits := []TraitData{}

	for traitType, traitData := range d.Traits {
		// fmt.Println(traitType, traitData)
		if traitType != wantedTraitType {
			continue
		}
		for _, traitDatum := range traitData.Values {
			// fmt.Println(traitType, traitDatum.Name)
			img, err := GetImage(traitDatum.File)
			if err != nil {
				return nil, err
			}
			totalProbability += traitDatum.Chance
			traits = append(traits, TraitData{
				TraitValue:       traitDatum.Name,
				TraitProbability: traitDatum.Chance,
				TraitImage:       img,
			})
		}
	}

	traitMap := make(map[string]TraitData)
	fmt.Println(wantedTraitType)
	w := tabwriter.NewWriter(os.Stdout, 1, 1, 1, ' ', 0)

	for _, trait := range traits {
		fmt.Fprintf(w, "%s\t%s\t%.01f%%\n", wantedTraitType, trait.TraitValue, float64(trait.TraitProbability)/10)
		traitMap[trait.TraitValue] = trait
	}

	fmt.Fprintf(w, "Total\tPercentage:\t%.01f%%\n", float64(totalProbability)/10)

	if totalProbability < 1000 {
		traitMap["__NONE__"] = TraitData{
			TraitType:        wantedTraitType,
			TraitValue:       "__NONE__",
			TraitProbability: 1000 - totalProbability,
		}
	}

	w.Flush()
	fmt.Println()
	return traitMap, nil
}

// func (g *Generator) GetRandomTrait(traitType string) (string, error) {
// 	traitMap := g.TraitMaps[traitType]
// 	traitValue, err := GetRandomTrait(traitMap)
// 	if err != nil {
// 		return "", err
// 	}
// 	return traitValue, nil
// }

type Generator struct {
	TraitChoosers map[string]*weightedrand.Chooser
	TraitMaps     map[string]map[string]TraitData
	SpecialImages map[string]image.Image
}

func newGenerator(imagesPath string) (*Generator, error) {
	g := &Generator{
		TraitMaps:     make(map[string]map[string]TraitData),
		TraitChoosers: make(map[string]*weightedrand.Chooser),
		SpecialImages: make(map[string]image.Image),
	}
	traits := []string{"Background", "Fur", "Clothes", "Eyes", "Mouth", "Head", "Jewelry"}
	for _, trait := range traits {
		traitMap, err := GetTraits(trait)
		if err != nil {
			return nil, err
		}
		g.TraitMaps[trait] = traitMap

		keys := make([]string, 0, len(traitMap))
		for k := range traitMap {
			keys = append(keys, k)
		}
		sort.Strings(keys)

		choices := []weightedrand.Choice{}
		for _, k := range keys {
			// fmt.Println(k)
			choices = append(choices, weightedrand.Choice{
				Item:   k,
				Weight: uint(traitMap[k].TraitProbability),
			})
		}

		chooser, err := weightedrand.NewChooser(choices...)
		if err != nil {
			return nil, err
		}

		g.TraitChoosers[trait] = chooser
	}

	imagePath := fmt.Sprintf("%s/Special/Grin Left.png", imagesPath)
	img, err := GetImage(imagePath)
	if err != nil {
		return nil, fmt.Errorf("GetImage(%s) error: %w", imagePath, err)
	}
	g.SpecialImages["Grin Left"] = img

	imagePath = fmt.Sprintf("%s/Special/Bandana Left.png", imagesPath)
	img, err = GetImage(imagePath)
	if err != nil {
		return nil, fmt.Errorf("GetImage(%s) error: %w", imagePath, err)
	}
	g.SpecialImages["Bandana Left"] = img

	imagePath = fmt.Sprintf("%s/Special/BTC Ballers Top.png", imagesPath)
	img, err = GetImage(imagePath)
	if err != nil {
		return nil, fmt.Errorf("GetImage(%s) error: %w", imagePath, err)
	}
	g.SpecialImages["BTC Ballers Top"] = img

	imagePath = fmt.Sprintf("%s/Special/Trooper Hat Right.png", imagesPath)
	img, err = GetImage(imagePath)
	if err != nil {
		return nil, fmt.Errorf("GetImage(%s) error: %w", imagePath, err)
	}
	g.SpecialImages["Trooper Hat Right"] = img

	imagePath = fmt.Sprintf("%s/Special/Joint Smoke.png", imagesPath)
	img, err = GetImage(imagePath)
	if err != nil {
		return nil, fmt.Errorf("GetImage(%s) error: %w", imagePath, err)
	}
	g.SpecialImages["Joint Smoke"] = img

	imagePath = fmt.Sprintf("%s/Special/Sport shades cut bottom.png", imagesPath)
	img, err = GetImage(imagePath)
	if err != nil {
		return nil, fmt.Errorf("GetImage(%s) error: %w", imagePath, err)
	}
	g.SpecialImages["Sport shades"] = img

	imagePath = fmt.Sprintf("%s/Special/Flame shades cut bottom.png", imagesPath)
	img, err = GetImage(imagePath)
	if err != nil {
		return nil, fmt.Errorf("GetImage(%s) error: %w", imagePath, err)
	}
	g.SpecialImages["Flame shades"] = img

	imagePath = fmt.Sprintf("%s/Special/Plasma vision cut bottom.png", imagesPath)
	img, err = GetImage(imagePath)
	if err != nil {
		return nil, fmt.Errorf("GetImage(%s) error: %w", imagePath, err)
	}
	g.SpecialImages["Plasma vision cut bottom"] = img

	imagePath = fmt.Sprintf("%s/Special/Plasma vision.png", imagesPath)
	img, err = GetImage(imagePath)
	if err != nil {
		return nil, fmt.Errorf("GetImage(%s) error: %w", imagePath, err)
	}
	g.SpecialImages["Plasma vision"] = img

	imagePath = fmt.Sprintf("%s/Special/Bitcoin ballers cut bottom.png", imagesPath)
	img, err = GetImage(imagePath)
	if err != nil {
		return nil, fmt.Errorf("GetImage(%s) error: %w", imagePath, err)
	}
	g.SpecialImages["Bitcoin ballers"] = img

	imagePath = fmt.Sprintf("%s/Special/Nostril.png", imagesPath)
	img, err = GetImage(imagePath)
	if err != nil {
		return nil, fmt.Errorf("GetImage(%s) error: %w", imagePath, err)
	}
	g.SpecialImages["Nostril"] = img

	imagePath = fmt.Sprintf("%s/Special/Plasma vision bottom.png", imagesPath)
	img, err = GetImage(imagePath)
	if err != nil {
		return nil, fmt.Errorf("GetImage(%s) error: %w", imagePath, err)
	}
	g.SpecialImages["Plasma vision bottom"] = img

	imagePath = fmt.Sprintf("%s/Special/Laser.png", imagesPath)
	img, err = GetImage(imagePath)
	if err != nil {
		return nil, fmt.Errorf("GetImage(%s) error: %w", imagePath, err)
	}
	g.SpecialImages["Laser"] = img

	imagePath = fmt.Sprintf("%s/Special/Oversized Goggle Line.png", imagesPath)
	img, err = GetImage(imagePath)
	if err != nil {
		return nil, fmt.Errorf("GetImage(%s) error: %w", imagePath, err)
	}
	g.SpecialImages["Oversized Goggle Line"] = img

	imagePath = fmt.Sprintf("%s/Special/Trooper Hat Bandana.png", imagesPath)
	img, err = GetImage(imagePath)
	if err != nil {
		return nil, fmt.Errorf("GetImage(%s) error: %w", imagePath, err)
	}
	g.SpecialImages["Trooper Hat Bandana"] = img

	imagePath = fmt.Sprintf("%s/Special/bored-puffer-mouth.png", imagesPath)
	img, err = GetImage(imagePath)
	if err != nil {
		return nil, fmt.Errorf("GetImage(%s) error: %w", imagePath, err)
	}
	g.SpecialImages["bored-puffer-mouth"] = img

	imagePath = fmt.Sprintf("%s/Special/bored-unshaven-puffer-mouth.png", imagesPath)
	img, err = GetImage(imagePath)
	if err != nil {
		return nil, fmt.Errorf("GetImage(%s) error: %w", imagePath, err)
	}
	g.SpecialImages["bored-unshaven-puffer-mouth"] = img

	imagePath = fmt.Sprintf("%s/Special/phenome-puffer-mouth.png", imagesPath)
	img, err = GetImage(imagePath)
	if err != nil {
		return nil, fmt.Errorf("GetImage(%s) error: %w", imagePath, err)
	}
	g.SpecialImages["phenome-puffer-mouth"] = img

	imagePath = fmt.Sprintf("%s/Special/tongue-puffer-mouth.png", imagesPath)
	img, err = GetImage(imagePath)
	if err != nil {
		return nil, fmt.Errorf("GetImage(%s) error: %w", imagePath, err)
	}
	g.SpecialImages["tongue-puffer-mouth"] = img

	imagePath = fmt.Sprintf("%s/Special/grin-puffer-mouth.png", imagesPath)
	img, err = GetImage(imagePath)
	if err != nil {
		return nil, fmt.Errorf("GetImage(%s) error: %w", imagePath, err)
	}
	g.SpecialImages["grin-puffer-mouth"] = img

	imagePath = fmt.Sprintf("%s/Special/small-grin-puffer-mouth.png", imagesPath)
	img, err = GetImage(imagePath)
	if err != nil {
		return nil, fmt.Errorf("GetImage(%s) error: %w", imagePath, err)
	}
	g.SpecialImages["small-grin-puffer-mouth"] = img

	imagePath = fmt.Sprintf("%s/Special/discomfort-puffer-mouth.png", imagesPath)
	img, err = GetImage(imagePath)
	if err != nil {
		return nil, fmt.Errorf("GetImage(%s) error: %w", imagePath, err)
	}
	g.SpecialImages["discomfort-puffer-mouth"] = img

	imagePath = fmt.Sprintf("%s/Special/rose-puffer-mouth.png", imagesPath)
	img, err = GetImage(imagePath)
	if err != nil {
		return nil, fmt.Errorf("GetImage(%s) error: %w", imagePath, err)
	}
	g.SpecialImages["rose-puffer-mouth"] = img

	imagePath = fmt.Sprintf("%s/Special/beanie-oversized-eyes.png", imagesPath)
	img, err = GetImage(imagePath)
	if err != nil {
		return nil, fmt.Errorf("GetImage(%s) error: %w", imagePath, err)
	}
	g.SpecialImages["beanie-oversized-eyes"] = img

	imagePath = fmt.Sprintf("%s/Special/beanie-oversized-head.png", imagesPath)
	img, err = GetImage(imagePath)
	if err != nil {
		return nil, fmt.Errorf("GetImage(%s) error: %w", imagePath, err)
	}
	g.SpecialImages["beanie-oversized-head"] = img

	imagePath = fmt.Sprintf("%s/Special/backwards-bandana-thick-frame-glasses.png", imagesPath)
	img, err = GetImage(imagePath)
	if err != nil {
		return nil, fmt.Errorf("GetImage(%s) error: %w", imagePath, err)
	}
	g.SpecialImages["backwards-bandana-thick-frame-glasses"] = img

	imagePath = fmt.Sprintf("%s/Special/goggles-robot-head.png", imagesPath)
	img, err = GetImage(imagePath)
	if err != nil {
		return nil, fmt.Errorf("GetImage(%s) error: %w", imagePath, err)
	}
	g.SpecialImages["goggles-robot-head"] = img

	imagePath = fmt.Sprintf("%s/Special/blonde-braids-oversized-glasses.png", imagesPath)
	img, err = GetImage(imagePath)
	if err != nil {
		return nil, fmt.Errorf("GetImage(%s) error: %w", imagePath, err)
	}
	g.SpecialImages["blonde-braids-oversized-glasses"] = img

	return g, nil
}

type Metadata struct {
	TokenID int
	Traits  []struct {
		TraitType  string
		TraitValue string
	}
}

	func (g *Generator) GenerateMetadata(tokenID int) (*Metadata, error) {
	traits := []string{"Background", "Fur", "Clothes", "Eyes", "Head", "Mouth", "Jewelry"}
	traitValues := []string{}
	for _, trait := range traits {
		traitValue, err := g.GetRandomTrait(trait)
		if err != nil {
			return nil, err
		}

		traitValues = append(traitValues, traitValue)
	}

	m := &Metadata{
		TokenID: tokenID,
	}
	for i, trait := range traits {
		m.Traits = append(m.Traits, struct {
			TraitType  string
			TraitValue string
		}{
			TraitType:  trait,
			TraitValue: traitValues[i],
		})
	}
	return m, nil
}


func GetImage(path string) (image.Image, error) {
	imageFile, err := os.Open(filepath.FromSlash(path))
	if err != nil {
		return nil, err
	}
	defer imageFile.Close()
	img, _, err := image.Decode(imageFile)
	return img, err
}

type colorChanger interface {
	Set(x, y int, c color.Color)
	At(x, y int) color.Color
}

func (g *Generator) GenerateImage(m *Metadata) error {
	r := image.Rectangle{image.Point{0, 0}, image.Point{1262, 1262}}
	newImage := image.NewRGBA(r)

	isTrooper := false
	isGrin := false
	isSmallGrin := false
	isDiscomfort := false
	isBored := false
	isBoredUnshaven := false
	isPhenome := false
	isHelmet := false
	isJoint := false
	isBackwardHat := false
	isPlasmaVision := false
	isSportShades := false

	isLasers := false

	isBandanaMouth := false
	isDumbfounded := false
	isRose := false
	isTongue := false

	isSakura := false
	isFlameShades := false

	isBeanie := false
	isKnitBeanie := false
	isSweatband := false
	isBTCBallers := false
	isBackwardBandana := false
	isBandanaHead := false
	isPanelHat := false
	isBlondeBraids := false

	isGeometricShades := false
	isTheDonShades := false
	isThickFrameShades := false
	isThinShades := false
	isOversized := false
	isGlasses := false
	isBigGlasses := false
	isRobot := false

	isDreadlocks := false
	isMessyHair := false
	isTwoToneBraids := false

	isZippedPuffer := false

	isGoggles := false
	for _, trait := range m.Traits {

		if trait.TraitValue == "ETH Lasers" {
			isLasers = true
		}

		if trait.TraitValue == "Robot" {
			isRobot = true
		}

		if strings.Contains(trait.TraitValue, "Goggles") {
			isGoggles = true
		}

		if trait.TraitValue == "Two Tone Braids" {
			isTwoToneBraids = true
		}
		if trait.TraitValue == "Dreadlocks" {
			isDreadlocks = true
		}
		if trait.TraitValue == "Messy Hair" {
			isMessyHair = true
		}
		if trait.TraitValue == "Blonde Braids" {
			isBlondeBraids = true
		}

		if trait.TraitValue == "Trooper Hat" {
			isTrooper = true
		}
		if strings.Contains(trait.TraitValue, "Grin") && trait.TraitValue != "Small Grin" {
			isGrin = true
		} else if trait.TraitValue == "Small Grin" {
			isSmallGrin = true
		}

		if strings.Contains(trait.TraitValue, "Bored") && trait.TraitValue != "Bored Unshaven" && trait.TraitType == "Mouth" {
			isBored = true
		} else if trait.TraitValue == "Bored Unshaven" && trait.TraitType == "Mouth" {
			isBoredUnshaven = true
		}

		if trait.TraitValue == "Phoneme Vuh" {
			isPhenome = true
		}

		if trait.TraitValue == "Discomfort" {
			isDiscomfort = true
		}

		if trait.TraitValue == "Army Helmet" {
			isHelmet = true
		}
		if trait.TraitValue == "Bored Joint" {
			isJoint = true
		}
		if trait.TraitValue == "Backwards Hat" {
			isBackwardHat = true
		}
		if trait.TraitValue == "Plasma Vision" {
			isPlasmaVision = true
		}
		if trait.TraitValue == "Flame Shades" {
			isFlameShades = true
		}
		if trait.TraitValue == "Sport Shades" {
			isSportShades = true
		}
		if trait.TraitValue == "The Don Shades" {
			isTheDonShades = true
		}
		if trait.TraitValue == "Thick Frames" {
			isThickFrameShades = true
		}
		if trait.TraitValue == "Thin Shades" {
			isThinShades = true
		}
		if trait.TraitValue == "Oversized" {
			isOversized = true
		}
		if trait.TraitValue == "Geometric Shades" {
			isGeometricShades = true
		}

		if trait.TraitValue == "Bitcoin Ballers" {
			isBTCBallers = true
		}

		if trait.TraitValue == "Backwards Bandana" {
			isBackwardBandana = true
		}

		if trait.TraitValue == "Beanie" {
			isBeanie = true
		}

		if trait.TraitValue == "Knit Beanie" {
			isKnitBeanie = true
		}

		if trait.TraitValue == "Sweatband" {
			isSweatband = true
		}

		if trait.TraitValue == "Zipped Puffer" {
			isZippedPuffer = true
		}

		if trait.TraitValue == "Bandana" && trait.TraitType == "Head" {
			isBandanaHead = true
		}

		if trait.TraitValue == "Bandana" && trait.TraitType == "Mouth" {
			isBandanaMouth = true
		}
		if trait.TraitValue == "Dumbfounded" && trait.TraitType == "Mouth" {
			isDumbfounded = true
		}
		if trait.TraitValue == "Rose" && trait.TraitType == "Mouth" {
			isRose = true
		}
		if trait.TraitValue == "Tongue" && trait.TraitType == "Mouth" {
			isTongue = true
		}
		if trait.TraitValue == "Sakura" && trait.TraitType == "Head" {
			isSakura = true
		}
		if trait.TraitValue == "Panel Hat" && trait.TraitType == "Head" {
			isPanelHat = true
		}
	}

	if isFlameShades || isGeometricShades || isTheDonShades || isThickFrameShades || isThinShades || isSportShades || isOversized || isPlasmaVision || isBTCBallers {
		isGlasses = true
	}

	if isFlameShades || isOversized || isPlasmaVision || isBTCBallers {
		isBigGlasses = true
	}

	var mouthImage image.Image
	var headImage image.Image
	var glassesImage image.Image

	isBigHead := false
	if isSakura || isMessyHair || isTwoToneBraids || isDreadlocks || isTrooper {
		isBigHead = true
	}

	for _, trait := range m.Traits {

		if isTrooper && trait.TraitType == "Eyes" {
			draw.Draw(newImage, r, g.SpecialImages["Trooper Hat Right"], image.Point{0, 0}, draw.Over)
		}

		if trait.TraitType == "Eyes" && isGlasses {
			glassesImage = g.TraitMaps[trait.TraitType][trait.TraitValue].TraitImage

			if isOversized && isBeanie {
				glassesImage = g.SpecialImages["beanie-oversized-eyes"]
			} else if isThickFrameShades && isBackwardBandana {
				glassesImage = g.SpecialImages["backwards-bandana-thick-frame-glasses"]
			} else if isOversized && isBlondeBraids {
				glassesImage = g.SpecialImages["blonde-braids-oversized-glasses"]
			}

			draw.Draw(newImage, r, g.TraitMaps["Eyes"]["Bored"].TraitImage, image.Point{0, 0}, draw.Over)
			continue
		}

		if trait.TraitType == "Head" {
			headImage = g.TraitMaps[trait.TraitType][trait.TraitValue].TraitImage
		}

		if isBeanie && isOversized && trait.TraitType == "Head" {
			headImage = g.SpecialImages["beanie-oversized-head"]
			continue
		}

		if isRobot && isGoggles && trait.TraitType == "Head" {
			headImage = g.SpecialImages["goggles-robot-head"]
			draw.Draw(newImage, r, headImage, image.Point{0, 0}, draw.Over)
			continue
		}

		if isBeanie && isOversized && trait.TraitType == "Jewelry" {
			draw.Draw(newImage, r, headImage, image.Point{0, 0}, draw.Over)
			continue
		}

		if trait.TraitType == "Mouth" {
			mouthImage = g.TraitMaps[trait.TraitType][trait.TraitValue].TraitImage
			if isGlasses && !(isFlameShades && (isTrooper || isHelmet)) {
				draw.Draw(newImage, r, glassesImage, image.Point{0, 0}, draw.Over)
			}
			if isBigHead {
				draw.Draw(newImage, r, headImage, image.Point{0, 0}, draw.Over)
			}

			if isGlasses && isHelmet {
				croppedHelmet, err := cutter.Crop(headImage, cutter.Config{
					Width:   headImage.Bounds().Dx(),
					Height:  550,
					Options: cutter.Copy,
				})
				if err != nil {
					log.Fatal(err)
				}
				draw.Draw(newImage, r, croppedHelmet, image.Point{0, 0}, draw.Over)
			}

			if isGlasses && isGoggles && !(isPlasmaVision || isFlameShades) {
				croppedGoggles, err := cutter.Crop(headImage, cutter.Config{
					Width:   535,
					Height:  headImage.Bounds().Dy(),
					Options: cutter.Copy,
				})
				if err != nil {
					log.Fatal(err)
				}
				draw.Draw(newImage, r, croppedGoggles, image.Point{0, 0}, draw.Over)
				if isOversized {
					draw.Draw(newImage, r, g.SpecialImages["Oversized Goggle Line"], image.Point{0, 0}, draw.Over)
				}
			}

			if isTrooper && isBandanaMouth {
				draw.Draw(newImage, r, g.SpecialImages["Trooper Hat Bandana"], image.Point{0, 0}, draw.Over)
				continue
			}
		}

		if isSakura && trait.TraitType == "Jewelry" {
			draw.Draw(newImage, r, mouthImage, image.Point{0, 0}, draw.Over)
		}

		if (isTwoToneBraids || isDreadlocks) && trait.TraitType == "Jewelry" {
			draw.Draw(newImage, r, headImage, image.Point{0, 0}, draw.Over)
		}

		if (isTrooper || isBackwardHat) && isGrin && trait.TraitType == "Jewelry" {
			draw.Draw(newImage, r, g.SpecialImages["Grin Left"], image.Point{0, 0}, draw.Over)
		}

		if (isTrooper || isBackwardHat) && isRose && trait.TraitType == "Jewelry" {
			draw.Draw(newImage, r, mouthImage, image.Point{0, 0}, draw.Over)
		}

		// if (isTrooper || isBackwardHat || isBackwardBandana || isBandanaHead || isBeanie || isSweatband) && isBTCBallers && trait.TraitType == "Jewelry" {
		// 	draw.Draw(newImage, r, g.SpecialImages["BTC Ballers Top"], image.Point{0, 0}, draw.Over)
		// }

		if (isTrooper || isBackwardHat || isBackwardBandana || isBandanaHead || isBeanie || isSweatband) && isPlasmaVision && trait.TraitType == "Jewelry" {
			if isBandanaMouth {
				draw.Draw(newImage, r, g.SpecialImages["Plasma vision cut bottom"], image.Point{0, 0}, draw.Over)
			} else {
				draw.Draw(newImage, r, g.SpecialImages["Plasma vision"], image.Point{0, 0}, draw.Over)
				if isDumbfounded {
					draw.Draw(newImage, r, g.SpecialImages["Nostril"], image.Point{0, 0}, draw.Over)
				}
			}
		}

		// if (isHelmet) && isFlameShades && trait.TraitType == "Jewelry" {
		// 	draw.Draw(newImage, r, g.TraitMaps["Eyes"]["Flame Shades"].TraitImage, image.Point{0, 0}, draw.Over)
		// }

		// if (isTrooper) && isFlameShades && trait.TraitType == "Jewelry" {
		// 	draw.Draw(newImage, r, g.TraitMaps["Eyes"]["Flame Shades"].TraitImage, image.Point{0, 0}, draw.Over)
		// }

		// if (isTrooper || isBackwardHat || isBackwardBandana || isBandanaHead || isBeanie || isSweatband) && isSportShades && trait.TraitType == "Jewelry" {
		// 	draw.Draw(newImage, r, g.SpecialImages["Sport shades"], image.Point{0, 0}, draw.Over)
		// }

		// if (isTrooper || isBackwardHat || isBackwardBandana || isBandanaHead || isBeanie || isSweatband) && isFlameShades && trait.TraitType == "Jewelry" {
		// 	draw.Draw(newImage, r, g.SpecialImages["Flame shades"], image.Point{0, 0}, draw.Over)
		// }

		// if (isTrooper || isBackwardHat || isBackwardBandana || isBandanaHead || isBeanie || isSweatband) && isOversized && trait.TraitType == "Jewelry" {
		// 	draw.Draw(newImage, r, g.TraitMaps["Eyes"]["Oversized"].TraitImage, image.Point{0, 0}, draw.Over)
		// }

		// if (isTrooper || isBackwardHat || isBackwardBandana || isBandanaHead || isBeanie || isSweatband) && isGeometricShades && trait.TraitType == "Jewelry" {
		// 	draw.Draw(newImage, r, g.TraitMaps["Eyes"]["Geometric Shades"].TraitImage, image.Point{0, 0}, draw.Over)
		// }

		if isHelmet && trait.TraitType == "Clothes" {
			backgroundColor := newImage.At(500, 0)
			maskImage := image.NewRGBA(image.Rect(280, 400, 350, 530))
			draw.Draw(newImage, maskImage.Bounds(), &image.Uniform{backgroundColor}, image.ZP, draw.Src)
		}

		if isBackwardHat && isBandanaMouth && trait.TraitType == "Jewelry" {
			draw.Draw(newImage, r, g.SpecialImages["Bandana Left"], image.Point{0, 0}, draw.Over)
		}

		if isPlasmaVision && isBandanaMouth && trait.TraitType == "Head" {
			draw.Draw(newImage, r, g.SpecialImages["Plasma vision"], image.Point{0, 0}, draw.Over)
		} else if isPlasmaVision && trait.TraitType == "Head" {
			draw.Draw(newImage, r, g.SpecialImages["Plasma vision bottom"], image.Point{0, 0}, draw.Over)
			if isDumbfounded && trait.TraitType == "Head" {
				draw.Draw(newImage, r, g.SpecialImages["Nostril"], image.Point{0, 0}, draw.Over)
			}
		}

		if isTrooper && isRose && trait.TraitType == "Jewelry" {
			draw.Draw(newImage, r, mouthImage, image.Point{0, 0}, draw.Over)
		}

		if isFlameShades && (isTrooper || isHelmet) && trait.TraitType == "Jewelry" {
			draw.Draw(newImage, r, g.TraitMaps["Eyes"]["Flame Shades"].TraitImage, image.Point{0, 0}, draw.Over)
		}

		if isPanelHat && isSportShades && trait.TraitType == "Jewelry" {
			draw.Draw(newImage, r, g.TraitMaps["Head"]["Panel Hat"].TraitImage, image.Point{0, 0}, draw.Over)
		}

		if isZippedPuffer && isGrin && trait.TraitType == "Mouth" {
			mouthImage = g.SpecialImages["grin-puffer-mouth"]
			draw.Draw(newImage, r, g.SpecialImages["grin-puffer-mouth"], image.Point{0, 0}, draw.Over)
			continue
		}

		if isZippedPuffer && isSmallGrin && trait.TraitType == "Mouth" {
			mouthImage = g.SpecialImages["small-grin-puffer-mouth"]
			draw.Draw(newImage, r, g.SpecialImages["small-grin-puffer-mouth"], image.Point{0, 0}, draw.Over)
			continue
		}

		if isZippedPuffer && isRose && trait.TraitType == "Mouth" {
			mouthImage = g.SpecialImages["rose-puffer-mouth"]
			draw.Draw(newImage, r, g.SpecialImages["rose-puffer-mouth"], image.Point{0, 0}, draw.Over)
			continue
		}

		if isZippedPuffer && isDiscomfort && trait.TraitType == "Mouth" {
			mouthImage = g.SpecialImages["discomfort-puffer-mouth"]
			draw.Draw(newImage, r, g.SpecialImages["discomfort-puffer-mouth"], image.Point{0, 0}, draw.Over)
			continue
		}

		if isZippedPuffer && isBored && trait.TraitType == "Mouth" {
			mouthImage = g.SpecialImages["bored-puffer-mouth"]
			draw.Draw(newImage, r, g.SpecialImages["bored-puffer-mouth"], image.Point{0, 0}, draw.Over)
			continue
		}

		if isZippedPuffer && isBoredUnshaven && trait.TraitType == "Mouth" {
			mouthImage = g.SpecialImages["bored-unshaven-puffer-mouth"]
			draw.Draw(newImage, r, g.SpecialImages["bored-unshaven-puffer-mouth"], image.Point{0, 0}, draw.Over)
			continue
		}

		if isZippedPuffer && isPhenome && trait.TraitType == "Mouth" {
			mouthImage = g.SpecialImages["phenome-puffer-mouth"]
			draw.Draw(newImage, r, g.SpecialImages["phenome-puffer-mouth"], image.Point{0, 0}, draw.Over)
			continue
		}

		if isZippedPuffer && isTongue && trait.TraitType == "Mouth" {
			mouthImage = g.SpecialImages["tongue-puffer-mouth"]
			draw.Draw(newImage, r, g.SpecialImages["tongue-puffer-mouth"], image.Point{0, 0}, draw.Over)
			continue
		}

		if isBackwardBandana && isGlasses && !isBigGlasses && trait.TraitType == "Jewelry" {
			if isGeometricShades {
				croppedGlasses, err := cutter.Crop(headImage, cutter.Config{
					Width:   535,
					Height:  headImage.Bounds().Dy(),
					Options: cutter.Copy,
				})
				if err != nil {
					log.Fatal(err)
				}
				draw.Draw(newImage, r, croppedGlasses, image.Point{0, 0}, draw.Over)
			} else {
				draw.Draw(newImage, r, headImage, image.Point{0, 0}, draw.Over)
			}
		}

		if isRobot && (isKnitBeanie || isPanelHat) && trait.TraitType == "Jewelry" {
			robotImage := g.TraitMaps["Eyes"]["Robot"].TraitImage
			croppedRobot, err := cutter.Crop(robotImage, cutter.Config{
				Mode:    cutter.TopLeft,
				Anchor:  image.Point{855, 355},
				Width:   50,
				Height:  75,
				Options: cutter.Copy,
			})
			if err != nil {
				log.Fatal(err)
			}
			draw.Draw(newImage, r, croppedRobot, image.Point{0, 0}, draw.Over)
		}

		if isBackwardBandana && isRobot && trait.TraitType == "Jewelry" {
			robotImage := g.TraitMaps["Eyes"]["Robot"].TraitImage
			croppedRobot, err := cutter.Crop(robotImage, cutter.Config{
				Width:   robotImage.Bounds().Dx(),
				Height:  450,
				Options: cutter.Copy,
			})
			if err != nil {
				log.Fatal(err)
			}
			draw.Draw(newImage, r, croppedRobot, image.Point{0, 0}, draw.Over)
		}

		if isLasers && trait.TraitType == "Jewelry" {
			draw.Draw(newImage, r, g.SpecialImages["Laser"], image.Point{0, 0}, draw.Over)
		}

		if isJoint && trait.TraitType == "Jewelry" {
			draw.Draw(newImage, r, g.SpecialImages["Joint Smoke"], image.Point{0, 0}, draw.Over)
		}

		if trait.TraitValue == "__NONE__" {
			continue
		}

		draw.Draw(newImage, r, g.TraitMaps[trait.TraitType][trait.TraitValue].TraitImage, image.Point{0, 0}, draw.Over)
	}

	newImagePath := fmt.Sprintf("./tokens/%d.png", m.TokenID)
	newImageFile, err := os.Create(filepath.FromSlash(newImagePath))
	if err != nil {
		return fmt.Errorf("os.Create(%s) error: %w", newImagePath, err)
	}
	defer newImageFile.Close()

	err = png.Encode(newImageFile, newImage)
	if err != nil {
		return err
	}
	return nil
}

func main() {
	// rand.Seed(time.Now().UTC().UnixNano())
	// rand.Seed(int64(time.Now().Day()))
	rand.Seed(int64(time.Now().Year()))

	g, err := newGenerator("./traits")
	if err != nil {
		log.Fatal(err)
	}

	err = os.Mkdir(filepath.FromSlash("./tokens"), 0777)
	if err != nil && !os.IsExist(err) {
		log.Fatal(err)
	}

	count := 1000
	bar := progressbar.Default(int64(count))
	sem := make(chan struct{}, 1)
	wg := &sync.WaitGroup{}
	for i := 0; i < count; i++ {
		sem <- struct{}{}
		wg.Add(1)
		go func(i int) {

			defer func() {
				bar.Add(1)
				wg.Done()
				<-sem
			}()

			metadata, err := g.GenerateMetadata(i)
			if err != nil {
				log.Fatal(err)
			}

			err = g.GenerateImage(metadata)
			if err != nil {
				log.Fatal(err)
			}

		}(i)
	}
	wg.Wait()
}
