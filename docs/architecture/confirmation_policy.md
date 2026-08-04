# Future Architecture Enhancement — Confirmation Policy

**Status:** Deferred (Post-Restoration)

### Objective

Introduce a dedicated `ConfirmationPolicy` to explicitly handle user confirmations for destructive or significant system state changes.

### Future Motivation

Operations such as:
- `shutdown`
- `restart`
- `logoff`
- `sign out`
- `delete user data`

should pass through a dedicated Confirmation Policy before execution. Currently, `PermissionPolicy` handles the strict binary authorization (Allowed/Blocked), but confirmation flows require stateful or multi-turn interaction.

### Proposed Evolution

The future architecture would insert the Confirmation Policy explicitly before the Native Execution:

```
Understanding
        ↓
Planning
        ↓
Executive
        ↓
app-system-1
        ↓
PermissionPolicy
        ↓
ConfirmationPolicy  <-- New dedicated layer for confirmations
        ↓
NativeSystemCapability
```

### Constraints During Restoration

For the current restoration phase, **PermissionPolicy alone is sufficient**. The `ConfirmationPolicy` should remain a documented future enhancement rather than being built into Sprint 6.
