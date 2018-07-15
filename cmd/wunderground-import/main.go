package main

import (
	"bufio"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"reflect"
	"strconv"
	"strings"
	"time"

	influx "github.com/influxdata/influxdb/client/v2"
	flag "github.com/spf13/pflag"
)

const flagDateFormat = "2006-01-02"
const dataDateFormat = "2006-01-02 15:04:05"

var recordNames = []string{"Time", "TemperatureC", "DewpointC", "PressurehPa", "WindDirection", "WindDirectionDegrees", "WindSpeedKMH", "WindSpeedGustKMH", "Humidity", "HourlyPrecipMM", "Conditions", "Clouds", "dailyrainMM", "SolarRadiationWatts/m^2", "SoftwareType", "DateUTC"}

func main() {
	var upload = flag.Bool("upload", false, "pass to upload data to InfluxDB, otherwise the data will be output")
	var influxAddr = flag.String("influx-addr", "http://localhost:8086", "InfluxDB HTTP address")
	var influxUser = flag.String("influx-user", "", "InfluxDB username")
	var influxPass = flag.String("influx-password", "", "InfluxDB password")
	var influxDB = flag.String("influx-db", "weather", "InfluxDB database")
	var measurementName = flag.String("measurement-name", "weather", "measurement name")
	var stationID = flag.StringP("station-id", "i", "", "Wunderground station ID")
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

	baseURL := &url.URL{
		Scheme: "https",
		Host:   "www.wunderground.com",
		Path:   "/weatherstation/WXDailyHistory.asp",
	}

	currentDate := startDate

	for currentDate.Unix() <= endDate.Unix() {
		pts, err := getPoints(*stationID, *measurementName, baseURL, &currentDate, httpC)
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

func getPoints(stationID string, measurementName string, baseURL *url.URL, date *time.Time, client *http.Client) ([]*influx.Point, error) {
	q := url.Values{
		"ID":        []string{stationID},
		"graphspan": []string{"day"},
		"format":    []string{"0"},
	}

	q.Set("day", strconv.Itoa(date.Day()))
	q.Set("month", strconv.Itoa(int(date.Month())))
	q.Set("year", strconv.Itoa(date.Year()))

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

	r := bufio.NewScanner(res.Body)

	var points []*influx.Point

	i := 0
	for r.Scan() {
		line := strings.TrimSpace(r.Text())
		if line == "" || line == "<br>" {
			continue
		}

		line = strings.TrimSuffix(line, "<br>")

		records := strings.Split(line, ",")

		if i == 0 {
			if !reflect.DeepEqual(records, recordNames) {
				return nil, fmt.Errorf("invalid first record in %s: got %#v, expected %#v", u.String(), records, recordNames)
			}

			i++
			continue
		}

		if len(records) < 16 {
			return nil, fmt.Errorf("invalid number of records in %s in record %d: %d", u.String(), i, len(records))
		}

		p, err := parseRecord(records, measurementName, map[string]string{
			"station":  stationID,
			"provider": "wunderground",
		})
		if err != nil {
			return nil, fmt.Errorf("error while parsing record %d in %s: %v", i, u.String(), err)
		}

		points = append(points, p)

		i++
	}

	return points, nil
}

func parseRecord(records []string, measurementName string, tags map[string]string) (*influx.Point, error) {
	d, err := time.ParseInLocation(dataDateFormat, records[15], time.UTC)
	if err != nil {
		return nil, fmt.Errorf("failed to parse date %q: %v", records[15], err)
	}

	tags["software"] = records[14]

	fields := make(map[string]interface{})

	fields["temperature"], err = strconv.ParseFloat(records[1], 64)
	if err != nil {
		return nil, fmt.Errorf("failed to parse temperature %q: %v", records[1], err)
	}

	fields["dewpoint"], err = strconv.ParseFloat(records[2], 64)
	if err != nil {
		return nil, fmt.Errorf("failed to parse dewpoint %q: %v", records[2], err)
	}

	fields["pressure"], err = strconv.ParseFloat(records[3], 64)
	if err != nil {
		return nil, fmt.Errorf("failed to parse pressure %q: %v", records[3], err)
	}

	fields["wind_direction_name"] = records[4]

	fields["wind_direction"], err = strconv.ParseFloat(records[5], 64)
	if err != nil {
		return nil, fmt.Errorf("failed to parse wind direction %q: %v", records[5], err)
	}

	fields["wind_speed"], err = strconv.ParseFloat(records[6], 64)
	if err != nil {
		return nil, fmt.Errorf("failed to parse wind speed %q: %v", records[6], err)
	}

	fields["wind_speed_gust"], err = strconv.ParseFloat(records[7], 64)
	if err != nil {
		return nil, fmt.Errorf("failed to parse wind speed gust %q: %v", records[7], err)
	}

	fields["humidity"], err = strconv.ParseInt(records[8], 10, 64)
	if err != nil {
		return nil, fmt.Errorf("failed to parse humidity %q: %v", records[8], err)
	}

	fields["hourly_precipitation"], err = strconv.ParseFloat(records[9], 64)
	if err != nil {
		return nil, fmt.Errorf("failed to parse hourly precipitation mm %q: %v", records[9], err)
	}

	fields["solar_radiation"], err = strconv.ParseFloat(records[13], 64)
	if err != nil {
		return nil, fmt.Errorf("failed to parse solar radiation %q: %v", records[13], err)
	}

	return influx.NewPoint(measurementName, tags, fields, d)
}
