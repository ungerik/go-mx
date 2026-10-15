package schemaorg_test

import (
	"testing"

	"github.com/ungerik/go-mx/web/schemaorg"
)

func TestService(t *testing.T) {
	requireJSON(t,
		`{"@context":"https://schema.org","@type":"Service","name":"Code review"}`,
		schemaorg.NewService("Code review"),
	)

	provider := schemaorg.NewOrganization("Example")
	provider.Context = ""
	provider.URL = "https://example.com/"

	austria := schemaorg.NewThing("Country", "Austria")
	austria.Context = ""

	hourly := schemaorg.NewOffer(schemaorg.Price(15000), "EUR")
	hourly.Context = ""
	hourly.Name = "Per hour"
	daily := schemaorg.NewOffer(schemaorg.Price(100000), "EUR")
	daily.Context = ""
	daily.Name = "Per day"

	service := schemaorg.NewService("Code review")
	service.Description = "A review of a Go code base by two senior engineers."
	service.ServiceType = "Software consulting"
	service.Provider = provider
	service.AreaServed = []any{austria, "Germany"}
	service.Offers = []*schemaorg.Offer{hourly, daily}
	service.Category = "Consulting"
	requireGolden(t, "service", service)
}
