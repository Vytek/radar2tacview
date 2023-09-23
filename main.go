package main

import (
	"fmt"
	"log"
	"math"
	"os"
	"strconv"
	"time"

	"github.com/gocarina/gocsv"
	geo "github.com/rbsns/golang-geo"
	"github.com/scylladb/termtables"
	"github.com/vjeantet/jodaTime"
)

// Version
const Version = "0.0.1"

// Constans
const DM = 1.828 //https://en.wikipedia.org/wiki/Data_mile Km
// http://wikimapia.org/25820161/it/Centro-Radar-Poggio-Ballone
const lat_RadarPB = 42.82638889  // Coordinate:   42°49'35"N   10°53'3"E
const long_RadarPB = 10.88416667 //
// Start time
const ST = "180000"

type TargetCSV struct { // Our example struct, you can use "-" to ignore a field
	TIME  string `csv:"TIME"`
	NTN   string `csv:"NTN"`
	ENT   string `csv:"ENT"`
	X     string `csv:"X"`
	Y     string `csv:"Y"`
	SPEED string `csv:"SPEED"`
	BEAR  string `csv:"BEAR"`
	ALT   string `csv:"ALT"`
}

func StringToInt(data string) int {
	n, _ := strconv.Atoi(data)
	return n
}

func IntToString(n int) string {
	return strconv.Itoa(n)
}

func Float64ToString(f float64) string {
	/** converting the f variable into a string */
	/** 5 is the number of decimals */
	/** 64 is for float64 type*/
	return strconv.FormatFloat(f, 'f', 5, 64)
}

func Float64ToTimeString(f float64) string {
	/** converting the f variable into a string */
	/** 5 is the number of decimals */
	/** 64 is for float64 type*/
	return strconv.FormatFloat(f, 'f', 0, 64)
}

func Float32ToString(f float64) string {
	/** converting the f variable into a string */
	/** 5 is the number of decimals */
	/** 32 is for float32 type*/
	return strconv.FormatFloat(f, 'f', 5, 32)
}

func StringToFloat32(data string) float32 {
	if s, err := strconv.ParseFloat(data, 32); err == nil {
		return float32(s)
	} else {
		return 0.0
	}
}

func StringToFloat64(data string) float64 {
	if s, err := strconv.ParseFloat(data, 64); err == nil {
		return s
	} else {
		return 0.0
	}
}

// https://stackoverflow.com/a/37247762
func round(val float64) int {
	if val < 0 {
		return int(val - 0.5)
	}
	return int(val + 0.5)
}

// https://go.dev/play/p/Q-Ufgrw3vZL
func DDHHMMZ() string {
	current_time := time.Now().UTC()
	return fmt.Sprintf("%d%02d%02dZ\n", current_time.Day(), current_time.Hour(), current_time.Minute())
}

func DDHHMMZmmmYY() string {
	current_time := time.Now().UTC()
	return fmt.Sprintf(current_time.Format("021504ZJan06\n"))
}

// https://ispycode.com/GO/Math/Metric-Conversions/Distance/Feet-to-meters
func feet2meters(feet float64) float64 {
	return feet * 0.3048
}

