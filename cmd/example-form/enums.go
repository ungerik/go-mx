//go:generate go -C ../../tools tool go-enum ../cmd/example-form/$GOFILE

package main

import "fmt"

// Tier is a small enum used to demonstrate the enum dispatch path. go-enum
// generates its EnumStrings method, which makes the field a select, and its
// Validate method, which rejects a submitted value that is none of the options.
type Tier string //#enum

const (
	// TierFree is the free plan.
	TierFree Tier = "free"
	// TierTeam is the plan for teams.
	TierTeam Tier = "team"
	// TierEnterprise is the plan that requires SSO.
	TierEnterprise Tier = "enterprise"
)

// Valid indicates if t is any of the valid values for Tier
func (t Tier) Valid() bool {
	switch t {
	case
		TierFree,
		TierTeam,
		TierEnterprise:
		return true
	}
	return false
}

// Validate returns an error if t is none of the valid values for Tier
func (t Tier) Validate() error {
	if !t.Valid() {
		return fmt.Errorf("invalid value %#v for type main.Tier", t)
	}
	return nil
}

// Enums returns all valid values for Tier
func (Tier) Enums() []Tier {
	return []Tier{
		TierFree,
		TierTeam,
		TierEnterprise,
	}
}

// EnumStrings returns all valid values for Tier as strings
func (Tier) EnumStrings() []string {
	return []string{
		"free",
		"team",
		"enterprise",
	}
}

// String implements the fmt.Stringer interface for Tier
func (t Tier) String() string {
	return string(t)
}

// Feature is the enum used to populate the multi-select / set field. Each
// element of the set is checked with the generated Validate method.
type Feature string //#enum

const (
	// FeatureSSO is single sign-on.
	FeatureSSO Feature = "sso"
	// FeatureAuditLog is the audit log.
	FeatureAuditLog Feature = "audit-log"
	// FeatureAdvancedAnalytics is the advanced analytics dashboard.
	FeatureAdvancedAnalytics Feature = "advanced-analytics"
	// FeatureExports is the data export.
	FeatureExports Feature = "exports"
)

// Valid indicates if f is any of the valid values for Feature
func (f Feature) Valid() bool {
	switch f {
	case
		FeatureSSO,
		FeatureAuditLog,
		FeatureAdvancedAnalytics,
		FeatureExports:
		return true
	}
	return false
}

// Validate returns an error if f is none of the valid values for Feature
func (f Feature) Validate() error {
	if !f.Valid() {
		return fmt.Errorf("invalid value %#v for type main.Feature", f)
	}
	return nil
}

// Enums returns all valid values for Feature
func (Feature) Enums() []Feature {
	return []Feature{
		FeatureSSO,
		FeatureAuditLog,
		FeatureAdvancedAnalytics,
		FeatureExports,
	}
}

// EnumStrings returns all valid values for Feature as strings
func (Feature) EnumStrings() []string {
	return []string{
		"sso",
		"audit-log",
		"advanced-analytics",
		"exports",
	}
}

// String implements the fmt.Stringer interface for Feature
func (f Feature) String() string {
	return string(f)
}
