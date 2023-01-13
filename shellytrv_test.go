package shelly

import (
	"encoding/json"
	"fmt"
	"testing"
)

const ShellyTRVInfoJSON = `
{
    "thermostats": [
        {
            "pos": -100.0,
            "target_t": {
                "enabled": true,
                "value": 31.0,
                "units": "C"
            },
            "tmp": {
                "value": 17.4,
                "units": "C",
                "is_valid": true
            },
            "schedule": false,
            "schedule_profile": 2,
            "boost_minutes": 0
        }
    ],
    "calibrated": false,
    "bat": {
        "value": 0,
        "voltage": 3.127
    },
    "charger": true,
    "ps_mode": 0,
    "dbg_flags": 0
}
`

func TestShellyTRVDatumFromJSON(t *testing.T) {
	sd := ShellyTRVInfo{}
	err := json.Unmarshal([]byte(ShellyTRVInfoJSON), &sd)
	if err != nil {
		t.Errorf("%s", err)
	}
	fmt.Printf("sd: %+v", sd)
}
