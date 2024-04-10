# datapoint

A simple Go library for the Met Office DataPoint API

Interactions have been developed against the endpoints described in this document: https://www.metoffice.gov.uk/binaries/content/assets/metofficegovuk/pdf/data/datapoint_api_reference.pdf along with experimental testing

This module aims to provide some wrapper types around the raw responses to make them easier to work with in code, and
therefore the objects will not map to exactly what is returned from the server.

# Quick Start

This will give you a quick setup with the DataPoint library and show you how to print the temperature in Exeter

```go
client, err := dp.NewClient(dp.WithApiKey("00000000-0000-0000-0000-000000000000"))
if err != nil {
    panic(err)
}

list, err := client.ForecastSiteList()
if err != nil {
    panic(err)
}

exeterId := -1
for _, site := range list {
    if site.Name == "Exeter" {
        exeterId = site.Id
    }
}

if exeterId == -1 {
    panic("Could not find exeter")
}

forecast, err := client.FiveDayForecast(dp.ResolutionDaily, exeterId, nil)
if err != nil {
    panic(err)
}

fmt.Printf("Forecast was generated at: %v\n", forecast.DataDate)
for _, period := range forecast.Location.Period {
    fmt.Printf("  %v (%v)\n", period.Time, period.Type)
    for _, fore := range period.Forecasts {
        temp := 0
        unit := ""
        t, ok := fore.IntParams[string(dp.KnownParameterTemperature)]
        if ok {
            temp = t.Value
            unit = t.Units
        } else {
            t, ok = fore.IntParams[string(dp.KnownParameterDayMaximumTemperature)]
            if ok {
                temp = t.Value
                unit = t.Units
            } else {
                t, ok = fore.IntParams[string(dp.KnownParameterNightMinimumTemperature)]
                if ok {
                    temp = t.Value
                    unit = t.Units
                } else {
                    panic("Could not find temperature in forecast!")
                }
            }
        }
        fmt.Printf("    %v: %v%v\n", fore.Time, temp, unit)
    }
}
```

> [!NOTE]
> This is quite a bad way to do this, but it gives you an idea