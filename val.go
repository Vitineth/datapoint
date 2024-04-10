package datapoint

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"slices"
	"strconv"
	"time"
)

type siteResponse struct {
	Locations struct {
		Location []struct {
			Elevation       string `json:"elevation"`
			Id              string `json:"id"`
			Latitude        string `json:"latitude"`
			Longitude       string `json:"longitude"`
			Name            string `json:"name"`
			Region          string `json:"region"`
			UnitaryAuthArea string `json:"unitaryAuthArea"`
		} `json:"Location"`
	} `json:"Locations"`
}

// A Site represents a single location for which 3 hourly forecasts are available
type Site struct {
	// Id [official]: The ID number of the location e.g. '310069'
	Id int `json:"id"`
	// Latitude [official]: The latitude of the location in decimal degrees e.g. '50.7179'
	Latitude float64 `json:"latitude"`
	// Longitude [official]: The longitude of the location in decimal degrees e.g. '-3.5327'
	Longitude float64 `json:"longitude"`
	// Name [official]: The name of the location e.g. 'Exeter'
	Name string `json:"name"`
	// Elevation [undocumented]: The elevation of the location
	Elevation float64 `json:"elevation"`
	// Region [undocumented]: The region of the location
	Region string `json:"region"`
	// UnitaryAuthArea [undocumented]: The unitary auth area of the location
	UnitaryAuthArea string `json:"unitaryAuthArea"`
}

// ForecastSiteList returns the 5,000 UK locations forecast site list data feed provides a list of the locations (also known as sites) for which
// results are available for the 5,000 UK locations three hourly forecast and 5,000 UK locations daily forecast data
// feeds. You can use this data feed to find details such as the ID of the site that you are interested in finding data
// for.
func (d *DataPointClient) ForecastSiteList() ([]Site, error) {
	return d.siteList("wxfcs")
}

// ObservationSiteList returns a list of locations (also known as sites) for which
// results are available for the hourly observations data feed.
// You can use this to find the ID of the site that you are
// interested in
func (d *DataPointClient) ObservationSiteList() ([]Site, error) {
	return d.siteList("wxobs")
}

func (d *DataPointClient) siteList(id string) ([]Site, error) {
	body, target, err := d.fetch("sitelist", "val/"+id+"/all/json/sitelist", nil)
	if err != nil {
		return nil, err
	}

	var result siteResponse
	err = json.Unmarshal(body, &result)
	if err != nil {
		return nil, fmt.Errorf("failed to deserialise body from %v for sitelist: %w", target, err)
	}

	entries := make([]Site, len(result.Locations.Location))
	for i, site := range result.Locations.Location {
		latitude, err := strconv.ParseFloat(site.Latitude, 64)
		if err != nil {
			return nil, fmt.Errorf("failed to parse latitude %v for sitelist: %w", site.Latitude, err)
		}
		longitude, err := strconv.ParseFloat(site.Longitude, 64)
		if err != nil {
			return nil, fmt.Errorf("failed to parse longitude %v for sitelist: %w", site.Longitude, err)
		}
		id, err := strconv.ParseInt(site.Id, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("failed to parse id %v for sitelist: %w", site.Id, err)
		}
		var elevation float64 = 0
		if site.Elevation != "" {
			elev, err := strconv.ParseFloat(site.Elevation, 64)
			if err != nil {
				slog.Warn("failed to parse elevation", "elevation", site.Elevation, "err", err)
			} else {
				elevation = elev
			}
		}

		entries[i] = Site{
			Id:              int(id),
			Latitude:        latitude,
			Longitude:       longitude,
			Name:            site.Name,
			Elevation:       elevation,
			Region:          site.Region,
			UnitaryAuthArea: site.UnitaryAuthArea,
		}
	}

	return entries, nil
}

type capabilitiesResponse struct {
	Resource struct {
		DataDate   string     `json:"dataDate"`
		Resolution Resolution `json:"res"`
		Type       string     `json:"type"`
		TimeSteps  struct {
			TS []string `json:"TS"`
		} `json:"TimeSteps"`
	} `json:"Resource"`
}

