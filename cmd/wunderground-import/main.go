package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"time"

	influx "github.com/influxdata/influxdb/client/v2"
	flag "github.com/spf13/pflag"
)

const flagDateFormat = "2006-01-02"

// API in use: https://docs.google.com/document/d/1w8jbqfAk0tfZS5P7hYnar1JiitM0gQZB-clxDfG3aD0/edit
func main() {
	var upload = flag.Bool("upload", false, "pass to upload data to InfluxDB, otherwise the data will be output")
	var influxAddr = flag.String("influx-addr", "http://localhost:8086", "InfluxDB HTTP address")
	var influxUser = flag.String("influx-user", "", "InfluxDB username")
	var influxPass = flag.String("influx-password", "", "InfluxDB password")
	var influxDB = flag.String("influx-db", "weather", "InfluxDB database")
	var measurementName = flag.String("measurement-name", "weather", "measurement name")
	var stationID = flag.StringP("station-id", "i", "", "Wunderground station ID")
	var apiKey = flag.String("api-key", "", "Wunderground API key")
	var startDateFlag = flag.StringP("start-date", "s", "", "start date (yyyy-mm-dd), default is yesterday")
	var endDateFlag = flag.StringP("end-date", "e", "", "end date (yyyy-mm-dd), default is today")

	flag.Parse()

	if *startDateFlag == "" {
		*startDateFlag = time.Now().AddDate(0, 0, -1).Format(flagDateFormat)
	}

	if *endDateFlag == "" {
		*endDateFlag = time.Now().Format(flagDateFormat)
	}

	startDate, err := time.ParseInLocation(flagDateFormat, *startDateFlag, time.UTC)
	if err != nil {
		log.Fatal(err)
	}

	endDate, err := time.ParseInLocation(flagDateFormat, *endDateFlag, time.UTC)
	if err != nil {
		log.Fatal(err)
	}

	if *stationID == "" {
		flag.Usage()
		log.Fatal("please specify a station ID")
	}

	httpC := &http.Client{}

	// Create a new point batch
	bp, err := influx.NewBatchPoints(influx.BatchPointsConfig{
		Database:  *influxDB,
		Precision: "s",
	})
	if err != nil {
		log.Fatal(err)
	}

	baseQuery := url.Values{
		"stationId": []string{*stationID},
		"apiKey":    []string{*apiKey},
		"format":    []string{"json"},
		"units":     []string{"m"},
	}

	baseURL := &url.URL{
		Scheme: "https",
		Host:   "api.weather.com",
		Path:   "/v2/pws/history/all",
	}

	currentDate := startDate

	for currentDate.Unix() <= endDate.Unix() {
		pts, err := getPoints(*measurementName, baseURL, baseQuery, &currentDate, httpC)
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed to get point for %s: %v\n", currentDate, err)
		}

		bp.AddPoints(pts)

		currentDate = currentDate.AddDate(0, 0, 1)
	}

	if *upload {
		// Create a new HTTPClient
		c, err := influx.NewHTTPClient(influx.HTTPConfig{
			Addr:     *influxAddr,
			Username: *influxUser,
			Password: *influxPass,
		})
		if err != nil {
			log.Fatal(err)
		}
		defer c.Close()

		if _, _, err := c.Ping(1 * time.Second); err != nil {
			log.Fatal(err)
		}

		if err := c.Write(bp); err != nil {
			log.Fatal(err)
		}

		// Close client resources
		if err := c.Close(); err != nil {
			log.Fatal(err)
		}

		return
	}

	for _, p := range bp.Points() {
		if p == nil {
			continue
		}
		fmt.Println(p.PrecisionString("ns"))
	}
}

func getPoints(measurementName string, baseURL *url.URL, baseQuery url.Values, date *time.Time, client *http.Client) ([]*influx.Point, error) {
	q := baseQuery
	q.Set("date", date.Format("20060102"))

	u := baseURL.ResolveReference(&url.URL{
		RawQuery: q.Encode(),
	})

	res, err := client.Get(u.String())
	if err != nil {
		return nil, fmt.Errorf("failed to get %s: %v", u.String(), err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get %s: status code %d %s", u.String(), res.StatusCode, res.Status)
	}

	response := &APIResponse{}
	if err := json.NewDecoder(res.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode API response: %v", err)
	}

	var points []*influx.Point

	for _, observation := range response.Observations {
		p, err := observationToPoint(observation, measurementName, map[string]string{
			"station":  observation.StationID,
			"provider": "wunderground",
		})
		if err != nil {
			return nil, fmt.Errorf("error while converting observation %+v: %v", observation, err)
		}

		points = append(points, p)
	}

	return points, nil
}

func observationToPoint(observation *Observation, measurementName string, tags map[string]string) (*influx.Point, error) {
	fields := make(map[string]interface{})

	// Legacy fields
	fields["temperature"] = observation.Metric.TemperatureAverage
	fields["dewpoint"] = observation.Metric.DewpointAverage
	fields["pressure"] = (observation.Metric.PressureMaximum + observation.Metric.PressureMinimum) / 2.0
	fields["wind_direction"] = observation.WindDirectionAverage
	fields["wind_speed"] = observation.Metric.WindspeedAverage
	fields["wind_speed_gust"] = observation.Metric.WindgustAverage
	fields["humidity"] = int64(observation.HumidityAverage)
	fields["solar_radiation"] = observation.SolarRadiationHigh

	fields["solar_radiation_high"] = observation.SolarRadiationHigh
	fields["uv_high"] = observation.UVHigh
	fields["wind_direction_average"] = observation.WindDirectionAverage
	fields["humidity_high"] = observation.HumidityHigh
	fields["humidity_low"] = observation.HumidityLow
	fields["humidity_average"] = observation.HumidityAverage

	fields["temperature_high"] = observation.Metric.TemperatureHigh
	fields["temperature_low"] = observation.Metric.TemperatureLow
	fields["temperature_average"] = observation.Metric.TemperatureAverage

	fields["wind_speed_high"] = observation.Metric.WindspeedHigh
	fields["wind_speed_low"] = observation.Metric.WindspeedLow
	fields["wind_speed_average"] = observation.Metric.WindspeedAverage

	fields["wind_speed_gust_high"] = observation.Metric.WindgustHigh
	fields["wind_speed_gust_low"] = observation.Metric.WindgustLow
	fields["wind_speed_gust_average"] = observation.Metric.WindgustAverage

	fields["dewpoint_high"] = observation.Metric.DewpointHigh
	fields["dewpoint_low"] = observation.Metric.DewpointLow
	fields["dewpoint_average"] = observation.Metric.DewpointAverage

	fields["windchill_high"] = observation.Metric.WindchillHigh
	fields["windchill_low"] = observation.Metric.WindchillLow
	fields["windchill_average"] = observation.Metric.WindchillAverage

	fields["heat_index_high"] = observation.Metric.HeatIndexHigh
	fields["heat_index_low"] = observation.Metric.HeatIndexLow
	fields["heat_index_average"] = observation.Metric.HeatIndexAverage

	fields["pressure_maximum"] = observation.Metric.PressureMaximum
	fields["pressure_minimum"] = observation.Metric.PressureMinimum
	fields["pressure_trend"] = observation.Metric.PressureTrend

	fields["precipitation_rate"] = observation.Metric.PrecipitationRate
	fields["precipitation_total"] = observation.Metric.PrecipitationTotal

	return influx.NewPoint(measurementName, tags, fields, observation.ObservationTimeUTC)
}
