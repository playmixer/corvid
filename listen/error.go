package listen

type ErrorString string

func (e ErrorString) Error() string {
	return string(e)
}

const (
	//device
	ErrNotFoundDevice ErrorString = "Error: Not found device"

	//wav
	ErrNotEnoughDataToParseWav ErrorString = "Error: Not enough data to parse the WAV header"
	ErrInvalidWav              ErrorString = "Error: Invalid WAV signature"
	ErrFileIsNotPCM            ErrorString = "Error: Only PCM files are supported"
)
