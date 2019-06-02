package main

import "time"

type APIResponse struct {
	Observations []*Observation `json:"observations"`
}

type Observation struct {
	StationID            string            `json:"stationID"`
	Timezone             string            `json:"tz"`
	ObservationTimeUTC   time.Time         `json:"obsTimeUtc"`
	ObservationTimeLocal string            `json:"obsTimeLocal"`
	ObservationEpoch     int               `json:"epoch"`
	Latitude             float64           `json:"lat"`
	Longitude            float64           `json:"lon"`
	SolarRadiationHigh   float64           `json:"solarRadiationHigh"`
	UVHigh               float64           `json:"uvHigh"`
	WindDirectionAverage int               `json:"winddirAvg"`
	HumidityHigh         int               `json:"humidityHigh"`
	HumidityLow          int               `json:"humidityLow"`
	HumidityAverage      int               `json:"humidityAvg"`
	QCStatus             int               `json:"qcStatus"`
	Metric               MetricObservation `json:"metric"`
}

type MetricObservation struct {
	TemperatureHigh    int     `json:"tempHigh"`
	TemperatureLow     int     `json:"tempLow"`
	TemperatureAverage int     `json:"tempAvg"`

	WindspeedHigh      int     `json:"windspeedHigh"`
	WindspeedLow       int     `json:"windspeedLow"`
	WindspeedAverage   int     `json:"windspeedAvg"`

	WindgustHigh       int     `json:"windgustHigh"`
	WindgustLow        int     `json:"windgustLow"`
	WindgustAverage    int     `json:"windgustAvg"`

	DewpointHigh       int     `json:"dewptHigh"`
	DewpointLow        int     `json:"dewptLow"`
	DewpointAverage    int     `json:"dewptAvg"`

	WindchillHigh      int     `json:"windchillHigh"`
	WindchillLow       int     `json:"windchillLow"`
	WindchillAverage   int     `json:"windchillAvg"`

	HeatIndexHigh      int     `json:"heatindexHigh"`
	HeatIndexLow       int     `json:"heatindexLow"`
	HeatIndexAverage   int     `json:"heatindexAvg"`

	PressureMaximum    float64 `json:"pressureMax"`
	PressureMinimum    float64 `json:"pressureMin"`
	PressureTrend      float64 `json:"pressureTrend"`

	PrecipitationRate  float64 `json:"precipRate"`
	PrecipitationTotal float64 `json:"precipTotal"`
}

