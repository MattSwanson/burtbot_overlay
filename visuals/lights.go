package visuals

import (
	"crypto/tls"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strings"
)

var HUE_APP_KEY = os.Getenv("HUE_USER_ID")

func SetLightsColor(color int) {
	endPoint := "https://192.168.0.5/clip/v2/resource/light/7f7db8cf-5a99-46bd-958c-671e0c975cba"
	colorX, colorY := rand.Float32(), rand.Float32()
	reqBody := fmt.Sprintf(`{"on":{"on":true}, "dimming":{"brightness":50.0},"color":{"xy":{"x":%.2f,"y":%.2f}}}`, colorX, colorY)
	br := strings.NewReader(reqBody)
	req, err := http.NewRequest("PUT", endPoint, br)
	if err != nil {
		log.Println(err.Error())
		return
	}
	req.Header.Set("hue-application-key", HUE_APP_KEY)
	ct := http.DefaultTransport.(*http.Transport).Clone()
	ct.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	client := &http.Client{Transport: ct}
	_, err = client.Do(req)
	if err != nil {
		log.Println(err.Error())
		return
	}
}
