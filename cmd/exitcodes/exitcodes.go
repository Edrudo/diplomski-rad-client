package exitcodes

const (
	ExitSuccess = iota
	ExitUnkownCommand
	ExitFailedInit
	ExitBadArguments
	ExitFailedReadingFile
	ExitFailedProcessingData
)
