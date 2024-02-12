package planes

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

    "github.com/MattSwanson/burtbot_overlay/visuals"
)

func init() {
	// loadAircraftTypeInfo()
    loadAircraftRegistryFromFile()
}

const (
	masterRegFileName    = "MASTER.txt"
	aircraftInfoFileName = "ACFTREF.txt"
	engineInfoFileName   = "ENGINE.txt"
)

type AircraftType struct {
	Manufacturer string
	Model        string
}

type AircraftRegistration struct {
	Nnumber      string
	FlightNumber string
	AircraftType *AircraftType
    Seen         bool
}

type ADSBAircraftInfo struct {
	Hex           string
	Type          string
	Flight        string
	AltBarometric float32     `json:"alt_baro"`
	AltGeometric  float32     `json:"alt_geom"`
	GroundSpeed   float32 `json:"gs"`
	Track         float32
	Emergency     string
	Latitude      float32 `json:"lat"`
	Longitude     float32 `json:"long"`
	RDistance     float32 `json:"r_dist"`
	RDirection    float32 `json:"r_dir"`
	Seen          float32
	RSSI          float32
}

func (ar AircraftRegistration) String() string {
	return fmt.Sprintf("%s - %s - %s - %s",
		ar.Nnumber,
        ar.FlightNumber,
		ar.AircraftType.Manufacturer,
		ar.AircraftType.Model,
	)
}

var (
	aircraftTypes    = make(map[string]AircraftType)
	aircraftRegistry = make(map[string]AircraftRegistration)
)

func loadAircraftRegistryFromFile() {
    j, err := os.ReadFile("./planes/aircraft_registry.json")
    if err != nil {
        log.Println("Couldn't open ac registry file", err.Error())
        return
    }
    if err = json.Unmarshal(j, &aircraftRegistry); err != nil {
        log.Println("Couldn't unmarshal json in ac registry file", err.Error())
    }
}

func loadAircraftTypeInfo() {
	f, err := os.Open(fmt.Sprintf("./planes/%s", aircraftInfoFileName))
	if err != nil {
		log.Println("Couldn't open aircraft info file", err.Error())
        return
	}
	scanner := bufio.NewScanner(f)
	scanner.Scan() // dump the header row
	for scanner.Scan() {
		ln := strings.Split(scanner.Text(), ",")
		ac := AircraftType{
			Manufacturer: strings.TrimSpace(ln[1]),
			Model:        strings.TrimSpace(ln[2]),
		}
		aircraftTypes[ln[0]] = ac
	}
	fmt.Println("Aircraft Info Loaded")
	f.Close()

	f, err = os.Open(fmt.Sprintf("./planes/%s", masterRegFileName))
	if err != nil {
		log.Println("Couldn't open ac reg file", err.Error())
	}
	scanner = bufio.NewScanner(f)
	scanner.Scan()
	for scanner.Scan() {
		ln := strings.Split(scanner.Text(), ",")
		acftType := aircraftTypes[ln[2]]
		reg := AircraftRegistration{
			Nnumber:      strings.TrimSpace(ln[0]),
			AircraftType: &acftType,
		}
		aircraftRegistry[strings.TrimSpace(ln[33])] = reg
	}
	fmt.Println("AC Registry Loaded")
	f.Close()
}

func ShowPlanes() {
	fmt.Println("SHOW TAH PLANWS")
	fmt.Println(aircraftTypes["00301BS"])
	fmt.Println(aircraftRegistry["A00719"])
}

func CheckForPlanes() {
	respStruct := struct {
		Now      float32
		Messages int
		Aircraft []ADSBAircraftInfo
	}{}

	req, err := http.NewRequest("GET", "http://192.168.0.30/tar1090/data/aircraft.json", nil)
	if err != nil {
		log.Println("Couldn't create request for tar1090 data", err.Error())
		return
	}
	req.Header.Set("Accept", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Println("Couldn't reach tar1090 server", err.Error())
	}

	err = json.NewDecoder(resp.Body).Decode(&respStruct)
	if err != nil {
		fmt.Println("Error decoding response from tar1090 server", err.Error())
	}

    for _, plane := range respStruct.Aircraft {
        reg, exists := aircraftRegistry[strings.ToUpper(plane.Hex)]
        if !exists {
            continue
        }

        if !reg.Seen && plane.AltBarometric != 0{
            reg.Seen = true
            reg.FlightNumber = plane.Flight
            aircraftRegistry[strings.ToUpper(plane.Hex)] = reg
            msg := visuals.MarqueeMsg{
                RawMessage: reg.String(),
            }
            json, err := json.Marshal(msg)
            if err != nil {
                log.Println("Couldn't marshall ac info to marquee", err.Error())
                continue
            }
            alt := float64(plane.AltBarometric)
            if alt > 60000 {
                alt = 60000
            }
            mPos := 1.0 - (alt / 60000.0)
            visuals.NewMarqueeWithPosition(string(json),
                mPos, 
                true)
            saveRegistry()
        }
    }
}

func resetSeenAircraft() {
    for k, v := range aircraftRegistry {
        v.Seen = false
        aircraftRegistry[k] = v
    }
    saveRegistry()
}

func saveRegistry() {
    json, err := json.Marshal(aircraftRegistry)
    if err != nil {
        log.Println("Error with Jason Marshall", err.Error())
        return
    }
    if err := os.WriteFile("./planes/aircraft_registry.json", json, 0644); err != nil {
        log.Println("Error writing registry to disk", err.Error())
    }
}
