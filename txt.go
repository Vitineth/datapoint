package datapoint

import (
	"encoding/json"
	"fmt"
	"strconv"
	"time"
)

// ExtremeCapabilities describes the last update to a UK extreme measurement
type ExtremeCapabilities struct {
	// ExtremeDate [official] is the date of the observation
	ExtremeDate time.Time
	// IssuedAt [official] is the date at which the observation was issued
	IssuedAt time.Time
}

type extremeCapabilitiesResponse struct {
	UkExtremes struct {
		ExtremeDate string `json:"extremeDate"`
		IssuedAt    string `json:"issuedAt"`
	} `json:"UkExtremes"`
}

// UkExtremesCapabilities indicates when the regional extremes observations data feed was last updated, and the period it covers
func (d *DataPointClient) UkExtremesCapabilities() (*ExtremeCapabilities, error) {
	body, target, err := d.fetch("sitelist", "txt/wxobs/ukextremes/json/capabilities", nil)
	if err != nil {
		return nil, err
	}

	var result extremeCapabilitiesResponse
	err = json.Unmarshal(body, &result)
	if err != nil {
		return nil, fmt.Errorf("failed to deserialise body from %v for capabilities: %w", target, err)
	}

	extremeDate, err := time.Parse(time.DateOnly, result.UkExtremes.ExtremeDate)
	if err != nil {
		return nil, fmt.Errorf("failed to parse date %v for extreme date: %w", result.UkExtremes.ExtremeDate, err)
	}
	issuedAt, err := time.Parse(time.RFC3339, result.UkExtremes.IssuedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to parse date %v for issued at date: %w", result.UkExtremes.IssuedAt, err)
	}

	return &ExtremeCapabilities{
		ExtremeDate: extremeDate,
		IssuedAt:    issuedAt,
	}, nil
}

// Extreme represents a single extreme reading
type Extreme struct {
	// LocationId [official] is location ID of the location where the extreme was observed. The location ID may not
	// be listed in the 5,000 locations resource.
	LocationId int
	// LocationName [official] is the full name of the location where the extreme was observed.
	LocationName string
	// Type [official] is the type of the extreme. For example 'HMAXT' would represent the highest maximum
	//temperature, and 'LMINT' would represent the lowest minimum temperature.
	Type string
	// UnitOfMeasurement [official - uom] is the unit of measurement for the extreme
	UnitOfMeasurement string
	// Value [official - $] is the value fo the observed extreme, in units specified in UnitOfMeasurement
	Value float64
}

// Region represents a region in which extreme values have been observed
type Region struct {
	// Id [official] is the short name of the region
	Id string
	// Name [official] is the full name of the region
	Name string
	// Extremes [official] holds the set of extreme measurements found in this region
	Extremes []Extreme
}

// LatestExtremes represents the response from the latest UK extremes endpoints
type LatestExtremes struct {
	// ExtremeDate [official] is the date of the observation
	ExtremeDate time.Time
	// IssuedAt [official] is the date at which the observation was issued
	IssuedAt time.Time
	// Regions [official] are the regions in which observations are found
	Regions []Region
}

type latestExtremesResponse struct {
	UkExtremes struct {
		ExtremeDate string `json:"extremeDate"`
		IssuedAt    string `json:"issuedAt"`
		Regions     struct {
			Region []struct {
				Id       string `json:"id"`
				Name     string `json:"name"`
				Extremes struct {
					Extreme []struct {
						LocationId   string `json:"locId"`
						LocationName string `json:"locationName"`
						Type         string `json:"type"`
						Uom          string `json:"uom"`
						Value        string `json:"$"`
					} `json:"Extreme"`
				} `json:"Extremes"`
			} `json:"Region"`
		} `json:"Regions"`
	} `json:"UkExtremes"`
}

