package main

import (
	"errors"
	"os"
	"strconv"

	"http3-client-poc/cmd/bootstrap"
	"http3-client-poc/cmd/exitcodes"
	"http3-client-poc/internal/application"
	"http3-client-poc/internal/utils"
)

func main() {
	// extracting image path from args
	args := os.Args
	if len(args) < 2 {
		utils.DefaultLogger.Fatalf(
			errors.New(
				"arguments needed for the program: "+
					"\t - command number"+
					"\t - arguments needed for the command",
			), exitcodes.ExitBadArguments,
		)
	}

	// initlize client
	client, roundTripper := bootstrap.NewClient()
	defer func() {
		// this defer is unreachable because os.Exit ignores all defers left
		err := roundTripper.Close()
		if err != nil {
			utils.DefaultLogger.Errorf("Error closing round tripper: %s", err)
		}
	}()

	commandNum, err := strconv.Atoi(args[1])
	if err != nil {
		utils.DefaultLogger.Fatalf(
			errors.New(
				"command number is not a number",
			), exitcodes.ExitUnkownCommand,
		)
	}

	client.ExecuteCommand(
		application.Command{
			Name: application.CommandName(commandNum),
			Args: args[2:],
		},
	)

	os.Exit(exitcodes.ExitSuccess)
}