// TimeSteps represents the set of available time step that can be queried for the given start and resolution values
type TimeSteps struct {
	// DataDate [official]: The date and time at which the data was last updated, expressed according to the ISO
	// 8601 combined date and time convention. e.g. '2012-11-21T15:00:00Z'
	DataDate time.Time
	// Resolution [official]: The temporal resolution of the web service for which the capabilities have been returned.
	// This is set to the temporal resolution specified in the query. e.g. 'daily', '3hourly' or
	// 'hourly'
	Resolution Resolution
	// Type [official]: The resource type of the web service for which the capabilities have been returned. e.g.
	// 'wxfcs' or 'wxobs'
	Type string
	// TimeSteps [official]: The value of each TS object (or each element in the TS array in the JSON representation)
	// provides a description of a single available timestep, expressed according to the ISO 8601
	// combined date and time convention. e.g. '2012-11-21T06:00:00Z
	TimeSteps []time.Time
}

// ForecastTimeStepCapabilities exposes the capabilities data feed which provides a summary of the timesteps for which results are available for the 5,000 UK
// locations daily and three hourly forecast data feed. You can use this data feed to check that the timestep you are
// interested in is available before querying the relevant web service to get the data. In this way you can minimise the
// number of redundant calls that have to be made.
func (d *DataPointClient) ForecastTimeStepCapabilities(resolution Resolution) (*TimeSteps, error) {
	body, _, err := d.fetch(
		"capabilities",
		"val/wxfcs/all/json/capabilities",
		map[string]string{
			"res": string(resolution),
		},
	)
	if err != nil {
		return nil, err
	}

	var ts capabilitiesResponse
	err = json.Unmarshal(body, &ts)
	if err != nil {
		return nil, fmt.Errorf("failed to deserialise body for capabilities: %w", err)
	}

	dataDate, err := time.Parse(time.RFC3339, ts.Resource.DataDate)
	if err != nil {
		return nil, fmt.Errorf("failed to parse time step for capabilities: %w", err)
	}

	times := make([]time.Time, len(ts.Resource.TimeSteps.TS))
	for i, t := range ts.Resource.TimeSteps.TS {
		t, err := time.Parse(time.RFC3339, t)
		if err != nil {
			return nil, fmt.Errorf("failed to parse time step for capabilities: %w", err)
		}
		times[i] = t
	}

	return &TimeSteps{
		DataDate:   dataDate,
		Resolution: ts.Resource.Resolution,
		Type:       ts.Resource.Type,
		TimeSteps:  times,
	}, nil
}

type wxResponse struct {
	Param []struct {
		Name        string `json:"name"`
		Units       string `json:"units"`
		Description string `json:"$"`
	} `json:"Param"`
}

type locationEntry struct {
	Id        string `json:"i"`
	Latitude  string `json:"lat"`
	Longitude string `json:"lon"`
	Name      string `json:"name"`
	Country   string `json:"country"`
	Continent string `json:"continent"`
	Elevation string `json:"elevation"`
	Period    []struct {
		Type  string              `json:"type"`
		Value string              `json:"value"`
		Rep   []map[string]string `json:"Rep"`
	} `json:"Period"`
}

type dvEntry struct {
	DataDate string        `json:"dataDate"`
	Type     string        `json:"type"`
	Location locationEntry `json:"Location"`
}

type dvMultipleEntry struct {
	DataDate string          `json:"dataDate"`
	Type     string          `json:"type"`
	Location []locationEntry `json:"Location"`
}

type siteRepResponse struct {
	SiteRep struct {
		Wx wxResponse `json:"Wx"`
		Dv dvEntry    `json:"DV"`
	} `json:"SiteRep"`
}
type siteRepAllResponse struct {
	SiteRep struct {
		Wx wxResponse      `json:"Wx"`
		Dv dvMultipleEntry `json:"DV"`
	} `json:"SiteRep"`
}

// ParameterDescriptor contains the definition of one of the attributes in a single forecast
type ParameterDescriptor struct {
	// Name [official] is the name of the attribute
	Name string
	// Units [official] is the unit in which the attribute is represented
	Units string
	// Description [official - $] is the textual description of what the attribute represents
	Description string
}

// StringParameterValue is a combination of a string value and a parameter definition
type StringParameterValue struct {
	ParameterDescriptor
	// Value [official] the value of the measure
	Value string
}

// IntParameterValue is a combination of an int value and a parameter definition
type IntParameterValue struct {
	ParameterDescriptor
	// Value [official] the value of the measure
	Value int
}

// Forecast represents a single forecast issues for a specific time, which contains a set of parameters describing the weather at that time
type Forecast struct {
	// Time [unofficial] is the time at which this forecast represents, this is derived from the time offset returned in the response
	Time time.Time
	// IntParams contains all known parameters which will always be numeric, anything else will appear in StringParams
	IntParams map[string]IntParameterValue
	// StringParams contains all parameters which are not known to be integers
	StringParams map[string]StringParameterValue
}

