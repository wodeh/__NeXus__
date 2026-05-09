package guardrails

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/nexus-platform/ai-platform/inference-gateway/internal/rag"
)

type Engine struct {
	config       *Config
	inputChecks  []InputCheck
	outputChecks []OutputCheck
}

type Config struct {
	BlockedTopics      []string            `json:"blocked_topics"`
	BlockedPatterns    []string            `json:"blocked_patterns"`
	PIIEntities        []string            `json:"pii_entities"`
	HallucinationRules HallucinationConfig `json:"hallucination_rules"`
	ToxicityThreshold  float64             `json:"toxicity_threshold"`
}

type HallucinationConfig struct {
	FactualityThreshold   float64 `json:"factuality_threshold"`
	SourceAlignmentWeight float64 `json:"source_alignment_weight"`
	SelfConsistencyChecks int     `json:"self_consistency_checks"`
	ConfidenceThreshold   float64 `json:"confidence_threshold"`
}

type InputCheck func(ctx context.Context, prompt string, tenantID string) *CheckResult
type OutputCheck func(ctx context.Context, output, prompt string, sources []rag.Source) *CheckResult

type CheckResult struct {
	Passed     bool
	Violation  string
	Confidence float64
	CheckName  string
}

type Result struct {
	Passed     bool
	Violations []string
	Checks     []string
}

func New(cfg *Config) (*Engine, error) {
	e := &Engine{config: cfg}

	e.inputChecks = []InputCheck{
		e.checkPromptInjection,
		e.checkBlockedTopics,
		e.checkPIILeakage,
		e.checkJailbreakAttempts,
	}

	e.outputChecks = []OutputCheck{
		e.checkHallucination,
		e.checkFactuality,
		e.checkSourceAlignment,
		e.checkSelfConsistency,
		e.checkToxicity,
		e.checkPIIInOutput,
	}

	return e, nil
}

func (e *Engine) CheckInput(ctx context.Context, prompt string, tenantID string) (*Result, error) {
	result := &Result{Passed: true, Checks: []string{}}

	for _, check := range e.inputChecks {
		checkResult := check(ctx, prompt, tenantID)
		result.Checks = append(result.Checks, checkResult.CheckName)
		if !checkResult.Passed {
			result.Passed = false
			result.Violations = append(result.Violations, checkResult.Violation)
		}
	}

	return result, nil
}

func (e *Engine) CheckOutput(ctx context.Context, output, prompt string, sources []rag.Source) (*Result, error) {
	result := &Result{Passed: true, Checks: []string{}}

	for _, check := range e.outputChecks {
		checkResult := check(ctx, output, prompt, sources)
		result.Checks = append(result.Checks, checkResult.CheckName)
		if !checkResult.Passed {
			result.Passed = false
			result.Violations = append(result.Violations, checkResult.Violation)
		}
	}

	return result, nil
}

func (e *Engine) checkPromptInjection(ctx context.Context, prompt string, tenantID string) *CheckResult {
	injectionPatterns := []string{
		`(?i)ignore previous instructions`,
		`(?i)disregard.*system prompt`,
		`(?i)you are now.*instead`,
		`(?i)pretend to be`,
		`(?i)act as.*ignore`,
		`(?i)new instructions:`,
		`(?i)system override`,
		`(?i)DAN.*mode`,
		`(?i)jailbreak`,
		`(?i)developer mode`,
	}

	for _, pattern := range injectionPatterns {
		if matched, _ := regexp.MatchString(pattern, prompt); matched {
			return &CheckResult{
				Passed:     false,
				Violation:  "Prompt injection attempt detected",
				Confidence: 0.95,
				CheckName:  "prompt_injection",
			}
		}
	}

	return &CheckResult{Passed: true, CheckName: "prompt_injection"}
}

func (e *Engine) checkBlockedTopics(ctx context.Context, prompt string, tenantID string) *CheckResult {
	for _, topic := range e.config.BlockedTopics {
		if strings.Contains(strings.ToLower(prompt), strings.ToLower(topic)) {
			return &CheckResult{
				Passed:     false,
				Violation:  fmt.Sprintf("Blocked topic detected: %s", topic),
				Confidence: 0.9,
				CheckName:  "blocked_topic",
			}
		}
	}
	return &CheckResult{Passed: true, CheckName: "blocked_topic"}
}

