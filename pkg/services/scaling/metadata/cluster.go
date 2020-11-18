package metadata

import (
	"strings"

	"io/ioutil"
	"net/http"
)

func GetClusterInfo() (projectId, zone string, err error) {
	httpclient := &http.Client{}
	projectIdReq, err := http.NewRequest("GET", "http://metadata.google.internal/computeMetadata/v1/project/project-id", nil)
	projectIdReq.Header.Add("Metadata-Flavor", "Google")
	projectIdResp, err := httpclient.Do(projectIdReq)
	zoneReq, err := http.NewRequest("GET", "http://metadata.google.internal/computeMetadata/v1/instance/zone", nil)
	zoneReq.Header.Add("Metadata-Flavor", "Google")
	zoneResp, err := httpclient.Do(zoneReq)
	defer zoneResp.Body.Close()
	defer projectIdResp.Body.Close()
	projectIdBody, err := ioutil.ReadAll(projectIdResp.Body)
	projectId = string(projectIdBody)
	zoneBody, err := ioutil.ReadAll(zoneResp.Body)
	zoneSlice := strings.Split(string(zoneBody), "/")
	zone = zoneSlice[len(zoneSlice)-1]

	return projectId, zone, err
}