// Period represents a single time period which can contain any number of forecasts based on the resolution.
type Period struct {
	// Type [official] is the type of period this represents, usually Day
	Type string
	// Time [official] is the start date of this period
	Time time.Time
	// Forecasts [official] is the set of forecasts for this period
	Forecasts []Forecast
}

// LocationRep contains a set of periods and forecasts for a single location area
type LocationRep struct {
	// Id [official - i] is the ID number of the location
	Id int
	// Latitude [official - lan] is the latitude of the location in decimal degrees
	Latitude float64
	// Longitude [official - lon] is the longitude of the location in decimal degrees
	Longitude float64
	// Name [official] is the name of the location
	Name string
	// Country [official] is the country of the location
	Country string
	// Continent [official] is the continent of the location
	Continent string
	// Elevation [unofficial] is the elevation of the location, not always returned
	Elevation float64
	// Period is the set of periods for which forecasts are available
	Period []Period
}

// SiteRep is a forecast entry for a single location
type SiteRep struct {
	// DataDate [official] is the date and time at which the forecast was run
	DataDate time.Time
	// Type [official] is the type of data that is returned (Forecast or Obs)
	Type string
	// Location [official] is the combination of location information and forecast data
	Location LocationRep
}

var (
	knownIntParams = []string{
		string(KnownParameterFeelsLikeTemp),
		string(KnownParameterWindGust),
		string(KnownParameterScreenRelativeHumidity),
		string(KnownParameterPrecipitationProbability),
		string(KnownParameterWindSpeed),
		string(KnownParameterTemperature),
		string(KnownParameterWeatherType),
		string(KnownParameterMaxUvIndex),
		string(KnownParameterWindGustNoon),
		string(KnownParameterScreenRelativeHumidityNoon),
		string(KnownParameterDayMaximumTemperature),
		string(KnownParameterNightMinimumTemperature),
		string(KnownParameterFeelsLikeDayMaximumTemperature),
		string(KnownParameterPrecipitationProbabilityDay),
		string(KnownParameterPrecipitationProbabilityNight),
		"$",
	}
)

func convertLocation(paramDefinitions map[string]ParameterDescriptor, typeName string, startTime time.Time, entry locationEntry) (*SiteRep, error) {
	periods := make([]Period, len(entry.Period))
	for i, entry := range entry.Period {
		periodTime, err := time.Parse("2006-01-02Z", entry.Value)
		if err != nil {
			return nil, fmt.Errorf("failed to parse period offset: %w", err)
		}

		forecasts := make([]Forecast, len(entry.Rep))
		for forecastIndex, rep := range entry.Rep {
			intParams := map[string]IntParameterValue{}
			stringParams := map[string]StringParameterValue{}

			offsetString, ok := rep["$"]
			if !ok {
				return nil, errors.New("could not find forecast offset")
			}

			var offset int
			if offsetString == "Day" {
				offset = 0
			} else if offsetString == "Night" {
				offset = 86399 // 23 hours, 59 minutes, 59 seconds
			} else {
				offset64, err := strconv.ParseInt(offsetString, 10, 16)
				if err != nil {
					return nil, fmt.Errorf("failed to parse forecast offset: %w", err)
				}
				offset = int(offset64) * 60
			}

			for k, v := range rep {
				if k == "$" {
					continue
				}

				descriptor, ok := paramDefinitions[k]
				if !ok {
					return nil, fmt.Errorf("could not find descriptor for parameter %v", k)
				}

				if slices.Contains(knownIntParams, k) {
					v, err := strconv.ParseInt(v, 10, 16)
					if err != nil {
						return nil, fmt.Errorf("failed to parse known int value %v: %w", k, err)
					}
					intParams[k] = IntParameterValue{
						ParameterDescriptor: descriptor,
						Value:               int(v),
					}
				} else {
					stringParams[k] = StringParameterValue{
						ParameterDescriptor: descriptor,
						Value:               v,
					}
				}
			}

			forecasts[forecastIndex] = Forecast{
				Time:         periodTime.Add(time.Duration(offset) * time.Second),
				IntParams:    intParams,
				StringParams: stringParams,
			}
		}

		periods[i] = Period{
			Type:      entry.Type,
			Time:      periodTime,
			Forecasts: forecasts,
		}
	}

	id, err := strconv.ParseInt(entry.Id, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("failed to parse id: %w", err)
	}

	lat, err := strconv.ParseFloat(entry.Latitude, 64)
	if err != nil {
		return nil, fmt.Errorf("failed to parse latitude: %w", err)
	}

	lon, err := strconv.ParseFloat(entry.Longitude, 64)
	if err != nil {
		return nil, fmt.Errorf("failed to parse longitude: %w", err)
	}

	var elevation float64 = 0
	if entry.Elevation != "" {
		elevation, err = strconv.ParseFloat(entry.Elevation, 64)
		if err != nil {
			return nil, fmt.Errorf("failed to parse elevation: %w", err)
		}
	}

	return &SiteRep{
		DataDate: startTime,
		Type:     typeName,
		Location: LocationRep{
			Id:        int(id),
			Latitude:  lat,
			Longitude: lon,
			Name:      entry.Name,
			Country:   entry.Country,
			Continent: entry.Continent,
			Elevation: elevation,
			Period:    periods,
		},
	}, nil
}