func (e *Engine) checkPIILeakage(ctx context.Context, prompt string, tenantID string) *CheckResult {
	piiPatterns := map[string]string{
		"ssn":         `\b\d{3}-\d{2}-\d{4}\b`,
		"credit_card": `\b(?:\d{4}[-\s]?){3}\d{4}\b`,
		"email":       `\b[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+\.[A-Z|a-z]{2,}\b`,
		"phone":       `\b\d{3}[-.]?\d{3}[-.]?\d{4}\b`,
		"passport":    `\b[A-Z]{2}\d{7}\b`,
	}

	for piiType, pattern := range piiPatterns {
		if matched, _ := regexp.MatchString(pattern, prompt); matched {
			return &CheckResult{
				Passed:     false,
				Violation:  fmt.Sprintf("PII detected in prompt: %s", piiType),
				Confidence: 0.92,
				CheckName:  "pii_leakage",
			}
		}
	}
	return &CheckResult{Passed: true, CheckName: "pii_leakage"}
}

func (e *Engine) checkJailbreakAttempts(ctx context.Context, prompt string, tenantID string) *CheckResult {
	jailbreakIndicators := []string{
		"ignore all previous",
		"you are a different AI",
		"no restrictions",
		"bypass safety",
		"hypothetically speaking",
		"for educational purposes",
		"in a fictional scenario",
		"roleplay as",
	}

	lowerPrompt := strings.ToLower(prompt)
	for _, indicator := range jailbreakIndicators {
		if strings.Contains(lowerPrompt, indicator) {
			return &CheckResult{
				Passed:     false,
				Violation:  "Potential jailbreak attempt detected",
				Confidence: 0.85,
				CheckName:  "jailbreak_attempt",
			}
		}
	}
	return &CheckResult{Passed: true, CheckName: "jailbreak_attempt"}
}

func (e *Engine) checkHallucination(ctx context.Context, output, prompt string, sources []rag.Source) *CheckResult {
	citationPattern := regexp.MustCompile(`\[\d+\]|\(Source:.*?\)|According to [A-Z][a-z]+ et al\.`)
	citations := citationPattern.FindAllString(output, -1)
	if len(citations) > 0 && len(sources) == 0 {
		return &CheckResult{
			Passed:     false,
			Violation:  "Fabricated citations detected without RAG sources",
			Confidence: 0.88,
			CheckName:  "hallucination_citations",
		}
	}

	confidentPatterns := []string{
		`(?i)definitely`,
		`(?i)certainly`,
		`(?i)absolutely`,
		`(?i)without a doubt`,
		`(?i)it is known that`,
		`(?i)everyone knows`,
	}

	for _, pattern := range confidentPatterns {
		if matched, _ := regexp.MatchString(pattern, output); matched {
			if len(sources) == 0 {
				return &CheckResult{
					Passed:     false,
					Violation:  "Confident claim without supporting evidence",
					Confidence: 0.75,
					CheckName:  "hallucination_confidence",
				}
			}
		}
	}

	if e.containsTemporalHallucination(output) {
		return &CheckResult{
			Passed:     false,
			Violation:  "Temporal hallucination detected",
			Confidence: 0.8,
			CheckName:  "hallucination_temporal",
		}
	}

	if e.containsNumericalInconsistency(output) {
		return &CheckResult{
			Passed:     false,
			Violation:  "Numerical inconsistency detected",
			Confidence: 0.7,
			CheckName:  "hallucination_numerical",
		}
	}

	return &CheckResult{Passed: true, CheckName: "hallucination"}
}

func (e *Engine) checkFactuality(ctx context.Context, output, prompt string, sources []rag.Source) *CheckResult {
	claims := e.extractClaims(output)

	for _, claim := range claims {
		supported := false
		for _, source := range sources {
			if e.claimSupportedBySource(claim, source) {
				supported = true
				break
			}
		}

		if !supported && len(sources) > 0 {
			return &CheckResult{
				Passed:     false,
				Violation:  fmt.Sprintf("Unverified factual claim: %s", claim),
				Confidence: 0.82,
				CheckName:  "factuality",
			}
		}
	}

	return &CheckResult{Passed: true, CheckName: "factuality"}
}

