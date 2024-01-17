package main

import (
	"fmt"
	"image"
	"log"
	"os"
	"path/filepath"
	"strings"

	"image/png"
	_ "image/png"

	"github.com/goccy/go-yaml"
)

func GetImage(path string) (image.Image, error) {
	imageFile, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer imageFile.Close()
	img, _, err := image.Decode(imageFile)
	return img, err
}

type TraitData struct {
	TraitType        string
	TraitValue       string
	TraitProbability int
	TraitImage       image.Image
	ImagePath        string
}

func GetTraits(traitType, path string) ([]TraitData, error) {
	traitPaths, err := filepath.Glob(path)
	if err != nil {
		return nil, err
	}

	traits := []TraitData{}
	for _, traitPath := range traitPaths {

		img, err := GetImage(traitPath)
		if err != nil {
			return nil, err
		}
		fileName := filepath.Base(traitPath)

		fileNameParts := strings.Split(fileName, " - ")
		if len(fileNameParts) != 2 {
			fileNameParts = strings.Split(fileName, ".")
			if len(fileNameParts) != 2 {
				log.Fatalf("%s split doesn't equal two or three parts %d", fileName, len(fileNameParts))
			}
		}
		traitValue := fileNameParts[0]
		traits = append(traits, TraitData{
			TraitType:        traitType,
			TraitValue:       traitValue,
			TraitProbability: 1000,
			TraitImage:       img,
			ImagePath:        traitPath,
		})
	}

	return traits, nil
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

func main() {
	// get all trait types and values

	d := Data{}
	d.Traits = make(map[string]YamlTraitData)

	traits := []string{"Background", "Fur", "Clothes", "Eyes", "Mouth", "Head", "Jewelry"}
	for _, traitType := range traits {
		traitMap, err := GetTraits(traitType, fmt.Sprintf("%s/%s/*.png", "./traits", traitType))
		if err != nil {
			log.Fatal(err)
		}
		for _, trait := range traitMap {
			fmt.Println(trait.TraitType, trait.TraitValue)
			imagePath := fmt.Sprintf("traits/%s/%s.png", traitType, trait.TraitValue)
			datum := Datum{trait.TraitValue, imagePath, trait.TraitProbability, true}

			yamlTraitData, ok := d.Traits[trait.TraitType]
			if !ok {
				d.Traits[trait.TraitType] = YamlTraitData{}
				yamlTraitData = d.Traits[trait.TraitType]
			}

			yamlTraitData.Values = append(yamlTraitData.Values, datum)

			d.Traits[trait.TraitType] = yamlTraitData

			err = os.MkdirAll(fmt.Sprintf("./traits/%s", trait.TraitType), 0755)
			if err != nil {
				log.Fatal(err)
			}

			f, err := os.Create(fmt.Sprintf("./traits/%s/%s.png", trait.TraitType, trait.TraitValue))
			if err != nil {
				log.Fatal(err)
			}

			err = png.Encode(f, trait.TraitImage)
			if err != nil {
				log.Fatal(err)
			}
		}
	}

	// save as yaml file
	f, err := os.Create("./abbc.yml")
	if err != nil {
		log.Fatal(err)
	}
	encoder := yaml.NewEncoder(f)
	err = encoder.Encode(d)
	if err != nil {
		log.Fatal(err)
	}
}