// FiveDayForecast provides access to daily and three hourly forecast data from the Met Office for each of the roughly 5,000 sites
// for which the Met Office provides data. The forecast data is provided for time steps that are three hours apart, or
// daily (day and night), starting with the time at which the forecast was last run, and ending approximately five days
// later (meaning that approximately 10 or 40 forecast timesteps are available for each site). The data provided by the
// web service is updated on an hourly basis, and at any given point in time the exact set of timesteps that are
// available can be obtained using the capabilities web service. For a full list of the 5,000 sites, call the 5,000 UK
// locations site list data feed.
func (d *DataPointClient) FiveDayForecast(resolution Resolution, locationID int, at *time.Time) (*SiteRep, error) {
	params := map[string]string{
		"res": string(resolution),
	}
	if at != nil {
		params["time"] = at.Format(time.RFC3339)
	}
	body, target, err := d.fetch("locationId", "val/wxfcs/all/json/"+strconv.Itoa(locationID), params)
	if err != nil {
		return nil, err
	}

	var result siteRepResponse
	err = json.Unmarshal(body, &result)
	if err != nil {
		return nil, fmt.Errorf("failed to deserialise body from %v for sitelist: %w", target, err)
	}

	if result.SiteRep.Dv.Location.Id == "" {
		return nil, nil
	}

	paramDefinitions := map[string]ParameterDescriptor{}
	for _, p := range result.SiteRep.Wx.Param {
		paramDefinitions[p.Name] = ParameterDescriptor{
			Name:        p.Name,
			Units:       p.Units,
			Description: p.Description,
		}
	}

	startTime, err := time.Parse(time.RFC3339, result.SiteRep.Dv.DataDate)
	if err != nil {
		return nil, fmt.Errorf("failed to parse period start time: %w", err)
	}

	return convertLocation(
		paramDefinitions,
		result.SiteRep.Dv.Type,
		startTime,
		result.SiteRep.Dv.Location,
	)
}

// FiveDayForecastForAllLocations implements the same functionality as FiveDayForecast but returns the results for all
// locations supported by the DataPoint service. The result for this will be quite large
func (d *DataPointClient) FiveDayForecastForAllLocations(resolution Resolution, at *time.Time) ([]SiteRep, error) {
	params := map[string]string{
		"res": string(resolution),
	}
	if at != nil {
		params["time"] = at.Format(time.RFC3339)
	}
	body, target, err := d.fetch("locationId", "val/wxfcs/all/json/all", params)
	if err != nil {
		return nil, err
	}

	var result siteRepAllResponse
	err = json.Unmarshal(body, &result)
	if err != nil {
		return nil, fmt.Errorf("failed to deserialise body from %v for sitelist: %w", target, err)
	}

	paramDefinitions := map[string]ParameterDescriptor{}
	for _, p := range result.SiteRep.Wx.Param {
		paramDefinitions[p.Name] = ParameterDescriptor{
			Name:        p.Name,
			Units:       p.Units,
			Description: p.Description,
		}
	}

	startTime, err := time.Parse(time.RFC3339, result.SiteRep.Dv.DataDate)
	if err != nil {
		return nil, fmt.Errorf("failed to parse period start time: %w", err)
	}

	reps := make([]SiteRep, len(result.SiteRep.Dv.Location))
	for i, entry := range result.SiteRep.Dv.Location {
		r, err := convertLocation(
			paramDefinitions,
			result.SiteRep.Dv.Type,
			startTime,
			entry,
		)
		if err != nil {
			return nil, err
		}

		reps[i] = *r
	}

	return reps, nil
}
