package media

// MediaOperation defines strongly-typed constants for permitted operations.
type MediaOperation string

const (
	// Audio
	OperationPlayAudio   MediaOperation = "play_audio"
	OperationStopAudio   MediaOperation = "stop_audio"
	OperationPauseAudio  MediaOperation = "pause_audio"
	OperationResumeAudio MediaOperation = "resume_audio"
	OperationRecordAudio MediaOperation = "record_audio"

	// Video
	OperationPlayVideo   MediaOperation = "play_video"
	OperationPauseVideo  MediaOperation = "pause_video"
	OperationResumeVideo MediaOperation = "resume_video"
	OperationStopVideo   MediaOperation = "stop_video"
	OperationRecordVideo MediaOperation = "record_video"

	// Image
	OperationCaptureImage MediaOperation = "capture_image"
	OperationLoadImage    MediaOperation = "load_image"
	OperationSaveImage    MediaOperation = "save_image"

	// Metadata
	OperationGetMetadata MediaOperation = "get_metadata"

	// Devices
	OperationListMediaDevices MediaOperation = "list_media_devices"
	OperationGetDevice        MediaOperation = "get_device"
	OperationListCodecs       MediaOperation = "list_codecs"
)

// IsValid validates if a string matches a known MediaOperation.
func (o MediaOperation) IsValid() bool {
	switch o {
	case OperationPlayAudio, OperationStopAudio, OperationPauseAudio,
		OperationResumeAudio, OperationRecordAudio, OperationPlayVideo,
		OperationPauseVideo, OperationResumeVideo, OperationStopVideo,
		OperationRecordVideo, OperationCaptureImage, OperationLoadImage,
		OperationSaveImage, OperationGetMetadata, OperationListMediaDevices,
		OperationGetDevice, OperationListCodecs:
		return true
	}
	return false
}
