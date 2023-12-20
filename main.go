package main

import (
	"fmt"
	"log"
	"math"
	"os"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/gocarina/gocsv"
	"github.com/putridparrot/GoUnits/speed"
	"github.com/pymaxion/geographiclib-go/geodesic"
	geo "github.com/rbsns/golang-geo"
	"github.com/scylladb/termtables"
	"github.com/vjeantet/jodaTime"
)

// Version
const Version = "0.0.3"

// Constans
const DM = 1.828   //https://en.wikipedia.org/wiki/Data_mile Km
const DMM = 1828.0 //Meters

// http://wikimapia.org/25820161/it/Centro-Radar-Poggio-Ballone
// const lat_RadarPB = 42.82638889  // Coordinate:   42°49'35"N   10°53'3"E
// const long_RadarPB = 10.88416667 //
// New coordinate Poggio Ballone calcolate usando Maps ed identificando proprio il RADAR
const lat_RadarPB = 42.828997
const long_RadarPB = 10.880370

const lat_RadarMA = 37.827630 //Marsala
const long_RadarMA = 12.537120

/*
43.36719444
13.67494444

43.366521, 13.673615
Potenza Picena,62018 MC

9M8F+JC5 Potenza Picena, Provincia di Macerata
*/
const lat_RadarPP = 43.366521
const long_RadarPP = 13.673615

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
	RADAR string `csv:"RADAR"`
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
	return fmt.Sprintf(current_time.Format("021504ZJan06"))
}

// https://ispycode.com/GO/Math/Metric-Conversions/Distance/Feet-to-meters
func feet2meters(feet float64) float64 {
	return feet * 0.3048
}

// https://siongui.github.io/2018/02/25/go-get-file-name-without-extension/
func FilenameWithoutExtension(fn string) string {
	return strings.TrimSuffix(fn, path.Ext(fn))
}

