package metro_tenerife

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type Panels []struct {
	Service                    string `json:"service"`
	Stop                       string `json:"stop"`
	StopSAE                    int    `json:"stopSAE"`
	DestinationStop            string `json:"destinationStop"`
	StopDescription            string `json:"stopDescription"`
	DestinationStopDescription string `json:"destinationStopDescription"`
	Route                      int    `json:"route"`
	Direction                  int    `json:"direction"`
	LastUpdate                 int64  `json:"lastUpdate"`
	LastUpdateFormatted        string `json:"lastUpdateFormatted"`
	RemainingMinutes           int    `json:"remainingMinutes"`
	OrderStop                  int    `json:"orderStop"`
}

type Trams []struct {
	Vehicle int    `json:"VEHICULO"`
	Date    string `json:"FECHA"`
	Service int    `json:"SERVICIO"`
	Load    int    `json:"CARGA"`
}

type Stop struct {
	Id   string
	Name string
}

type PanelTram struct {
	Stop                       string
	StopDescription            string
	DestinationStopDescription string
	RemainingMinutes           int
	Load                       int
}

func BaseURL(locId string) string {
	return fmt.Sprintf("https://www.notams.faa.gov/dinsQueryWeb/queryRetrievalMapAction.do?reportType=Raw&retrieveLocId=%s&actionType=notamRetrievalbyICAOs", locId)
}

var myClient = &http.Client{Timeout: 10 * time.Second}

func GetJson(url string, target interface{}) error {
	r, err := myClient.Get(url)
	if err != nil {
		return err
	}
	defer r.Body.Close()

	return json.NewDecoder(r.Body).Decode(target)
}
