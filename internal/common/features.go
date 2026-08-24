package common

import "github.com/withakedo/terraform_checkmk_provider/internal/client"

// Features represents version-dependent feature flags
// These flags control which API endpoints and features are available
// based on the detected CheckMK version.
type Features struct {
	// UseLabelGroups indicates if the label groups API is available
	// Available in CheckMK 2.3.0+
	UseLabelGroups bool

	// HasNotificationAPI indicates if the notification rule API is available
	// Available in CheckMK 2.2.0p5+
	HasNotificationAPI bool

	// SupportsRuleBulk indicates if bulk rule operations are supported
	// Available in CheckMK 2.3.0+
	SupportsRuleBulk bool

	// SupportsHostLabels indicates if host labels API is available
	// Available in CheckMK 2.1.0+
	SupportsHostLabels bool

	// HasPasswordStore indicates if password store API is available
	// Available in CheckMK 2.2.0+
	HasPasswordStore bool

	// SupportsCustomAttributes indicates if custom host attributes API is available
	// Available in CheckMK 2.0.0+
	SupportsCustomAttributes bool
}

// NewFeatures creates feature flags based on detected version
func NewFeatures(version *client.Version) *Features {
	if version == nil {
		return &Features{
			UseLabelGroups:           false,
			HasNotificationAPI:       false,
			SupportsRuleBulk:         false,
			SupportsHostLabels:       false,
			HasPasswordStore:         false,
			SupportsCustomAttributes: false,
		}
	}

	return &Features{
		UseLabelGroups:           version.AtLeast(2, 3),
		SupportsRuleBulk:         version.AtLeast(2, 3),
		HasNotificationAPI:       version.AtLeast(2, 2, 0, 5),
		HasPasswordStore:         version.AtLeast(2, 2),
		SupportsHostLabels:       version.AtLeast(2, 1),
		SupportsCustomAttributes: version.AtLeast(2, 0),
	}
}

// String returns a human-readable description of the feature flags
func (f *Features) String() string {
	return "CheckMK Feature Flags: " +
		"LabelGroups=" + boolToString(f.UseLabelGroups) +
		", NotificationAPI=" + boolToString(f.HasNotificationAPI) +
		", RuleBulk=" + boolToString(f.SupportsRuleBulk) +
		", HostLabels=" + boolToString(f.SupportsHostLabels) +
		", PasswordStore=" + boolToString(f.HasPasswordStore) +
		", CustomAttributes=" + boolToString(f.SupportsCustomAttributes)
}

func boolToString(b bool) string {
	if b {
		return "enabled"
	}
	return "disabled"
}
