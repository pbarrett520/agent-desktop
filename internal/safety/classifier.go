// Package safety classifies az CLI commands into risk tiers so the agent
// knows which ones can execute freely, which need human approval, and which
// need extra-loud approval because they are hard to undo.
package safety

import (
	"regexp"
	"strings"
)

// Tier is the risk classification of an az command.
type Tier int

const (
	// Read is allowlisted read-only verbs. Executes freely.
	Read Tier = iota
	// Mutate changes state but is recoverable. Requires user approval.
	Mutate
	// Destructive is hard or impossible to reverse. Requires approval AND
	// must be rendered in the danger color with the resource name echoed back.
	Destructive
)

func (t Tier) String() string {
	switch t {
	case Read:
		return "READ"
	case Mutate:
		return "MUTATE"
	case Destructive:
		return "DESTRUCTIVE"
	default:
		return "UNKNOWN"
	}
}

// Classification is the result of classifying a command.
type Classification struct {
	Tier Tier
	// Reason is the name of the rule that matched, for logging/audit.
	Reason string
	// Sensitive marks commands that are READ-tier but still worth logging
	// prominently, e.g. reading a keyvault secret.
	Sensitive bool
}

// rule is one entry in the classification table. Rules are evaluated in
// order; the first match for a given command segment wins.
type rule struct {
	name      string
	pattern   *regexp.Regexp
	tier      Tier
	sensitive bool
}

// rules is evaluated top to bottom. Specific overrides for named resource
// types come first so they win over the generic verb tables below them.
var rules = []rule{
	{
		name:    "help flag",
		pattern: regexp.MustCompile(`(^|\s)--help\b`),
		tier:    Read,
	},
	{
		name:    "az ad (identity/directory changes)",
		pattern: regexp.MustCompile(`^\s*az\s+ad\b`),
		tier:    Destructive,
	},
	{
		name:    "role assignment mutation",
		pattern: regexp.MustCompile(`^\s*az\s+role\s+assignment\s+(create|delete)\b`),
		tier:    Destructive,
	},
	{
		name:    "role definition mutation",
		pattern: regexp.MustCompile(`^\s*az\s+role\s+definition\s+(create|update|delete)\b`),
		tier:    Destructive,
	},
	{
		name:    "keyvault permission change",
		pattern: regexp.MustCompile(`^\s*az\s+keyvault\s+(set-policy|delete-policy)\b`),
		tier:    Destructive,
	},
	{
		name:      "keyvault secret/key/certificate read",
		pattern:   regexp.MustCompile(`^\s*az\s+keyvault\s+(secret|key|certificate)\s+(show|list|download|backup)\b`),
		tier:      Read,
		sensitive: true,
	},
	{
		name:    "az account show/list",
		pattern: regexp.MustCompile(`^\s*az\s+account\s+(show|list)\b`),
		tier:    Read,
	},
	{
		name:    "az account set",
		pattern: regexp.MustCompile(`^\s*az\s+account\s+set\b`),
		tier:    Mutate,
	},
	{
		name:    "az graph query",
		pattern: regexp.MustCompile(`^\s*az\s+graph\s+query\b`),
		tier:    Read,
	},
	{
		name:    "browse (no-op)",
		pattern: regexp.MustCompile(`^\s*az\s+.*\bbrowse\b`),
		tier:    Read,
	},
	{
		name:    "destructive verbs",
		pattern: regexp.MustCompile(`\b(delete|purge|remove|cancel|revoke|reset)\b`),
		tier:    Destructive,
	},
	{
		name:    "read verbs",
		pattern: regexp.MustCompile(`\b(show|list|get|describe|top)\b|\bget-[a-z-]+\b`),
		tier:    Read,
	},
	{
		name:    "mutate verbs",
		pattern: regexp.MustCompile(`\b(create|update|set|add|start|stop|restart|deploy|scale|tag)\b`),
		tier:    Mutate,
	},
}

// splitPattern breaks a command string into pipeline/chain segments so each
// segment can be classified independently. Handles &&, ||, ;, and |.
var splitPattern = regexp.MustCompile(`&&|\|\||[;|]`)

// azSegmentPattern identifies a pipeline/chain segment that is itself an az
// invocation, as opposed to a downstream filter like `grep` or `jq` that
// merely processes az's output and carries no risk of its own.
var azSegmentPattern = regexp.MustCompile(`(?i)^\s*az\b`)

// ClassifyAzCommand returns the risk tier for an az CLI command. For chained
// or piped commands, it classifies every az segment and returns the most
// dangerous tier found, fail-closed to Mutate for unrecognized verbs.
// Non-az segments (e.g. `| grep foo`, `| jq .`) are ignored since they only
// filter az's output and carry no risk of their own.
func ClassifyAzCommand(cmd string) Classification {
	segments := splitPattern.Split(cmd, -1)

	var best Classification
	initialized := false

	for _, seg := range segments {
		seg = strings.TrimSpace(seg)
		if seg == "" || !azSegmentPattern.MatchString(seg) {
			continue
		}
		c := classifySegment(seg)
		if !initialized || c.Tier > best.Tier {
			best = c
			initialized = true
		} else if c.Tier == best.Tier && c.Sensitive {
			best.Sensitive = true
		}
	}

	if !initialized {
		return Classification{Tier: Mutate, Reason: "no az command found (default deny to human review)"}
	}

	return best
}

// classifySegment classifies a single az command segment.
func classifySegment(seg string) Classification {
	for _, r := range rules {
		if r.pattern.MatchString(seg) {
			return Classification{Tier: r.tier, Reason: r.name, Sensitive: r.sensitive}
		}
	}
	return Classification{Tier: Mutate, Reason: "unknown verb (default deny to human review)"}
}
