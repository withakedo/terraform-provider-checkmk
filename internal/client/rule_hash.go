package client

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"sort"
)

// RuleIdentity represents the fields that uniquely identify a rule
type RuleIdentity struct {
	Ruleset     string              `json:"ruleset"`
	Description string              `json:"description"`
	Conditions  CanonicalConditions `json:"conditions"`
}

// CanonicalConditions is a normalized form of conditions for consistent hashing
type CanonicalConditions struct {
	HostName           *HostNameCondition    `json:"host_name,omitempty"`
	HostTags           []TagCondition        `json:"host_tags,omitempty"`
	HostLabels         []LabelCondition      `json:"host_labels,omitempty"`
	HostLabelGroups    []LabelGroupCondition `json:"host_label_groups,omitempty"`
	ServiceLabels      []LabelCondition      `json:"service_labels,omitempty"`
	ServiceLabelGroups []LabelGroupCondition `json:"service_label_groups,omitempty"`
	ServiceDescription *ServiceDescCondition `json:"service_description,omitempty"`
}

// Supporting condition types
type HostNameCondition struct {
	MatchOn  []string `json:"match_on"`
	Operator string   `json:"operator"`
}

type TagCondition struct {
	Key      string `json:"key"`
	Operator string `json:"operator"`
	Value    string `json:"value,omitempty"`
}

type LabelCondition struct {
	Key      string `json:"key"`
	Operator string `json:"operator"`
	Value    string `json:"value"`
}

type LabelGroupCondition struct {
	Operator   string           `json:"operator"`
	LabelGroup []LabelCondition `json:"label_group"`
}

type ServiceDescCondition struct {
	MatchOn  []string `json:"match_on"`
	Operator string   `json:"operator"`
}

// NormalizeConditions converts conditions to canonical form for consistent hashing
func NormalizeConditions(conditions interface{}) CanonicalConditions {
	if conditions == nil {
		return CanonicalConditions{}
	}

	// Convert to map for flexible handling
	var condMap map[string]interface{}

	// Handle both map[string]interface{} and struct inputs
	switch v := conditions.(type) {
	case map[string]interface{}:
		condMap = v
	case CanonicalConditions:
		// Already canonical
		return sortCanonicalConditions(v)
	default:
		// Try to marshal and unmarshal to get map representation
		data, err := json.Marshal(conditions)
		if err != nil {
			return CanonicalConditions{}
		}
		if err := json.Unmarshal(data, &condMap); err != nil {
			return CanonicalConditions{}
		}
	}

	canonical := CanonicalConditions{}

	// Process host_name
	if hostName, ok := condMap["host_name"].(map[string]interface{}); ok {
		canonical.HostName = normalizeHostNameCondition(hostName)
	}

	// Process host_tags
	if hostTags, ok := condMap["host_tags"].([]interface{}); ok {
		canonical.HostTags = normalizeTagConditions(hostTags)
	}

	// Process host_labels
	if hostLabels, ok := condMap["host_labels"].([]interface{}); ok {
		canonical.HostLabels = normalizeLabelConditions(hostLabels)
	}

	// Process host_label_groups
	if hostLabelGroups, ok := condMap["host_label_groups"].([]interface{}); ok {
		canonical.HostLabelGroups = normalizeLabelGroupConditions(hostLabelGroups)
	}

	// Process service_labels
	if serviceLabels, ok := condMap["service_labels"].([]interface{}); ok {
		canonical.ServiceLabels = normalizeLabelConditions(serviceLabels)
	}

	// Process service_label_groups
	if serviceLabelGroups, ok := condMap["service_label_groups"].([]interface{}); ok {
		canonical.ServiceLabelGroups = normalizeLabelGroupConditions(serviceLabelGroups)
	}

	// Process service_description
	if serviceDesc, ok := condMap["service_description"].(map[string]interface{}); ok {
		canonical.ServiceDescription = normalizeServiceDescCondition(serviceDesc)
	}

	return canonical
}

// normalizeHostNameCondition normalizes a host name condition
func normalizeHostNameCondition(cond map[string]interface{}) *HostNameCondition {
	result := &HostNameCondition{}

	if operator, ok := cond["operator"].(string); ok {
		result.Operator = operator
	}

	if matchOn, ok := cond["match_on"].([]interface{}); ok {
		result.MatchOn = make([]string, 0, len(matchOn))
		for _, item := range matchOn {
			if str, ok := item.(string); ok {
				result.MatchOn = append(result.MatchOn, str)
			}
		}
		sort.Strings(result.MatchOn)
	}

	return result
}