// UkExtremesLatest provides access to the observed extremes of weather across the UK for the day of issue. The data provided by
// the web service is updated on a daily basis.
func (d *DataPointClient) UkExtremesLatest() (*LatestExtremes, error) {
	body, target, err := d.fetch("ukextremeslatest", "txt/wxobs/ukextremes/json/latest", nil)
	if err != nil {
		return nil, err
	}

	var result latestExtremesResponse
	err = json.Unmarshal(body, &result)
	if err != nil {
		return nil, fmt.Errorf("failed to deserialise body from %v for uk extremes latest: %w", target, err)
	}

	regions := make([]Region, len(result.UkExtremes.Regions.Region))
	for i, region := range result.UkExtremes.Regions.Region {
		extremes := make([]Extreme, len(region.Extremes.Extreme))
		for j, extreme := range region.Extremes.Extreme {
			locationId, err := strconv.ParseInt(extreme.LocationId, 10, 64)
			if err != nil {
				return nil, fmt.Errorf("failed to parse location id: %w", err)
			}

			value, err := strconv.ParseFloat(extreme.Value, 64)
			if err != nil {
				return nil, fmt.Errorf("failed to parse value: %w", err)
			}

			extremes[j] = Extreme{
				LocationId:        int(locationId),
				LocationName:      extreme.LocationName,
				Type:              extreme.Type,
				UnitOfMeasurement: extreme.Uom,
				Value:             value,
			}
		}

		regions[i] = Region{
			Id:       region.Id,
			Name:     region.Name,
			Extremes: extremes,
		}
	}

	extremeDate, err := time.Parse(time.DateOnly, result.UkExtremes.ExtremeDate)
	if err != nil {
		return nil, fmt.Errorf("failed to parse the extreme date: %w", err)
	}

	issuedAt, err := time.Parse(time.RFC3339, result.UkExtremes.IssuedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to parse the issued at date: %w", err)
	}

	return &LatestExtremes{
		ExtremeDate: extremeDate,
		IssuedAt:    issuedAt,
		Regions:     regions,
	}, nil
}

// RegionalForecastSite is an individual location for which a regional forecast is available
type RegionalForecastSite struct {
	// Id [official] The ID of region
	Id int
	// Name [official] The short name of the region
	Name string
}

type regionalForecastSiteListResponse struct {
	Locations struct {
		Location []struct {
			Id   string `json:"@id"`
			Name string `json:"@name"`
		} `json:"Location"`
	} `json:"Locations"`
}

// RegionalForecastSiteList provides a list of the locations (also known as sites) for which results are
// available for the regional forecast data feed. You can use this data feed to find details such as the ID of the region
// that you are interested in finding data for
func (d *DataPointClient) RegionalForecastSiteList() ([]RegionalForecastSite, error) {
	body, target, err := d.fetch("regional forecast site list", "txt/wxfcs/regionalforecast/json/sitelist", nil)
	if err != nil {
		return nil, err
	}

	var result regionalForecastSiteListResponse
	err = json.Unmarshal(body, &result)
	if err != nil {
		return nil, fmt.Errorf("failed to deserialise body from %v for regional forecast site list: %w", target, err)
	}

	locations := make([]RegionalForecastSite, len(result.Locations.Location))
	for i, s := range result.Locations.Location {
		id, err := strconv.ParseInt(s.Id, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("could not parse location id: %w", err)
		}

		locations[i] = RegionalForecastSite{
			Id:   int(id),
			Name: s.Name,
		}
	}

	return locations, nil
}

// RegionalForecastCapabilities indicates when the last set of regional forecasts were issues by the Met Office
type RegionalForecastCapabilities struct {
	IssuedAt time.Time
}

type regionalForecastCapabilitiesResponse struct {
	RegionalForecast struct {
		IssuedAt string `json:"issuedAt"`
	} `json:"RegionalFcst"`
}

// RegionalForecastCapabilities provides a summary of the results that are available from the regional forecast data feed,
// specifying when the forecast was last updated
func (d *DataPointClient) RegionalForecastCapabilities() (*RegionalForecastCapabilities, error) {
	body, target, err := d.fetch("regional forecast capabilities", "txt/wxfcs/regionalforecast/json/capabilities", nil)
	if err != nil {
		return nil, err
	}

	var result regionalForecastCapabilitiesResponse
	err = json.Unmarshal(body, &result)
	if err != nil {
		return nil, fmt.Errorf("failed to deserialise body from %v for regional forecast capabilities: %w", target, err)
	}

	issued, err := time.Parse(result.RegionalForecast.IssuedAt, time.RFC3339)
	if err != nil {
		return nil, fmt.Errorf("failed to parse issued at time: %w", err)
	}

	return &RegionalForecastCapabilities{IssuedAt: issued}, nil
}

//
//type RegionalForecastPeriod struct {
//	Id string
//}
//
//type RegionalForecast struct {
//	Created  time.Time
//	IssuedAt time.Time
//	RegionId string
//}
//
//type regionalForecastSiteResponse struct {
//	RegionalForecast struct {
//		CreatedOn       string `json:"createdOn"`
//		IssuedAt        string `json:"issuedAt"`
//		RegionId        string `json:"regionId"`
//		ForecastPeriods struct {
//			Period []struct {
//				Id        string `json:"id"`
//				Paragraph []struct {
//					Title string `json:"title"`
//					Body  string `json:"$"`
//				} `json:"Paragraph"`
//			} `json:"Period"`
//		} `json:"FcstPeriods"`
//	} `json:"RegionalFcst"`
//}
