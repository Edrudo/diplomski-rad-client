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
	"strings"
	"sync"

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
	case SendPhoto:
		c.sendPhoto(command.Args)
	default:
		utils.DefaultLogger.Fatalf(
			errors.New(
				"unknown command",
			), exitcodes.ExitUnkownCommand,
		)
	}
}

func (c *Client) sendPhoto(args []string) {
	imagePartSize := 1400
	if len(args) < 2 {
		utils.DefaultLogger.Fatalf(
			errors.New(
				"arguments needed for the SendPhoto command: "+
					"\t - url where image will be sent"+
					"\t - at least one path to image that needs to be sent",
			), exitcodes.ExitBadArguments,
		)
	}
	addr := args[0]
	imagePaths := args[1:]

	for _, imagePath := range imagePaths {
		image, err := os.ReadFile(imagePath)
		if err != nil {
			utils.DefaultLogger.Fatalf(err, exitcodes.ExitFailedReadingImage)
		}

		s := strings.Split(imagePath, ".")
		imageFormat := s[len(s)-1]

		imageParts := make([]ImagePart, 0)
		numImageParts := len(image) / imagePartSize
		if len(imageParts)%1450 > 0 {
			numImageParts++
		}

		c.HashGenerator.Write(image)
		calculatedHash := base64.URLEncoding.EncodeToString(c.HashGenerator.Sum(nil))

		var wg sync.WaitGroup
		wg.Add(numImageParts)
		for i := 0; i < numImageParts; i++ {
			/*if used for testing purposes
			i == numImageParts/2 {
				fmt.Println("Sleeping for 10 seconds")
				time.Sleep(10 * time.Second)
			}*/
			go func(partNumber int) {
				bdy, err := json.Marshal(
					ImagePart{
						ImageHash:  fmt.Sprintf("%v.%v", calculatedHash, imageFormat),
						PartNumber: partNumber + 1,
						TotalParts: numImageParts,
						PartData:   image[partNumber*imagePartSize : (partNumber+1)*imagePartSize],
					},
				)
				if err != nil {
					utils.DefaultLogger.Fatalf(err, exitcodes.ExitFailedProcessingImage)
				}

				for true {
					utils.DefaultLogger.Infof("GET %s", addr)
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
	}
}
