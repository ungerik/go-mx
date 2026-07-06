// Command example-form runs a synthetic admin form to dogfood the
// [mx.ReflectFormHandler] dispatch table.
//
// The struct [Profile] exercises every FieldKind shadcn.FieldDecider
// handles today (string, number, bool, date, enum, enum-set, hidden,
// readonly, sensitive, nullable) plus a section grouping. The
// in-memory store treats a single record at /admin/profile.
//
//	go run ./cmd/example-form
//
// Then browse to http://localhost:8080/admin/profile.
package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/ungerik/go-mx"
	"github.com/ungerik/go-mx/shadcn"
)

// Tier is a small enum used to demonstrate the enum dispatch path.
type Tier string

// EnumStrings returns the string values of all valid Tier enum constants.
func (Tier) EnumStrings() []string {
	return []string{"free", "team", "enterprise"}
}

// Feature is the enum used to populate the multi-select / set field.
type Feature string

// EnumStrings returns the string values of all valid Feature enum constants.
func (Feature) EnumStrings() []string {
	return []string{"sso", "audit-log", "advanced-analytics", "exports"}
}

// AccountingInfo groups a handful of fields into one section so the
// example exercises the section/nested walk path.
type AccountingInfo struct {
	VATNumber string `form:"label=VAT number"`
	BillingTo string
	Currency  string `form:"widget=select,options=currencies"`
}

// LineItem is one row of the repeatable invoice-line editor, the
// canonical form:"repeatable" use case.
type LineItem struct {
	Description string  `form:"required,placeholder=Line description"`
	Quantity    int     `form:"min=1"`
	UnitPrice   float64 `form:"min=0,step=0.01"`
}

// Profile is the synthetic admin record. Every tag combination below
// is intentional and round-trips through the shadcn decider.
type Profile struct {
	ID       string `form:"hidden"`
	Name     string `form:"required,placeholder=Full name"`
	Email    string `form:"required,widget=email"`
	Bio      string `form:"widget=textarea,help=Short bio for the public profile"`
	Age      int    `form:"min=13,max=120,step=1"`
	Active   bool
	Tier     Tier
	Features map[Feature]struct{}
	Joined   time.Time      `form:"readonly"`
	Password string         `form:"widget=password,sensitive"`
	Account  AccountingInfo `form:"section=Accounting"`
	Lines    []LineItem     `form:"repeatable,label=Invoice lines"`
}

// store is the in-memory single-row "database" backing the form.
type store struct {
	mu sync.Mutex
	p  Profile
}

// Load returns a copy of the single stored Profile record.
func (s *store) Load(_ context.Context) (*Profile, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	cp := s.p
	return &cp, nil
}

// Save validates the cross-field invariant that the "enterprise" tier
// requires the SSO feature, returning a [mx.FieldErrors] if it is
// violated, and otherwise stores p as the single Profile record.
func (s *store) Save(_ context.Context, p *Profile) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	// Cross-field invariant: "enterprise" tier requires SSO. This
	// demonstrates the FieldErrors -> inline error routing path for a
	// rule the framework's per-field chain can't express.
	if p.Tier == "enterprise" {
		if _, ok := p.Features["sso"]; !ok {
			return mx.FieldErrors{
				"Features": errors.New("enterprise tier requires SSO"),
			}
		}
	}
	s.p = *p
	return nil
}

func main() {
	addr := flag.String("addr", ":8080", "listen address")
	flag.Parse()

	s := &store{
		p: Profile{
			ID:       "user-1",
			Name:     "Ada Lovelace",
			Email:    "ada@example.com",
			Bio:      "Original computing pioneer.",
			Age:      36,
			Active:   true,
			Tier:     "team",
			Features: map[Feature]struct{}{"audit-log": {}, "sso": {}},
			Joined:   time.Date(2026, 1, 15, 9, 0, 0, 0, time.UTC),
			Account: AccountingInfo{
				VATNumber: "GB-12345",
				BillingTo: "Accounts Payable, Acme Corp.",
				Currency:  "USD",
			},
			Lines: []LineItem{
				{Description: "Consulting", Quantity: 10, UnitPrice: 150},
				{Description: "Hosting", Quantity: 1, UnitPrice: 29.99},
			},
		},
	}

	mux := http.NewServeMux()
	mux.Handle("/admin/profile", mx.ReflectFormHandler(s.Load, s.Save))
	handler := mx.Middleware(shadcn.FieldDecider)(mux)

	fmt.Printf("listening on %s\n", *addr)
	fmt.Printf("open http://localhost%s/admin/profile\n", *addr)
	log.Fatal(http.ListenAndServe(*addr, handler))
}
