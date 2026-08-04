# TODO — Future Architecture Enhancement
## Phase X — Unified Policy Engine

**Status:** Deferred (Post-Restoration)

### Objective

Introduce a centralized `PolicyEngine` to unify authorization and policy decisions across all application capabilities.

### Current State

Sprint 5 introduced `PermissionPolicy` within `app-files-1` to enforce filesystem authorization. For the current restoration phase, this is the correct level of abstraction because only filesystem authorization is required.

### Future Motivation

As IDUN grows, multiple domains will require policy decisions, including:
- Trusted directories
- User roles and permissions
- Confirmation requirements for sensitive actions
- Plugin permissions and sandboxing
- Network and internet access policies
- Destructive operation approvals
- Privacy and data access policies
- Future cloud or external service permissions

Rather than implementing separate policy systems for each domain, these should be unified under a single `PolicyEngine`.

### Proposed Evolution

```
Understanding
        ↓
Planning
        ↓
Executive
        ↓
Application Capability
        ↓
PolicyEngine
        ├── FilesystemPolicy
        ├── NetworkPolicy
        ├── PluginPolicy
        ├── ConfirmationPolicy
        ├── UserRolePolicy
        ├── PrivacyPolicy
        └── Future Policies
        ↓
Native Capability
```

### Architectural Principles
- The `PolicyEngine` owns policy evaluation only.
- Individual policy modules remain independent and composable.
- Application capabilities request policy decisions but do not implement complex authorization logic themselves.
- Native capabilities remain purely mechanical and never enforce application-level policies.

### Migration Strategy
- **Do not replace `PermissionPolicy` during the restoration phase.**
- Keep the existing `PermissionPolicy` implementation stable.
- When the `PolicyEngine` is introduced, migrate `PermissionPolicy` into `FilesystemPolicy` without changing external behavior.
- This migration should preserve all existing runtime semantics and require no changes to Understanding, Planning, Executive, or Native capabilities.
