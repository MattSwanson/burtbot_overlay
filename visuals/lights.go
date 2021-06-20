package visuals

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
)

var bridgeID string = os.Getenv("HUE_BRIDGE_ID")

func SetLightsColor(color int) {
	endPoint := fmt.Sprintf("http://10.0.0.2/api/%s/groups/1/action", bridgeID)
	reqBody := fmt.Sprintf(`{"on":true, "hue":%d}`, color)
	br := strings.NewReader(reqBody)
	req, err := http.NewRequest("PUT", endPoint, br)
	if err != nil {
		log.Println(err.Error())
		return
	}
	_, err = http.DefaultClient.Do(req)
	if err != nil {
		log.Println(err.Error())
		return
	}
}