// normalizeTagConditions normalizes tag conditions and sorts them
func normalizeTagConditions(tags []interface{}) []TagCondition {
	result := make([]TagCondition, 0, len(tags))

	for _, tag := range tags {
		if tagMap, ok := tag.(map[string]interface{}); ok {
			tc := TagCondition{}
			if key, ok := tagMap["key"].(string); ok {
				tc.Key = key
			}
			if operator, ok := tagMap["operator"].(string); ok {
				tc.Operator = operator
			}
			if value, ok := tagMap["value"].(string); ok {
				tc.Value = value
			}
			result = append(result, tc)
		}
	}

	// Sort by key, then operator, then value
	sort.Slice(result, func(i, j int) bool {
		if result[i].Key != result[j].Key {
			return result[i].Key < result[j].Key
		}
		if result[i].Operator != result[j].Operator {
			return result[i].Operator < result[j].Operator
		}
		return result[i].Value < result[j].Value
	})

	return result
}

// normalizeLabelConditions normalizes label conditions and sorts them
func normalizeLabelConditions(labels []interface{}) []LabelCondition {
	result := make([]LabelCondition, 0, len(labels))

	for _, label := range labels {
		if labelMap, ok := label.(map[string]interface{}); ok {
			lc := LabelCondition{}
			if key, ok := labelMap["key"].(string); ok {
				lc.Key = key
			}
			if operator, ok := labelMap["operator"].(string); ok {
				lc.Operator = operator
			}
			if value, ok := labelMap["value"].(string); ok {
				lc.Value = value
			}
			result = append(result, lc)
		}
	}

	// Sort by key, then operator, then value
	sort.Slice(result, func(i, j int) bool {
		if result[i].Key != result[j].Key {
			return result[i].Key < result[j].Key
		}
		if result[i].Operator != result[j].Operator {
			return result[i].Operator < result[j].Operator
		}
		return result[i].Value < result[j].Value
	})

	return result
}

// normalizeLabelGroupConditions normalizes label group conditions
func normalizeLabelGroupConditions(groups []interface{}) []LabelGroupCondition {
	result := make([]LabelGroupCondition, 0, len(groups))

	for _, group := range groups {
		if groupMap, ok := group.(map[string]interface{}); ok {
			lgc := LabelGroupCondition{}
			if operator, ok := groupMap["operator"].(string); ok {
				lgc.Operator = operator
			}
			if labelGroup, ok := groupMap["label_group"].([]interface{}); ok {
				lgc.LabelGroup = normalizeLabelConditions(labelGroup)
			}
			result = append(result, lgc)
		}
	}

	// Sort by operator, then by first label in group
	sort.Slice(result, func(i, j int) bool {
		if result[i].Operator != result[j].Operator {
			return result[i].Operator < result[j].Operator
		}
		// Compare first label in each group
		if len(result[i].LabelGroup) > 0 && len(result[j].LabelGroup) > 0 {
			return result[i].LabelGroup[0].Key < result[j].LabelGroup[0].Key
		}
		return len(result[i].LabelGroup) < len(result[j].LabelGroup)
	})

	return result
}

// normalizeServiceDescCondition normalizes a service description condition
func normalizeServiceDescCondition(cond map[string]interface{}) *ServiceDescCondition {
	result := &ServiceDescCondition{}

	if operator, ok := cond["operator"].(string); ok {
		result.Operator = operator
	}

	if matchOn, ok := cond["match_on"].([]interface{}); ok {
		result.MatchOn = make([]string, 0, len(matchOn))
		for _, item := range matchOn {
			if str, ok := item.(string); ok {
				result.MatchOn = append(result.MatchOn, str)
			}
		}
		sort.Strings(result.MatchOn)
	}

	return result
}

