package visuals

import (
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"strconv"
	"strings"
	"time"
	"unsafe"

	"github.com/MattSwanson/msfs2020-go/simconnect"
	rl "github.com/MattSwanson/raylib-go/raylib"
)

type Report struct {
	simconnect.RecvSimobjectDataByType
	Title                   [256]byte `name:"TITLE"`
	Kohlsman                float64   `name:"Kohlsman setting hg" unit:"inHg"`
	Altitude                float64   `name:"Indicated Altitude" unit:"feet"`
	Latitude                float64   `name:"Plane Latitude" unit:"degrees"`
	Longitude               float64   `name:"Plane Longitude" unit:"degrees"`
	WindDirection           float64   `name:"AMBIENT WIND DIRECTION" unit:"degrees"`
	WindVelocity            float64   `name:"AMBIENT WIND VELOCITY" unit:"Knots"`
	TDVelocity              float64   `name:"PLANE TOUCHDOWN NORMAL VELOCITY" unit:"feet/second"`
	GPSWPNextID             [8]byte   `name:"GPS WP NEXT ID"`
	GPSPosLat               float64   `name:"GPS POSITION LAT" unit:"degrees"`
	GPSWPETE                float64   `name:"GPS WP ETE" unit:"seconds"`
	GPSETE                  float64   `name:"GPS ETE" unit:"seconds"`
	GPSTimeZoneDeviation    float64   `name:"GPS APPROACH TIMEZONE DEVIATION" unit:"seconds"`
	GPSApproachAirport      [8]byte   `name:"GPS APPROACH AIRPORT ID"`
	GPSIsApproachActive     int64     `name:"GPS APPROACH MODE ACTIVE"`
	GPSApproachMode         float64   `name:"GPS APPROACH MODE"`
	GPSApproachTransitionID [8]byte   `name:"GPS APPROACH TRANSITION ID"`
	GPSApproachID           [8]byte   `name:"GPS APPROACH APPROACH ID"`
	GPSWPCount              float64   `name:"GPS FLIGHT PLAN WP COUNT" unit:"number"`
}

type Sett struct {
	simconnect.RecvSimobjectDataByType
	CameraState int64 `name:"CAMERA STATE" unit:"number"`
}

type AltSet struct {
	simconnect.RecvSimobjectDataByType
	Altitude float64 `name:"Plane Altitude" unit:"feet"`
}

type Events struct {
	ToggleNavLights  simconnect.DWORD
	AutoPilotOff     simconnect.DWORD
	AutoPilotOn      simconnect.DWORD
	HeadingBugSet    simconnect.DWORD
	EngineOneFailure simconnect.DWORD
}

type FlightPlan struct {
	XMLName       xml.Name `xml:"SimBase.Document"`
	DepartureID   string   `xml:"FlightPlan.FlightPlan>DepartureID"`
	DestinationID string   `xml:"FlightPlan.FlightPlan>DestinationID"`
}

func (se *Sett) SetData(s *simconnect.SimConnect) {
	defineID := s.GetDefineID(se)
	buf := [1]int64{
		se.CameraState,
	}
	size := simconnect.DWORD(8)
	s.SetDataOnSimObject(defineID, simconnect.OBJECT_ID_USER, 0, 0, size, unsafe.Pointer(&buf[0]))
}

func (se *AltSet) SetAltitude(s *simconnect.SimConnect) {
	defineID := s.GetDefineID(se)
	buf := [1]float64{
		se.Altitude,
	}
	size := simconnect.DWORD(8)
	s.SetDataOnSimObject(defineID, simconnect.OBJECT_ID_USER, 0, 0, size, unsafe.Pointer(&buf[0]))
}

var currentAlt string
var nextWP string
var fsInput chan string
var fsFont rl.Font
var wpEta string
var destEta string
var apprAirport string
var onAppoach bool
var approachMode string
var approachTransID string
var approachID string
var approachModes = map[int64]string{
	0: "None",
	1: "Transition",
	2: "Final",
	3: "Missed",
}
var hasWPCount bool
var finalWPIndex int
var finalWPID string

var destinationID string
var departureID string

func init() {
	fsInput = make(chan string)
	fmt.Println("--- FS Init complete ---")
}

