# Native Devices Capability

## Purpose
The Native Devices Capability provides mechanical execution boundaries for physical hardware devices and sensors natively exposed by the operating system for IDUN V3. Following the Phase 3A Capability Philosophy, it handles mechanical sensor readings and hardware status discovery without applying any AI interpretation or spatial mapping reasoning.

## Scope Integration
**Important:** As defined in Phase 3B.6, `CategoryLocation` has been deprecated as a standalone capability and successfully merged into the `Native Devices Capability` under the `CategoryDevicesSensors` taxonomy. The Devices capability is now the exclusive mechanical interface for GPS and compass sensors.

## Architecture
- **Router**: Handlers split across `execute_usb.go`, `execute_bluetooth.go`, `execute_battery.go`, `execute_power.go`, `execute_location.go`, `execute_sensors.go`, `execute_hid.go`, and `execute_metadata.go`.
- **Provider Interface**: The `DevicesProvider` isolates hardware interaction.
- **Implementations**: `NativeProvider` executes hardware commands (stubbed for future binding). `MockProvider` simulates flawless, stateless device queries for continuous integration.

## Modules
- **USB**: Device listing and identification.
- **Bluetooth**: Pairing, unpairing, and nearby detection.
- **Power & Battery**: Charge levels, AC detection.
- **Location**: Raw mechanical GPS readouts.
- **Sensors**: Raw accelerometer, gyroscope, and compass data.
- **HID**: Enumerating connected input peripherals.
