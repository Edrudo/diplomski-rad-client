package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"os"
	"sync"

	"http3-client-poc/cmd/bootstrap"
	"http3-client-poc/cmd/exitcodes"
	"http3-client-poc/internal/utils"
)

type ImagePart struct {
	ImageHash  string `json:"imageHash"`
	PartNumber int    `json:"partNumber"`
	TotalParts int    `json:"totalParts"`
	PartData   []byte `json:"partData"`
}

func main() {
	// extracting image path from args
	args := os.Args
	if len(args) < 3 {
		utils.DefaultLogger.Fatalf(
			errors.New(
				"arguments needed for the program: "+
					"\t - url where image will be sent"+
					"\t - at least one path to image that needs to be sent",
			), exitcodes.ExitBadArguments,
		)
	}
	addr := args[1]
	imagePaths := args[2:]
	imagePartSize := 1400

	// initlize client
	client, roundTripper := bootstrap.NewClient()
	defer func() {
		// this defer is unreachable because os.Exit ignores all defers left
		err := roundTripper.Close()
		if err != nil {
			utils.DefaultLogger.Errorf("Error closing round tripper: %s", err)
		}
	}()

	for _, imagePath := range imagePaths {
		image, err := os.ReadFile(imagePath)
		if err != nil {
			utils.DefaultLogger.Fatalf(err, exitcodes.ExitFailedReadingImage)
		}

		imageParts := make([]ImagePart, 0)
		numImageParts := len(image) / imagePartSize
		if len(imageParts)%1450 > 0 {
			numImageParts++
		}

		client.HashGenerator.Write(image)
		calculatedHash := base64.URLEncoding.EncodeToString(client.HashGenerator.Sum(nil))

		var wg sync.WaitGroup
		wg.Add(numImageParts)
		for i := 0; i < numImageParts; i++ {
			go func(partNumber int) {
				bdy, err := json.Marshal(
					ImagePart{
						ImageHash:  calculatedHash,
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
					rsp, err := client.HttpClient.Post(addr, "application/json", bytes.NewBuffer(bdy))
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

	os.Exit(exitcodes.ExitSuccess)

}