func LoadFSAssets() {
	fmt.Println("--- Loading FS Assets ---")
	fsFont = rl.LoadFontEx("caskaydia.TTF", 72, nil)
	LoadFlightPlan()
}

func (r *Report) RequestData(s *simconnect.SimConnect) {
	defineID := s.GetDefineID(r)
	requestID := defineID
	s.RequestDataOnSimObjectType(requestID, defineID, 0, simconnect.SIMOBJECT_TYPE_USER)
}

func HandleFSCmd(args []string) {
	str := strings.Join(args, " ")
	fsInput <- str
}

func PollFS() error {
	s, err := simconnect.New("Request Data")
	if err != nil {
		return errors.New("couldn't connect to sim")
	}

	fmt.Println("Connected to Sim")
	report := &Report{}
	s.RegisterDataDefinition(report)
	sett := &Sett{}
	s.RegisterDataDefinition(sett)
	altset := &AltSet{}
	s.RegisterDataDefinition(altset)
	report.RequestData(s)

	events := &Events{
		ToggleNavLights:  s.GetEventID(),
		AutoPilotOff:     s.GetEventID(),
		AutoPilotOn:      s.GetEventID(),
		HeadingBugSet:    s.GetEventID(),
		EngineOneFailure: s.GetEventID(),
	}
	err = s.MapClientEventToSimEvent(events.ToggleNavLights, "TOGGLE_NAV_LIGHTS")
	if err != nil {
		panic(err)
	}
	err = s.MapClientEventToSimEvent(events.AutoPilotOff, "AUTOPILOT_OFF")
	if err != nil {
		panic(err)
	}
	err = s.MapClientEventToSimEvent(events.AutoPilotOn, "AUTOPILOT_ON")
	if err != nil {
		panic(err)
	}
	err = s.MapClientEventToSimEvent(events.HeadingBugSet, "HEADING_BUG_SET")
	if err != nil {
		panic(err)
	}
	err = s.MapClientEventToSimEvent(events.EngineOneFailure, "TOGGLE_ENGINE1_FAILURE")
	if err != nil {
		panic(err)
	}
	// se := Sett{CameraState: 5}
	// se.SetData(s)
	go func() {
	pollLoop:
		for {

			// if we have an event to send to sim, give that priority
			select {
			case in := <-fsInput:
				fmt.Println(in)
				args := strings.Fields(in)
				switch args[0] {
				case "navlights":
					err = s.TransmitClientID(events.ToggleNavLights, 0)
					if err != nil {
						fmt.Println(err)
					}
				case "camera":
					if len(args) < 2 {
						break
					}
					n, err := strconv.ParseInt(args[1], 10, 64)
					if err != nil {
						break
					}
					se := Sett{CameraState: n}
					se.SetData(s)
				case "autopilot":
					if len(args) < 2 {
						break
					}
					if args[1] == "off" {
						err = s.TransmitClientID(events.AutoPilotOff, 0)
						if err != nil {
							break
						}
					}
					if args[1] == "on" {
						err = s.TransmitClientID(events.AutoPilotOn, 0)
						if err != nil {
							break
						}
					}
				case "hdg":
					if len(args) < 2 {
						break
					}
					n, err := strconv.Atoi(args[1])
					if err != nil || (n < 0 || n > 359) {
						break
					}
					err = s.TransmitClientID(events.HeadingBugSet, simconnect.DWORD(n))
					if err != nil {
						break
					}
				case "alt":
					if len(args) < 2 {
						break
					}
					n, err := strconv.ParseFloat(args[1], 64)
					if err != nil || n < 0 {
						break
					}
					se := AltSet{Altitude: n}
					se.SetAltitude(s)
				case "eng1f":
					err = s.TransmitClientID(events.EngineOneFailure, 0)
					if err != nil {
						break
					}
				}
			default:
				ppData, r1, err := s.GetNextDispatch()
				if r1 < 0 {
					if uint32(r1) == simconnect.E_FAIL {
						continue
					}
					log.Println("Lost connection to sim??", r1, err)
					break pollLoop
				}

				recvInfo := *(*simconnect.Recv)(ppData)

				switch recvInfo.ID {
				case simconnect.RECV_ID_EXCEPTION:
					recvErr := *(*simconnect.RecvException)(ppData)
					fmt.Printf("SIMCONNECT_RECV_ID_EXCEPTION %#v\n", recvErr)
				case simconnect.RECV_ID_OPEN:
					recvOpen := *(*simconnect.RecvOpen)(ppData)
					fmt.Printf("SIMCONNECT_RECV_ID_OPEN %s\n", recvOpen.ApplicationName)
				case simconnect.RECV_ID_EVENT:
					recvEvent := *(*simconnect.RecvEvent)(ppData)
					fmt.Println("SIMCONNECT_RECV_ID_EVENT")
					switch recvEvent.EventID {
					default:
						fmt.Println("unknown SIMCONNECT_RECV_ID_EVENT", recvEvent.EventID)
					}
				case simconnect.RECV_ID_SIMOBJECT_DATA_BYTYPE:
					recvData := *(*simconnect.RecvSimobjectDataByType)(ppData)
					//fmt.Println("SIMCONNECT_RECV_SIMOBJECT_DATA_BYTYPE")
					switch recvData.RequestID {
					case s.DefineMap["Report"]:
						report := (*Report)(ppData)
						// fmt.Printf("REPORT: %s: GPS: %.6f,%.6f Altitude: %.0f Wind: %.0fkts at %.0f\n", report.Title, report.Latitude, report.Longitude, report.Altitude, report.WindVelocity, report.WindDirection)
						// fmt.Println(report.TDVelocity)
						currentAlt = fmt.Sprintf("%.0f", report.Altitude)
						nextWP = strings.Trim(string(report.GPSWPNextID[:]), string([]byte{0}))
						apprAirport = strings.Trim(string(report.GPSApproachAirport[:]), string([]byte{0}))
						//fmt.Println(apprAirport)
						calcWPETAhr := int64((report.GPSWPETE) / 60 / 60)
						calcWPETAmn := int64((report.GPSWPETE)/60) % 60
						wpEta = fmt.Sprintf("%d:%02d", calcWPETAhr, calcWPETAmn)
						calcETEhr := int64((report.GPSETE) / 60 / 60)
						calcETEmn := int64((report.GPSETE)/60) % 60
						destEta = fmt.Sprintf("%d:%02d", calcETEhr, calcETEmn)
						onAppoach = report.GPSIsApproachActive > 0
						//approachMode = approachModes[report.GPSApproachMode]
						approachTransID = strings.Trim(string(report.GPSApproachTransitionID[:]), string([]byte{0}))
						approachID = strings.Trim(string(report.GPSApproachID[:]), string([]byte{0}))
						apprAirport = strings.Trim(string(report.GPSApproachAirport[:]), string([]byte{0}))
						report.RequestData(s)
					}
				default:
					fmt.Println("recvInfo.dwID unknown", recvInfo.ID)
				}
			}

			time.Sleep(500 * time.Millisecond)

		}
		log.Println("Closing connection to sim")
		if err = s.Close(); err != nil {
			log.Println("Error closing sim connection", err)
		}
	}()

	return nil
}

func DrawFSInfo() {
	rl.DrawRectangle(0, 1340, screenWidth, 100, rl.Color{0, 0, 0, 127})
	output := ""
	if !onAppoach {
		output = fmt.Sprintf("Alt: %sft | Next WP: %s | WP ETE: %s | ETE %s: %s", currentAlt, nextWP, wpEta, destinationID, destEta)
	} else {
		output = fmt.Sprintf("Appr: %s - %s - %s - %s - %s", approachID, currentAlt, approachTransID, approachID, apprAirport)
	}
	rl.DrawTextEx(fsFont, output, rl.Vector2{100, 1349}, 72, 0, rl.SkyBlue)
}

func LoadFlightPlan() {
	// file name should be flt.pln for now
	f, err := os.Open("flt.pln")
	if err != nil {
		log.Println("Flight plan file not found")
		return
	}
	raw, err := io.ReadAll(f)
	if err != nil {
		log.Println("Couldn't read from flt plan file")
		return
	}
	fp := &FlightPlan{}
	err = xml.Unmarshal(raw, fp)
	if err != nil {
		log.Println("couldn't unmarshall xml from flt plan file")
		return
	}
	departureID = fp.DepartureID
	destinationID = fp.DestinationID
}
