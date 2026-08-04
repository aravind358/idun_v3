package ontology

// TemporalType defines the kind of temporal reference extracted.
type TemporalType string

const (
	TempAbsoluteDate     TemporalType = "ABSOLUTE_DATE"
	TempRelativeDate     TemporalType = "RELATIVE_DATE"
	TempRelativeWeekday  TemporalType = "RELATIVE_WEEKDAY"
	TempClockTime        TemporalType = "CLOCK_TIME"
	TempRelativeDuration TemporalType = "RELATIVE_DURATION"
	TempTimeInterval     TemporalType = "TIME_INTERVAL"
	TempDaypart          TemporalType = "DAYPART"
	TempRecurrence       TemporalType = "RECURRENCE"
	TempUnknown          TemporalType = "UNKNOWN"
	
	// Legacy generic types (to be phased out or kept as fallbacks)
	TempAbsolute TemporalType = "ABSOLUTE"
	TempRelative TemporalType = "RELATIVE"
	TempDuration TemporalType = "DURATION"
)
