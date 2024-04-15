package main 

import (
	"io/fs"
	"fmt"
	"path/filepath"
	"os"
	"encoding/json"
	"strings"
	"runtime"

	"github.com/klauspost/compress/zstd"
)

var decoder, _ = zstd.NewReader(nil, zstd.WithDecoderConcurrency(0))
var resourcesFS fs.StatFS = getResourcesFS()

type FleetAircraft struct {
	ICAO  string
	Count int
}

type Airlines struct {
	ICAO     string `json:"icao"`
	Name     string `json:"name"`
	Callsign struct {
		Name            string   `json:"name"`
		CallsignFormats []string `json:"callsignFormats"`
	} `json:"callsign"`
	JSONFleets map[string][][2]interface{} `json:"fleets"`
	Fleets     map[string][]FleetAircraft
}

func LoadResource(path string) []byte {
	b, err := os.ReadFile("resources/openscope-airlines.json")
	if err != nil {
		panic(err)
	}

	return b
}

func decompressZstd(s string) string {
	b, err := decoder.DecodeAll([]byte(s), nil)
	if err != nil {
		fmt.Printf("Error decompressing buffer")
	}
	return string(b)
}

func parseAirlines() (map[string]Airlines, map[string]string) {
	openscopeAirlines := LoadResource("openscope-airlines.json")

	var alStruct struct {
		Airlines []Airlines `json:"airlines"`
	}
	if err := json.Unmarshal([]byte(openscopeAirlines), &alStruct); err != nil {
		fmt.Printf("error in JSON unmarshal of openscope-airlines: %v", err)
	}

	airlines := make(map[string]Airlines)
	callsigns := make(map[string]string)
	for _, al := range alStruct.Airlines {
		fixedAirline := al
		fixedAirline.Fleets = make(map[string][]FleetAircraft)
		for name, aircraft := range fixedAirline.JSONFleets {
			for _, ac := range aircraft {
				fleetAC := FleetAircraft{
					ICAO:  strings.ToUpper(ac[0].(string)),
					Count: int(ac[1].(float64)),
				}
				fixedAirline.Fleets[name] = append(fixedAirline.Fleets[name], fleetAC)
			}
		}
		fixedAirline.JSONFleets = nil

		airlines[strings.ToUpper(al.ICAO)] = fixedAirline
		callsigns[strings.ToUpper(al.ICAO)] = al.Callsign.Name
	}
	return airlines, callsigns
}

func getResourcesFS() fs.StatFS {
	path, err := os.Executable()
	if err != nil {
		panic(err)
	}

	dir := filepath.Dir(path)
	if runtime.GOOS == "darwin" {
		dir = filepath.Clean(filepath.Join(dir, "..", "Resources"))
	} else {
		dir = filepath.Join(dir, "resources")
	}

	fsys, ok := os.DirFS(dir).(fs.StatFS)
	if !ok {
		panic("FS from DirFS is not a StatFS?")
	}

	check := func(fs fs.StatFS) bool {
		_, errv := fsys.Stat("videomaps")
		_, errs := fsys.Stat("scenarios")
		return errv == nil && errs == nil
	}

	return fsys

	if check(fsys) {
		return fsys
	}

	// Try CWD (this is useful for development and debugging but shouldn't
	// be needed for release builds.

	wd, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	dir = filepath.Join(wd, "resources")

	fsys, ok = os.DirFS(dir).(fs.StatFS)
	if !ok {
		panic("FS from DirFS is not a StatFS?")
	}

	if check(fsys) {
		return fsys
	}
	panic("unable to find videomaps in CWD")
}

