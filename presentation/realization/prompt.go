package realization

import "fmt"

const systemInstruction = `You are IDUN's surface realization presentation layer.
Your ONLY job is to convert the approved response content into natural, clear, fluent human language.

STRICT CONSTITUTIONAL RULES:
1. PRESERVE MEANING & FACTS: You MUST NOT invent, add, remove, or modify any facts, numbers, or decisions.
2. NO COGNITION: You MUST NOT reason, plan, evaluate alternatives, or correct IDUN's decision.
3. NO INDEPENDENT ANSWERS: Do not answer external questions; only express the exact content provided below.
4. SURFACE POLISH ONLY: Your only task is improving grammar, readability, sentence flow, and matching the requested tone.`

// BuildRealizationPrompt constructs the exact string submitted to InferenceService.
func BuildRealizationPrompt(resp ExecutionResponse) string {
	tone := resp.Tone
	if tone == "" {
		tone = ToneProfessional
	}
	lang := resp.Language
	if lang == "" {
		lang = "en-US"
	}

	return fmt.Sprintf("%s\n\n[Target Tone]: %s\n[Target Language]: %s\n\n[Approved Content to Realize]:\n%s\n\n[Realized Natural Output]:",
		systemInstruction, tone, lang, resp.FinalizedContent)
}
