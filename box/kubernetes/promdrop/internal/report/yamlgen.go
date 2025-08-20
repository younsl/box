package report

import (
	"fmt"
	"os"
	"regexp"
	"sort"
	"strings"
	"text/tabwriter"

	m "younsl/promdrop/internal/metrics"

	"gopkg.in/yaml.v3"
)

// --- Custom YAML Types ---

// SingleQuotedString is a custom type to force single quotes during YAML marshalling.
type SingleQuotedString string

// MarshalYAML implements the yaml.Marshaler interface for SingleQuotedString.
func (s SingleQuotedString) MarshalYAML() (interface{}, error) {
	// Create a yaml.Node explicitly setting the style to SingleQuoted
	node := yaml.Node{
		Kind:  yaml.ScalarNode,
		Tag:   "!!str",
		Value: string(s),
		Style: yaml.SingleQuotedStyle,
	}
	// Return the node value, not a pointer
	return node, nil
}

// FlowStringSlice forces flow style (e.g., [item1, item2]) for string slices.
type FlowStringSlice []string

func (f FlowStringSlice) MarshalYAML() (interface{}, error) {
	seqNode := yaml.Node{
		Kind:  yaml.SequenceNode,
		Tag:   "!!seq",        // Standard sequence tag
		Style: yaml.FlowStyle, // Force flow style: [item1, item2]
	}
	for _, item := range f {
		itemNode := yaml.Node{
			Kind:  yaml.ScalarNode,
			Tag:   "!!str",
			Value: item,
		}
		seqNode.Content = append(seqNode.Content, &itemNode)
	}
	return seqNode, nil
}

// --- YAML Structures ---

// RelabelRule defines the structure for a single metric_relabel_configs entry.
type RelabelRule struct {
	// Use FlowStringSlice for source_labels to get [__name__] format
	SourceLabels FlowStringSlice    `yaml:"source_labels"`
	Regex        SingleQuotedString `yaml:"regex"`
	Action       string             `yaml:"action"`
}

// --- YAML Rule Generation Logic ---

// prefixRegex helps extract the prefix before the first underscore.
var prefixRegex = regexp.MustCompile(`^([^_]+)_`)

// groupMetricsByPrefix groups metrics by their prefix.
func groupMetricsByPrefix(metrics []string) map[string][]string {
	groups := make(map[string][]string)
	for _, metric := range metrics {
		match := prefixRegex.FindStringSubmatch(metric)
		var prefix string
		if len(match) > 1 {
			prefix = match[1]
		} else {
			prefix = "other" // Use "other" if no prefix
		}
		groups[prefix] = append(groups[prefix], metric)
	}
	// Ensure stable output by sorting metrics within each prefix group
	for prefix := range groups {
		sort.Strings(groups[prefix])
	}
	return groups
}

// GenerateRelabelRules generates a slice of RelabelRule structs and group info.
// Changed return type from string to []RelabelRule.
func GenerateRelabelRules(metrics []string, maxRegexLength int) ([]RelabelRule, []m.GroupInfo) {
	groups := groupMetricsByPrefix(metrics)
	var relabelRules []RelabelRule
	var groupInfoList []m.GroupInfo

	var prefixes []string
	for p := range groups {
		prefixes = append(prefixes, p)
	}
	sort.Strings(prefixes)

	for _, prefix := range prefixes {
		metricsGroup := groups[prefix] // Avoid shadowing outer metrics variable
		var chunks [][]string
		var currentChunk []string
		currentLength := 0

		for _, metric := range metricsGroup {
			metricLength := len(metric) + 1

			if currentLength+metricLength > maxRegexLength && len(currentChunk) > 0 {
				chunks = append(chunks, currentChunk)
				currentChunk = nil
				currentLength = 0
			}
			currentChunk = append(currentChunk, metric)
			currentLength += metricLength
		}
		if len(currentChunk) > 0 {
			chunks = append(chunks, currentChunk)
		}

		patternInfo := fmt.Sprintf("%s_*", prefix)
		if prefix == "other" {
			patternInfo = "(no prefix)"
		}

		for i, chunk := range chunks {
			partStr := fmt.Sprintf("%d/%d", i+1, len(chunks))
			regexContent := strings.Join(chunk, "|")

			rule := RelabelRule{
				// Cast source labels to the custom FlowStringSlice type
				SourceLabels: FlowStringSlice([]string{"__name__"}),
				Regex:        SingleQuotedString(regexContent),
				Action:       "drop",
			}
			relabelRules = append(relabelRules, rule)

			groupInfoList = append(groupInfoList, m.GroupInfo{
				Prefix:  prefix,
				Pattern: patternInfo,
				Part:    partStr,
				Count:   len(chunk),
			})
		}
	}

	return relabelRules, groupInfoList
}

// --- Group Summary Printing ---

// PrintGroupSummary prints the per-job group summary using tabwriter.
func PrintGroupSummary(groupInfo []m.GroupInfo) {
	if len(groupInfo) == 0 {
		return
	}
	fmt.Println("  [Metric Group Summary]")
	w := tabwriter.NewWriter(os.Stdout, 0, 8, 2, ' ', 0)
	fmt.Fprintln(w, "  PREFIX\tPATTERN\tPART\tMETRIC COUNT\t")
	totalProcessedInYAML := 0
	for _, info := range groupInfo {
		fmt.Fprintf(w, "  %s\t%s\t%s\t%d\t\n", info.Prefix, info.Pattern, info.Part, info.Count)
		totalProcessedInYAML += info.Count
	}
	w.Flush()
	fmt.Printf("  Total metrics included in YAML rules: %d\n", totalProcessedInYAML)
}
