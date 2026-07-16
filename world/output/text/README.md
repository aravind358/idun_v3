# Text Output — Future Enhancement Reference

**Architecture Version:** `2.0.0-FROZEN`
**Package:** `idun/world/output/text`
**Status:** REFERENCE — Phase 1 TextOutputAdapter is implemented at `idun/world/adapters/text`

The Phase 1 concrete text output implementation is co-located with the input adapter
at `idun/world/adapters/text/adapter.go` as `TextOutputAdapter`.

This directory is reserved for future output-specific overrides, formatters,
or multi-channel routing implementations that may warrant separation from the
bidirectional adapter package.