func main() {
	//Load Args
	argsWithoutProg := os.Args[1:]
	//<cmd> LL464.csv P|M F104 1000105
	//<cmd> LL464.csv P|M F104 1000105 "TF-104G Bergamini-Moretti"
	//P = Poggio Ballone
	//M = Marsala
	//PP = Potenza Picena
	//LoadCSV
	//fmt.Println(argsWithoutProg[4]) //DEBUG
	csvFile, err := os.OpenFile("data/"+argsWithoutProg[0], os.O_RDWR|os.O_CREATE, os.ModePerm)
	if err != nil {
		panic(err)
	}
	defer csvFile.Close()

	targets := []*TargetCSV{}

	if err := gocsv.UnmarshalFile(csvFile, &targets); err != nil { // Load targets from file
		panic(err)
	}

	table := termtables.CreateTable()
	table.AddHeaders("TIME", "X", "Y", "Distance from Radar", "Bearing to Radar", "Bearing", "Lat", "Long", "Lat V", "Long V")
	//Load targets from file and add to list //DEBUG
	for _, target := range targets {
		s_X, _ := strconv.ParseFloat(target.X, 64)
		s_Y, _ := strconv.ParseFloat(target.Y, 64)
		distance := (math.Sqrt(math.Pow(math.Abs(s_X), 2) + math.Pow(math.Abs(s_Y), 2))) * DM
		distance_m := (math.Sqrt(math.Pow(math.Abs(s_X), 2) + math.Pow(math.Abs(s_Y), 2))) * DMM
		var bearing_s float64 = 0.0
		var bearing float64 = 0.0
		var sqrt float64 = 0.0
		sqrt = (math.Sqrt(math.Pow(math.Abs(s_X), 2) + math.Pow(math.Abs(s_Y), 2)))
		//fmt.Println(Float32ToString(s_Y)) //DEBUG
		//bearingPB_s = (90.0 - (math.Atan(math.Abs(s_Y)/math.Abs(s_X)) * 180 / math.Pi))
		//bearingPB_s = (math.Atan(math.Abs(s_Y)/math.Abs(s_X)) * 180 / math.Pi)
		//fmt.Println(math.Acos(math.Abs(s_Y) / sqrt)) //DEBUG
		if (math.Signbit(s_X) == true) && (math.Signbit(s_Y) == false) {
			//- +
			bearing_s = ((math.Acos(math.Abs(s_X) / sqrt)) * 180 / math.Pi)
			bearing = bearing_s + 270.0
		} else if (math.Signbit(s_X) == true) && (math.Signbit(s_Y) == true) {
			// - -
			bearing_s = ((math.Acos(math.Abs(s_Y) / sqrt)) * 180 / math.Pi)
			bearing = bearing_s + 180.0
		} else if (math.Signbit(s_X) == false) && (math.Signbit(s_Y) == true) {
			//+ -
			bearing_s = ((math.Acos(math.Abs(s_X) / sqrt)) * 180 / math.Pi)
			bearing = bearing_s + 90.0
		} else if (math.Signbit(s_X) == false) && (math.Signbit(s_Y) == false) {
			//+ +
			bearing_s = ((math.Acos(math.Abs(s_Y) / sqrt)) * 180 / math.Pi)
			bearing = bearing_s
		}

		///Choose Radar
		var lat_Radar, long_Radar float64
		if target.RADAR == "P" {
			lat_Radar = lat_RadarPB
			long_Radar = long_RadarPB
		}
		if target.RADAR == "M" {
			lat_Radar = lat_RadarMA
			long_Radar = long_RadarMA
		}
		if target.RADAR == "PP" {
			lat_Radar = lat_RadarPP
			long_Radar = long_RadarPP
		}

		//New Lat, Long Position
		p_radar := geo.NewPoint(lat_Radar, long_Radar)
		new_p := p_radar.PointAtDistanceAndBearing(distance, bearing)

		//Vincenty
		r := geodesic.WGS84.Direct(lat_Radar, long_Radar, bearing, distance_m)
		//dmsCoordinate, err := New(LatLon{Latitude: new_p.Lat(), Longitude: new_p.Lng()})
		if err != nil {
			log.Fatal(err)
		}
		table.AddRow(target.TIME, Float64ToString(s_X), Float64ToString(s_Y), fmt.Sprintf("%.2f", distance), fmt.Sprintf("%.2f", bearing), Float64ToString(bearing_s), Float64ToString(new_p.Lat()), Float64ToString(new_p.Lng()), Float64ToString(r.Lat2), Float64ToString(r.Lon2))
	}
	fmt.Println(table.Render())

	//Create and save acmi file (TacView)
	BOF := "FileType=text/acmi/tacview\nFileVersion=2.2\n"
	GIOF := "0,Author=Enrico Speranza\n0,Title=Radar activity near ITAVIA I-TIGI IH870 A1136\n0,ReferenceTime=1980-06-27T18:00:00Z\n"
	//Open with name
	f, err := os.Create("out/nearadaractivity19800627180000Z" + FilenameWithoutExtension(argsWithoutProg[0]) + "v" + DDHHMMZmmmYY() + ".acmi")
	if err != nil {
		panic(err)
	}
	defer f.Close()
	_, _ = f.WriteString(BOF)
	_, _ = f.WriteString(GIOF)
	//Create data
	dateTimeST, _ := jodaTime.Parse("HHmmss", ST)
	//fmt.Println(dateTimeST) //DEBUG
	var strTimeToWrite string
	var sumDuration int32
	for _, target := range targets {
		s_X, _ := strconv.ParseFloat(target.X, 64)
		s_Y, _ := strconv.ParseFloat(target.Y, 64)
		s_ALT, _ := strconv.ParseFloat(target.ALT, 64)
		s_NTN := target.NTN
		s_ALT = feet2meters(s_ALT)                                                           //Feet To Meters (ASL)
		s_SPEED := speed.Knots.ToMetresPerSecond(speed.Knots(StringToFloat64(target.SPEED))) //IAS Indicated airspeed (m/s)
		//distancePB := (math.Sqrt(math.Pow(math.Abs(s_X), 2) + math.Pow(math.Abs(s_Y), 2))) * DM
		distance_m := (math.Sqrt(math.Pow(math.Abs(s_X), 2) + math.Pow(math.Abs(s_Y), 2))) * DMM
		var bearing_s float64 = 0.0
		var bearing float64 = 0.0
		var sqrt float64 = 0.0
		sqrt = (math.Sqrt(math.Pow(math.Abs(s_X), 2) + math.Pow(math.Abs(s_Y), 2)))
		//fmt.Println(Float32ToString(s_Y)) //DEBUG
		//bearingPB_s = ((math.Acos(math.Abs(s_Y) / sqrt)) * 180 / math.Pi)
		//bearingPB_s = (90.0 - (math.Atan(math.Abs(s_Y)/math.Abs(s_X)) * 180 / math.Pi))
		//bearingPB_s = (math.Atan(math.Abs(s_Y)/math.Abs(s_X)) * 180 / math.Pi)
		//fmt.Println(math.Acos(math.Abs(s_Y) / sqrt)) //DEBUG

		//WARNING: could be: 360-tan^-1(85.20/94.31)*180/3.1415 NO, doesn't works!
		if (math.Signbit(s_X) == true) && (math.Signbit(s_Y) == false) {
			// -X +Y
			bearing_s = ((math.Acos(math.Abs(s_X) / sqrt)) * 180 / math.Pi)
			bearing = bearing_s + 270.0
		} else if (math.Signbit(s_X) == true) && (math.Signbit(s_Y) == true) {
			// -X -Y
			bearing_s = ((math.Acos(math.Abs(s_Y) / sqrt)) * 180 / math.Pi)
			bearing = bearing_s + 180.0
		} else if (math.Signbit(s_X) == false) && (math.Signbit(s_Y) == true) {
			// +X -Y
			bearing_s = ((math.Acos(math.Abs(s_X) / sqrt)) * 180 / math.Pi)
			bearing = bearing_s + 90.0
		} else if (math.Signbit(s_X) == false) && (math.Signbit(s_Y) == false) {
			// +X +Y
			bearing_s = ((math.Acos(math.Abs(s_Y) / sqrt)) * 180 / math.Pi)
			bearing = bearing_s
		}
		//bearingPB = bearingPB - 1.0 //Correction?
		//bearingPB := ((90.0 - (math.Atan(math.Abs(s_Y)/math.Abs(s_X)) * 180 / math.Pi)) + 180.0)
		//New Lat, Long Position
		//p_radarPB := geo.NewPoint(lat_RadarPB, long_RadarPB)
		//new_p := p_radarPB.PointAtDistanceAndBearing(distancePB, bearingPB)

		//Choose Radar
		var lat_Radar, long_Radar float64
		if target.RADAR == "P" {
			lat_Radar = lat_RadarPB
			long_Radar = long_RadarPB
		}
		if target.RADAR == "M" {
			lat_Radar = lat_RadarMA
			long_Radar = long_RadarMA
		}
		if target.RADAR == "PP" {
			lat_Radar = lat_RadarPP
			long_Radar = long_RadarPP
		}

		//Vincenty
		r := geodesic.WGS84.Direct(lat_Radar, long_Radar, bearing, distance_m)
		//Time Next
		dateTimeNow, _ := jodaTime.Parse("HHmmss", target.TIME) //Read TIME from CSV
		if dateTimeNow.After(dateTimeST) {
			sumDuration = sumDuration + int32(dateTimeNow.Sub(dateTimeST).Seconds()) //TODO: Ricontrollare tutti i tempi!
			strTimeToWrite = fmt.Sprintf("#%s.%s\n", IntToString(int(sumDuration)), "00")
			dateTimeST = dateTimeNow
			//strTimeToWrite = fmt.Sprintf("#%s.%s\n", Float64ToTimeString(dateTimeNow.Sub(dateTimeST).Minutes()), Float64ToTimeString(dateTimeNow.Sub(dateTimeST).Seconds()))
		}
		_, _ = f.WriteString(strTimeToWrite)
		//Coodinates
		strToWrite := fmt.Sprintf("%s,T=%s|%s|%s,IAS=%s,Name=%s,Squawk=%s,Label=%s\n",
			argsWithoutProg[3],
			Float64ToString(r.Lon2),
			Float64ToString(r.Lat2),
			Float64ToString(s_ALT),
			Float64ToString(s_SPEED),
			argsWithoutProg[2],
			s_NTN,
			argsWithoutProg[4])
		_, _ = f.WriteString(strToWrite)
	}

	f.Sync()
}
