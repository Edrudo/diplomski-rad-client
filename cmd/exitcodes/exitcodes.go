package exitcodes

const (
	ExitSuccess = iota
	ExitFailedInit
	ExitBadArguments
	ExitFailedReadingImage
	ExitFailedProcessingImage
)
