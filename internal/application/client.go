package application

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"hash"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/quic-go/quic-go/http3"

	"http3-client-poc/cmd/exitcodes"
	"http3-client-poc/internal/utils"
)

type Client struct {
	HashGenerator hash.Hash
	HttpClient    *http.Client
	roundTriper   *http3.RoundTripper
}

func NewClient(
	hashGenerator hash.Hash,
	httpClient *http.Client,
	roundTriper *http3.RoundTripper,
) *Client {
	return &Client{
		HashGenerator: hashGenerator,
		HttpClient:    httpClient,
		roundTriper:   roundTriper,
	}
}

func (c *Client) ExecuteCommand(command Command) {
	switch command.Name {
	case SendGeoshot:
		c.sendGeoshot(command.Args)
	default:
		utils.DefaultLogger.Fatalf(
			errors.New(
				"unknown command",
			), exitcodes.ExitUnkownCommand,
		)
	}
}

func (c *Client) sendGeoshot(args []string) {

	if len(args) < 2 {
		utils.DefaultLogger.Fatalf(
			errors.New(
				"arguments needed for the SendGeoshot command: "+
					"\t - url where data will be sent"+
					"\t - at least one path to json that needs to be sent",
			), exitcodes.ExitBadArguments,
		)
	}
	addr := args[0]
	geoshotPaths := args[1:]

	for _, geoshotPath := range geoshotPaths {
		readFile, err := os.ReadFile(geoshotPath)
		if err != nil {
			utils.DefaultLogger.Fatalf(err, exitcodes.ExitFailedReadingFile)
		}

		geoshot := &Geoshot{}
		err = json.Unmarshal(readFile, geoshot)
		if err != nil {
			utils.DefaultLogger.Fatalf(err, exitcodes.ExitFailedUnmarshalling)
		}

		wg := &sync.WaitGroup{}
		for i := 0; i < 21; i++ {
			wg.Add(1)
			tmpGeoShot := *geoshot
			tmpGeoShot.Timestamp = fmt.Sprintf("%v", time.Now())
			go c.send(wg, addr, tmpGeoShot)
			time.Sleep(2 * time.Second)
		}

		wg.Wait()
	}
}

func (c *Client) send(wgOut *sync.WaitGroup, addr string, geoshot Geoshot) {
	dataPartSize := 1400

	readFile, err := json.Marshal(geoshot)
	if err != nil {
		utils.DefaultLogger.Fatalf(err, exitcodes.ExitFailedUnmarshalling)
	}

	// imageParts := make([]DataPart, 0)
	numDataParts := len(readFile) / dataPartSize
	if len(readFile)%1450 > 0 {
		numDataParts++
	}

	c.HashGenerator.Write(readFile)
	calculatedHash := base64.URLEncoding.EncodeToString(c.HashGenerator.Sum(nil))

	var wg sync.WaitGroup
	wg.Add(numDataParts)
	for i := 0; i < numDataParts; i++ {
		go func(partNumber int) {
			var bdy []byte
			if partNumber == numDataParts-1 {
				bdy, err = json.Marshal(
					DataPart{
						DataHash:   fmt.Sprintf("%v", calculatedHash),
						PartNumber: partNumber + 1,
						TotalParts: numDataParts,
						PartData:   readFile[partNumber*dataPartSize:],
					},
				)
				if err != nil {
					utils.DefaultLogger.Fatalf(err, exitcodes.ExitFailedProcessingData)
				}
			} else {
				bdy, err = json.Marshal(
					DataPart{
						DataHash:   fmt.Sprintf("%v", calculatedHash),
						PartNumber: partNumber + 1,
						TotalParts: numDataParts,
						PartData:   readFile[partNumber*dataPartSize : (partNumber+1)*dataPartSize],
					},
				)
				if err != nil {
					utils.DefaultLogger.Fatalf(err, exitcodes.ExitFailedProcessingData)
				}
			}

			for true {
				utils.DefaultLogger.Infof("POST %s", addr)
				rsp, err := c.HttpClient.Post(addr, "application/json", bytes.NewBuffer(bdy))
				if err == nil {
					utils.DefaultLogger.Infof("Got response for %s: %#v", addr, rsp)
					wg.Done()
					break
				}
				utils.DefaultLogger.Errorf(err.Error())
			}
		}(i)
	}
	wg.Wait()
	wgOut.Done()
}

type Geoshot struct {
	Image           []byte      `json:"jpg"`
	Timestamp       string      `json:"ts"`
	Age             int         `json:"age"`
	DeviceId        int         `json:"id"`
	Datetime        string      `json:"dt"`
	Coordinates     [][]float64 `json:"geo"`
	VehicleStanding int         `json:"vel"`
}