func main() {
	//LoadCSV
	csvFile, err := os.OpenFile("data/AJ024.csv", os.O_RDWR|os.O_CREATE, os.ModePerm)
	if err != nil {
		panic(err)
	}
	defer csvFile.Close()

	targets := []*TargetCSV{}

	if err := gocsv.UnmarshalFile(csvFile, &targets); err != nil { // Load targets from file
		panic(err)
	}

	table := termtables.CreateTable()
	table.AddHeaders("TIME", "X", "Y", "Distance from PB Radar", "Bearing to PB Radar", "Lat", "Long", "Lat °/Long °")
	//Load targets from file and add to list //DEBUG
	for _, target := range targets {
		s_X, _ := strconv.ParseFloat(target.X, 64)
		s_Y, _ := strconv.ParseFloat(target.Y, 64)
		distancePB := (math.Sqrt(math.Pow(math.Abs(s_X), 2) + math.Pow(math.Abs(s_Y), 2))) * DM
		bearingPB := ((90.0 - (math.Atan(math.Abs(s_Y)/math.Abs(s_X)) * 180 / math.Pi)) + 180.0)
		//New Lat, Long Position
		p_radarPB := geo.NewPoint(lat_RadarPB, long_RadarPB)
		new_p := p_radarPB.PointAtDistanceAndBearing(distancePB, bearingPB)
		dmsCoordinate, err := New(LatLon{Latitude: new_p.Lat(), Longitude: new_p.Lng()})
		if err != nil {
			log.Fatal(err)
		}
		table.AddRow(target.TIME, fmt.Sprintf("%.2f", s_X), fmt.Sprintf("%.2f", s_Y), fmt.Sprintf("%.2f", distancePB), fmt.Sprintf("%.2f", bearingPB), new_p.Lat(), new_p.Lng(), dmsCoordinate.String())
	}
	fmt.Println(table.Render())

	//Create and save acmi file (TacView)
	BOF := "FileType=text/acmi/tacview\nFileVersion=2.2\n"
	GIOF := "0,Author=Enrico Speranza\n0,Title=Radar activity near ITAVIA I-TIGI IH870 A1136\n0,ReferenceTime=1980-06-27T18:00:00Z\n"
	f, err := os.Create("data/nearadaractivity19800627180000Z.acmi")
	if err != nil {
		panic(err)
	}
	defer f.Close()
	_, _ = f.WriteString(BOF)
	_, _ = f.WriteString(GIOF)
	//Create data
	dateTimeST, _ := jodaTime.Parse("HHmmss", ST)
	fmt.Println(dateTimeST)
	var strTimeToWrite string
	for _, target := range targets {
		s_X, _ := strconv.ParseFloat(target.X, 64)
		s_Y, _ := strconv.ParseFloat(target.Y, 64)
		s_ALT, _ := strconv.ParseFloat(target.ALT, 64)
		s_ALT = feet2meters(s_ALT) //Feet To Meters (ASL)
		distancePB := (math.Sqrt(math.Pow(math.Abs(s_X), 2) + math.Pow(math.Abs(s_Y), 2))) * DM
		bearingPB := ((90.0 - (math.Atan(math.Abs(s_Y)/math.Abs(s_X)) * 180 / math.Pi)) + 180.0)
		//New Lat, Long Position
		p_radarPB := geo.NewPoint(lat_RadarPB, long_RadarPB)
		new_p := p_radarPB.PointAtDistanceAndBearing(distancePB, bearingPB)
		if err != nil {
			log.Fatal(err)
		}
		//Time Next
		dateTimeNow, _ := jodaTime.Parse("HHmmss", target.TIME)
		if dateTimeNow.After(dateTimeST) {
			strTimeToWrite = fmt.Sprintf("#%s.%s\n", Float64ToTimeString(dateTimeNow.Sub(dateTimeST).Minutes()), Float64ToTimeString(dateTimeNow.Sub(dateTimeST).Seconds()))
			dateTimeST = dateTimeNow
		} else {
			strTimeToWrite = fmt.Sprintf("#%s.%s\n", Float64ToTimeString(dateTimeNow.Sub(dateTimeST).Minutes()), Float64ToTimeString(dateTimeNow.Sub(dateTimeST).Seconds()))
		}
		_, _ = f.WriteString(strTimeToWrite)
		//Coodinates
		strToWrite := fmt.Sprintf("1000102,T=%s|%s|%s,Name=AJ024\n", Float64ToString(new_p.Lat()), Float64ToString(new_p.Lng()), Float64ToString(s_ALT))
		_, _ = f.WriteString(strToWrite)
	}

	f.Sync()
}
