package datapoint

type Resolution string

const (
	ResolutionThreeHourly Resolution = "3hourly"
	ResolutionDaily       Resolution = "daily"
)

type KnownParameter string

const (
	KnownParameterFeelsLikeTemp                  KnownParameter = "F"
	KnownParameterWindGust                       KnownParameter = "G"
	KnownParameterWindGustNoon                   KnownParameter = "Gn"
	KnownParameterScreenRelativeHumidity         KnownParameter = "H"
	KnownParameterScreenRelativeHumidityNoon     KnownParameter = "Hn"
	KnownParameterTemperature                    KnownParameter = "T"
	KnownParameterDayMaximumTemperature          KnownParameter = "Dm"
	KnownParameterNightMinimumTemperature        KnownParameter = "Nm"
	KnownParameterFeelsLikeDayMaximumTemperature KnownParameter = "FDm"
	KnownParameterVisibility                     KnownParameter = "V"
	KnownParameterWindDirection                  KnownParameter = "D"
	KnownParameterWindSpeed                      KnownParameter = "S"
	KnownParameterMaxUvIndex                     KnownParameter = "U"
	KnownParameterWeatherType                    KnownParameter = "W"
	KnownParameterPrecipitationProbability       KnownParameter = "Pp"
	KnownParameterPrecipitationProbabilityDay    KnownParameter = "PPd"
	KnownParameterPrecipitationProbabilityNight  KnownParameter = "PPn"
)

type UvIndex int

func (i UvIndex) IsLowExposure() bool {
	return i >= 1 && i <= 2
}

func (i UvIndex) IsModerateExposure() bool {
	return i >= 3 && i <= 5
}

func (i UvIndex) IsHighExposure() bool {
	return i >= 6 && i <= 7
}

func (i UvIndex) IsVeryHighExposure() bool {
	return i >= 8 && i <= 10
}

func (i UvIndex) IsExtremeExposure() bool {
	return i >= 11
}

type WeatherType int

const (
	WeatherTypeClearNight           WeatherType = 0
	WeatherTypeSunnyDay             WeatherType = 1
	WeatherTypePartlyCloudyNight    WeatherType = 2
	WeatherTypePartlyCloudyDay      WeatherType = 3
	WeatherTypeMist                 WeatherType = 5
	WeatherTypeFog                  WeatherType = 6
	WeatherTypeCloudy               WeatherType = 7
	WeatherTypeOvercast             WeatherType = 8
	WeatherTypeLightRainShowerNight WeatherType = 9
	WeatherTypeLightRainShowerDay   WeatherType = 10
	WeatherTypeDrizzle              WeatherType = 11
	WeatherTypeLightRain            WeatherType = 12
	WeatherTypeHeavyRainShowerNight WeatherType = 13
	WeatherTypeHeavyRainShowerDay   WeatherType = 14
	WeatherTypeHeavyRain            WeatherType = 15
	WeatherTypeSleetShowerNight     WeatherType = 16
	WeatherTypeSleetShowerDay       WeatherType = 17
	WeatherTypeSleet                WeatherType = 18
	WeatherTypeHailShowerNight      WeatherType = 19
	WeatherTypeHailShowerDay        WeatherType = 20
	WeatherTypeHail                 WeatherType = 21
	WeatherTypeLightSnowShowerNight WeatherType = 22
	WeatherTypeLightSnowShowerDay   WeatherType = 23
	WeatherTypeLightSnow            WeatherType = 24
	WeatherTypeHeavySnowShowerNight WeatherType = 25
	WeatherTypeHeavySnowShowerDay   WeatherType = 26
	WeatherTypeHeavySnow            WeatherType = 27
	WeatherTypeThunderShowerNight   WeatherType = 28
	WeatherTypeThunderShowerDay     WeatherType = 29
	WeatherTypeThunder              WeatherType = 30
)

type Visibility string

const (
	VisibilityUnknown   Visibility = "UN"
	VisibilityVeryPoor  Visibility = "VP"
	VisibilityPoor      Visibility = "PO"
	VisibilityModerate  Visibility = "MO"
	VisibilityGood      Visibility = "GO"
	VisibilityVeryGood  Visibility = "VG"
	VisibilityExcellent Visibility = "EX"
)
