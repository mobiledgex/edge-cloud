package deploygen

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/mobiledgex/edge-cloud/edgeproto"
)

func SendReq(uri string, app *edgeproto.App) (string, error) {
	spec, err := NewAppFromApp(app)
	if err != nil {
		return "", err
	}
	out, err := json.Marshal(spec)
	if err != nil {
		return "", err
	}
	req, err := http.NewRequest("POST", uri, bytes.NewBuffer(out))
	if err != nil {
		return "", fmt.Errorf("post %s request failed, %s", uri, err.Error())
	}
	req.Header.Set("Context-Type", "application/json")
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("post %s failed, %s", uri, err.Error())
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("expected status ok (200) but got %s", resp.Status)
	}
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("error reading reply body, %s", err.Error())
	}
	return string(data), nil
}

func RunGen(gen string, app *edgeproto.App) (string, error) {
	fx, found := Generators[gen]
	if !found {
		return "", fmt.Errorf("generator %s not found", gen)
	}
	spec, err := NewAppFromApp(app)
	if err != nil {
		return "", err
	}
	return fx(spec)
}
