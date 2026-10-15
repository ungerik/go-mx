package mcp_test

import (
	"context"
	"log"
	"net/http"

	sdk "github.com/modelcontextprotocol/go-sdk/mcp"

	mxmcp "github.com/ungerik/go-mx/mcp"
	"github.com/ungerik/go-mx/web"
)

// The site's pages, read-only, next to a tool of the caller's own, served
// statelessly over HTTP behind the guard.
func ExampleAddPageTools() {
	site := &web.Site{
		BaseURL: "https://example.com",
		Title:   "Example",
		Lang:    "en",
		Sources: []web.PageSource{&web.GlobPageSource{Dir: "content", Pattern: "*.md"}},
	}

	server := sdk.NewServer(&sdk.Implementation{Name: "example", Version: "v1.0.0"}, nil)
	err := mxmcp.AddPageTools(server, site, site, mxmcp.PageToolsOptions{
		OnError: func(ctx context.Context, err error) { log.Println("mcp:", err) },
	})
	if err != nil {
		log.Fatal(err)
	}

	// A tool of the caller's own on the same server.
	type openingHoursOutput struct {
		Hours string `json:"hours" jsonschema:"the opening hours"`
	}
	sdk.AddTool(server, &sdk.Tool{
		Name:        "opening_hours",
		Description: "The opening hours of the shop.",
		Annotations: &sdk.ToolAnnotations{ReadOnlyHint: true},
	}, func(ctx context.Context, request *sdk.CallToolRequest, input struct{}) (*sdk.CallToolResult, openingHoursOutput, error) {
		return nil, openingHoursOutput{Hours: "Mon-Fri 9:00-18:00"}, nil
	})

	handler := sdk.NewStreamableHTTPHandler(
		func(*http.Request) *sdk.Server { return server },
		&sdk.StreamableHTTPOptions{Stateless: true, JSONResponse: true},
	)
	mux := http.NewServeMux()
	mux.Handle("/mcp", mxmcp.GuardHandler(handler, mxmcp.GuardOptions{
		AllowedOrigins: []string{"https://example.com"},
	}))
	// http.ListenAndServe(":8080", mux)
}
