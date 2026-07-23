# Native Media Capability

## Purpose
The Native Media Capability provides mechanical execution boundaries for operating system multimedia resources within IDUN V3. Following the Phase 3A Capability Philosophy, it handles mechanical playback, recording, capture, and device enumeration without applying any cognitive understanding, speech recognition, image processing, or AI analysis.

## Architecture
- **Router**: Handlers split across `execute_audio.go`, `execute_video.go`, `execute_image.go`, `execute_metadata.go`, and `execute_devices.go` based on the requested enum operation.
- **Provider Interface**: The `MediaProvider` isolates interaction with host hardware (cameras, microphones, speakers, and codecs).
- **Implementations**: `NativeProvider` executes the actual hardware commands. `MockProvider` enables completely isolated testing pipelines.

## Internal Modules
- **Audio**: Playback, recording, and session state control.
- **Video**: Playback, recording, and session state control.
- **Images**: Capturing from hardware, loading, and saving byte streams.
- **Metadata**: Reading media file durations, resolutions, and codecs.
- **Devices**: Enumerating available microphones, cameras, and display outputs.
