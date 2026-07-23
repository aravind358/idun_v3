# Native Capability Template

This directory provides the standard engineering scaffold for implementing new native capabilities in IDUN V3.

## Purpose

To ensure architectural consistency across the entire Capability Framework, this template enforces the required boundary protections, interface implementations, metrics tracking, and platform-abstraction patterns derived from the certified `Native System Capability`.

## How to Create a New Capability

1. **Copy the Template**: Duplicate this `template` directory and rename it to your target capability name (e.g. `files`, `network`).
2. **Rename Package**: Update the `package template` declaration across all files to match your new capability name.
3. **Define Operations**: Update `operations.go` with strongly-typed enums for your specific operations (e.g. `OperationReadFile`).
4. **Update Metadata**: Modify `metadata.go` to declare the correct `Name`, `Category`, `Description`, `Permissions`, and `Tags`.
5. **Implement Provider**: Update `provider.go` with specific methods, and implement them in `provider_native.go` (and `provider_mock.go`). Use build tags (`//go:build windows`) if you require OS-specific provider files.
6. **Update Router**: Modify `execute.go` and add specific `execute_*.go` files to route your operations to your provider methods safely.
7. **Update Validation & Permissions**: Map your operations to the correct checks in `validation.go` and `permissions.go`.
8. **Write Tests**: Expand `template_test.go` (renaming it) to cover your new operations using the `MockProvider`.

## Engineering Conventions

- **NO Direct Native Calls**: Do not import "os" or "os/exec" directly into the capability core. All native execution must be deferred to the `Provider`.
- **NO String Types for Operations**: Use `TemplateOperation` (rename to your domain) constants.
- **NEVER Bypass Permissions**: `checkPermission` must always be evaluated prior to provider dispatch.
- **STRICT Error Normalization**: Return only `capabilities.CapabilityResult` or `capabilities.CapabilityError`. No raw Go errors should reach the Intelligence layer.
