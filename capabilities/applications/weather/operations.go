package weather

// WeatherOperation defines strongly-typed constants for permitted operations.
type WeatherOperation string

const (
	OperationCurrent  WeatherOperation = "current"
	OperationForecast WeatherOperation = "forecast"
)

// IsValid validates if a string matches a known WeatherOperation.
func (o WeatherOperation) IsValid() bool {
	switch o {
	case OperationCurrent, OperationForecast:
		return true
	}
	return false
}
