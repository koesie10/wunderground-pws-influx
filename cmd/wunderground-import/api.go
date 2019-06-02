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
	WindDirectionAverage float64           `json:"winddirAvg"`
	HumidityHigh         float64           `json:"humidityHigh"`
	HumidityLow          float64           `json:"humidityLow"`
	HumidityAverage      float64           `json:"humidityAvg"`
	QCStatus             int               `json:"qcStatus"`
	Metric               MetricObservation `json:"metric"`
}

type MetricObservation struct {
	TemperatureHigh    float64 `json:"tempHigh"`
	TemperatureLow     float64 `json:"tempLow"`
	TemperatureAverage float64 `json:"tempAvg"`

	WindspeedHigh    float64 `json:"windspeedHigh"`
	WindspeedLow     float64 `json:"windspeedLow"`
	WindspeedAverage float64 `json:"windspeedAvg"`

	WindgustHigh    float64 `json:"windgustHigh"`
	WindgustLow     float64 `json:"windgustLow"`
	WindgustAverage float64 `json:"windgustAvg"`

	DewpointHigh    float64 `json:"dewptHigh"`
	DewpointLow     float64 `json:"dewptLow"`
	DewpointAverage float64 `json:"dewptAvg"`

	WindchillHigh    float64 `json:"windchillHigh"`
	WindchillLow     float64 `json:"windchillLow"`
	WindchillAverage float64 `json:"windchillAvg"`

	HeatIndexHigh    float64 `json:"heatindexHigh"`
	HeatIndexLow     float64 `json:"heatindexLow"`
	HeatIndexAverage float64 `json:"heatindexAvg"`

	PressureMaximum float64 `json:"pressureMax"`
	PressureMinimum float64 `json:"pressureMin"`
	PressureTrend   float64 `json:"pressureTrend"`

	PrecipitationRate  float64 `json:"precipRate"`
	PrecipitationTotal float64 `json:"precipTotal"`
}
