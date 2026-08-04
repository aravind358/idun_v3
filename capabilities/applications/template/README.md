# Application Capability Template

This directory provides the standard engineering scaffold for implementing new Application Capabilities in IDUN V3.

## Architectural Rules

Application Capabilities deliver user-facing functionality by orchestrating Native Capabilities or executing pure deterministic logic.

**CRITICAL RULES:**
1. ❌ MUST NOT import `os` or `os/exec`.
2. ❌ MUST NOT import `net/http`.
3. ❌ MUST NOT access hardware directly.
4. ❌ MUST NOT bypass Native Providers.

**Two Valid Execution Models:**
- **Model 1 (Orchestration)**: Composing multiple Native Capabilities. Use `c.Resolver.Resolve()` to get Native Capabilities.
- **Model 2 (Deterministic)**: Pure computation (e.g. math/logic) that doesn't need to interact with the external world.

## How to Create a New Application Capability

1. **Copy the Template**: Duplicate this `template` directory and rename it to your target capability name (e.g. `currency`).
2. **Rename Package**: Update the `package template` declaration across all files to match your new capability name.
3. **Define Operations**: Update `operations.go` with strongly-typed enums for your specific operations (e.g. `OperationConvert`).
4. **Update Metadata**: Modify `metadata.go` to declare the correct `Name`, `Category`, `Description`, `Permissions`, and `Tags`.
   - **Note**: Do not invent new categories. Assign the most appropriate existing `CapabilityCategory` from `types.go`.
5. **Implement Capability**: Update `capability.go` with specific execution logic. Use `c.Resolver` to invoke Native capabilities if necessary.
6. **Write Tests**: Expand tests to cover your logic.
7. **Register**: Add your capability to `LoadApplicationCapabilities()` in `loader.go`.
