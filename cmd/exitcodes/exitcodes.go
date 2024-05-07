package exitcodes

const (
	ExitSuccess = iota
	ExitUnkownCommand
	ExitFailedInit
	ExitBadArguments
	ExitFailedReadingImage
	ExitFailedProcessingImage
)
