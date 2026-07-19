# Presentation Subsystem: Language Realization Layer (`presentation/realization`)

**Architecture Version:** `2.0.0-FROZEN`  
**Layer Classification:** Presentation Subsystem ("IDUN's Mouth")

## Purpose
The Language Realization Layer converts an already-approved internal `ExecutionResponse` into natural human language (`RealizedOutput`). It is strictly a presentation layer and never participates in cognition.

## Constitutional Rules
1. **Statelessness**: The service never retains conversation history, session memory, prompts, or generated outputs across cycles. Every invocation follows `ExecutionResponse -> BuildPrompt -> Inference -> RealizedOutput -> Forget`.
2. **Model-Agnostic**: Communicates exclusively via `InferenceService` (`idun/intelligence/infrastructure/inference`) using a configurable `ModelID`. It possesses zero knowledge of physical models (`Ollama`, `Llama`, `Qwen`, `Gemma`, `GPT`, etc.).
3. **Absolute Meaning & Fact Preservation**: Never invents facts, removes information, or alters approved decisions.
4. **Zero Cognition**: Does not perform NLU (`Understanding`), inference (`Reasoning`), planning (`Planning`), or decision comparison (`Decision`).