// sortCanonicalConditions ensures all arrays in canonical conditions are sorted
func sortCanonicalConditions(cond CanonicalConditions) CanonicalConditions {
	// Sort host_name match_on
	if cond.HostName != nil && len(cond.HostName.MatchOn) > 0 {
		sort.Strings(cond.HostName.MatchOn)
	}

	// Sort host_tags
	if len(cond.HostTags) > 0 {
		sort.Slice(cond.HostTags, func(i, j int) bool {
			if cond.HostTags[i].Key != cond.HostTags[j].Key {
				return cond.HostTags[i].Key < cond.HostTags[j].Key
			}
			if cond.HostTags[i].Operator != cond.HostTags[j].Operator {
				return cond.HostTags[i].Operator < cond.HostTags[j].Operator
			}
			return cond.HostTags[i].Value < cond.HostTags[j].Value
		})
	}

	// Sort host_labels
	if len(cond.HostLabels) > 0 {
		sort.Slice(cond.HostLabels, func(i, j int) bool {
			if cond.HostLabels[i].Key != cond.HostLabels[j].Key {
				return cond.HostLabels[i].Key < cond.HostLabels[j].Key
			}
			if cond.HostLabels[i].Operator != cond.HostLabels[j].Operator {
				return cond.HostLabels[i].Operator < cond.HostLabels[j].Operator
			}
			return cond.HostLabels[i].Value < cond.HostLabels[j].Value
		})
	}

	// Sort host_label_groups
	if len(cond.HostLabelGroups) > 0 {
		for i := range cond.HostLabelGroups {
			if len(cond.HostLabelGroups[i].LabelGroup) > 0 {
				sort.Slice(cond.HostLabelGroups[i].LabelGroup, func(a, b int) bool {
					if cond.HostLabelGroups[i].LabelGroup[a].Key != cond.HostLabelGroups[i].LabelGroup[b].Key {
						return cond.HostLabelGroups[i].LabelGroup[a].Key < cond.HostLabelGroups[i].LabelGroup[b].Key
					}
					if cond.HostLabelGroups[i].LabelGroup[a].Operator != cond.HostLabelGroups[i].LabelGroup[b].Operator {
						return cond.HostLabelGroups[i].LabelGroup[a].Operator < cond.HostLabelGroups[i].LabelGroup[b].Operator
					}
					return cond.HostLabelGroups[i].LabelGroup[a].Value < cond.HostLabelGroups[i].LabelGroup[b].Value
				})
			}
		}
		sort.Slice(cond.HostLabelGroups, func(i, j int) bool {
			if cond.HostLabelGroups[i].Operator != cond.HostLabelGroups[j].Operator {
				return cond.HostLabelGroups[i].Operator < cond.HostLabelGroups[j].Operator
			}
			if len(cond.HostLabelGroups[i].LabelGroup) > 0 && len(cond.HostLabelGroups[j].LabelGroup) > 0 {
				return cond.HostLabelGroups[i].LabelGroup[0].Key < cond.HostLabelGroups[j].LabelGroup[0].Key
			}
			return len(cond.HostLabelGroups[i].LabelGroup) < len(cond.HostLabelGroups[j].LabelGroup)
		})
	}

	// Sort service_labels
	if len(cond.ServiceLabels) > 0 {
		sort.Slice(cond.ServiceLabels, func(i, j int) bool {
			if cond.ServiceLabels[i].Key != cond.ServiceLabels[j].Key {
				return cond.ServiceLabels[i].Key < cond.ServiceLabels[j].Key
			}
			if cond.ServiceLabels[i].Operator != cond.ServiceLabels[j].Operator {
				return cond.ServiceLabels[i].Operator < cond.ServiceLabels[j].Operator
			}
			return cond.ServiceLabels[i].Value < cond.ServiceLabels[j].Value
		})
	}

	// Sort service_label_groups
	if len(cond.ServiceLabelGroups) > 0 {
		for i := range cond.ServiceLabelGroups {
			if len(cond.ServiceLabelGroups[i].LabelGroup) > 0 {
				sort.Slice(cond.ServiceLabelGroups[i].LabelGroup, func(a, b int) bool {
					if cond.ServiceLabelGroups[i].LabelGroup[a].Key != cond.ServiceLabelGroups[i].LabelGroup[b].Key {
						return cond.ServiceLabelGroups[i].LabelGroup[a].Key < cond.ServiceLabelGroups[i].LabelGroup[b].Key
					}
					if cond.ServiceLabelGroups[i].LabelGroup[a].Operator != cond.ServiceLabelGroups[i].LabelGroup[b].Operator {
						return cond.ServiceLabelGroups[i].LabelGroup[a].Operator < cond.ServiceLabelGroups[i].LabelGroup[b].Operator
					}
					return cond.ServiceLabelGroups[i].LabelGroup[a].Value < cond.ServiceLabelGroups[i].LabelGroup[b].Value
				})
			}
		}
		sort.Slice(cond.ServiceLabelGroups, func(i, j int) bool {
			if cond.ServiceLabelGroups[i].Operator != cond.ServiceLabelGroups[j].Operator {
				return cond.ServiceLabelGroups[i].Operator < cond.ServiceLabelGroups[j].Operator
			}
			if len(cond.ServiceLabelGroups[i].LabelGroup) > 0 && len(cond.ServiceLabelGroups[j].LabelGroup) > 0 {
				return cond.ServiceLabelGroups[i].LabelGroup[0].Key < cond.ServiceLabelGroups[j].LabelGroup[0].Key
			}
			return len(cond.ServiceLabelGroups[i].LabelGroup) < len(cond.ServiceLabelGroups[j].LabelGroup)
		})
	}

	// Sort service_description match_on
	if cond.ServiceDescription != nil && len(cond.ServiceDescription.MatchOn) > 0 {
		sort.Strings(cond.ServiceDescription.MatchOn)
	}

	return cond
}

// GenerateRuleHash creates a deterministic hash for a rule
func GenerateRuleHash(ruleset, description string, conditions interface{}) string {
	canonical := NormalizeConditions(conditions)
	identity := RuleIdentity{
		Ruleset:     ruleset,
		Description: description,
		Conditions:  canonical,
	}

	// Marshal to JSON for deterministic representation
	data, err := json.Marshal(identity)
	if err != nil {
		// Fallback to simple hash if marshaling fails
		data = []byte(ruleset + description)
	}

	// Generate SHA256 hash
	hash := sha256.Sum256(data)

	// Return first 16 bytes as 32 hex characters
	return hex.EncodeToString(hash[:16])
}
