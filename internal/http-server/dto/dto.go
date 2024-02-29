package dto

import (
	"encoding/json"
	"io"
)

type Dto interface {
}

func JSONToDto[T Dto](r io.Reader, body *T) error {
	err := json.NewDecoder(r).Decode(body)
	if err != nil {
		return err
	}
	return nil
}

type CountryDTO struct {
	Count     int    `json:"count"`
	Name      string `json:"name"`
	Countries []struct {
		CountryId   string  `json:"country_id"`
		Probability float64 `json:"probability"`
	} `json:"country"`
}

func (d *CountryDTO) GetPossibleNation() string {
	maxPos := 0.0
	var country string
	for _, c := range d.Countries {
		if c.Probability > maxPos {
			maxPos = c.Probability
			country = c.CountryId
		}
	}
	return country
}

type AgeDTO struct {
	Count int    `json:"count"`
	Name  string `json:"name"`
	Age   int    `json:"age"`
}

type GenderDTO struct {
	Count       int     `json:"count"`
	Name        string  `json:"name"`
	Gender      string  `json:"gender"`
	Probability float32 `json:"probability"`
}