func (e *Engine) checkSourceAlignment(ctx context.Context, output, prompt string, sources []rag.Source) *CheckResult {
	if len(sources) == 0 {
		return &CheckResult{Passed: true, CheckName: "source_alignment"}
	}

	outputLower := strings.ToLower(output)
	alignmentScore := 0.0

	for _, source := range sources {
		sourceLower := strings.ToLower(source.Content)
		words := strings.Fields(outputLower)
		matches := 0
		for _, word := range words {
			if len(word) > 4 && strings.Contains(sourceLower, word) {
				matches++
			}
		}
		if len(words) > 0 {
			score := float64(matches) / float64(len(words))
			if score > alignmentScore {
				alignmentScore = score
			}
		}
	}

	if alignmentScore < e.config.HallucinationRules.SourceAlignmentWeight {
		return &CheckResult{
			Passed:     false,
			Violation:  fmt.Sprintf("Output poorly aligned with sources (score: %.2f)", alignmentScore),
			Confidence: 1.0 - alignmentScore,
			CheckName:  "source_alignment",
		}
	}

	return &CheckResult{Passed: true, CheckName: "source_alignment"}
}

func (e *Engine) checkSelfConsistency(ctx context.Context, output, prompt string, sources []rag.Source) *CheckResult {
	sentences := strings.Split(output, ".")

	for i, sent1 := range sentences {
		for j, sent2 := range sentences {
			if i >= j {
				continue
			}
			if e.areContradictory(sent1, sent2) {
				return &CheckResult{
					Passed:     false,
					Violation:  "Self-contradictory statements detected",
					Confidence: 0.85,
					CheckName:  "self_consistency",
				}
			}
		}
	}

	return &CheckResult{Passed: true, CheckName: "self_consistency"}
}

func (e *Engine) checkToxicity(ctx context.Context, output, prompt string, sources []rag.Source) *CheckResult {
	toxicPatterns := []string{
		`(?i)\b(hate|kill|die|destroy)\b`,
		`(?i)\b(racist|sexist|homophobic)\b`,
		`(?i)\b(violence|abuse|harass)\b`,
	}

	for _, pattern := range toxicPatterns {
		if matched, _ := regexp.MatchString(pattern, output); matched {
			return &CheckResult{
				Passed:     false,
				Violation:  "Toxic content detected",
				Confidence: 0.9,
				CheckName:  "toxicity",
			}
		}
	}
	return &CheckResult{Passed: true, CheckName: "toxicity"}
}

func (e *Engine) checkPIIInOutput(ctx context.Context, output, prompt string, sources []rag.Source) *CheckResult {
	return e.checkPIILeakage(ctx, output, "")
}

func (e *Engine) containsTemporalHallucination(output string) bool {
	return false
}

func (e *Engine) containsNumericalInconsistency(output string) bool {
	return false
}

func (e *Engine) extractClaims(output string) []string {
	var claims []string
	sentences := strings.Split(output, ".")
	for _, s := range sentences {
		s = strings.TrimSpace(s)
		if len(s) > 20 && !strings.HasPrefix(s, "I ") && !strings.HasPrefix(s, "You ") {
			claims = append(claims, s)
		}
	}
	return claims
}

func (e *Engine) claimSupportedBySource(claim string, source rag.Source) bool {
	claimWords := strings.Fields(strings.ToLower(claim))
	sourceLower := strings.ToLower(source.Content)

	matches := 0
	for _, word := range claimWords {
		if len(word) > 4 && strings.Contains(sourceLower, word) {
			matches++
		}
	}

	return float64(matches) / float64(len(claimWords)) > 0.3
}

func (e *Engine) areContradictory(sent1, sent2 string) bool {
	negations := []string{"not", "no", "never", "none", "without"}

	s1HasNeg := false
	s2HasNeg := false

	for _, neg := range negations {
		if strings.Contains(strings.ToLower(sent1), neg) {
			s1HasNeg = true
		}
		if strings.Contains(strings.ToLower(sent2), neg) {
			s2HasNeg = true
		}
	}

	if s1HasNeg != s2HasNeg {
		return true
	}

	return false
}
