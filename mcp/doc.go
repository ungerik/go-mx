// Package mcp exposes the public pages of a [web.Site] to AI agents over the
// Model Context Protocol, as read-only tools on a server of the official Go SDK
// (github.com/modelcontextprotocol/go-sdk).
//
// The package is composable by design: the caller owns the SDK server and the
// HTTP mux, and this package only adds tools to the one and offers a guard for
// the other. [AddPageTools] registers search_pages, list_pages and get_page,
// which return the Markdown alternates (see [web.Site.MarkdownAlternate]) of
// the [web.Page.Indexable] pages of a source and nothing else. The same server
// can carry any other tools the caller adds. [GuardHandler] wraps the HTTP
// handler serving the server with the checks every public MCP endpoint needs.
//
// Nothing is exposed without the caller passing a source: there is no default
// source, and the tools never return a page that is not public or has no
// Markdown alternate.
//
// It is a module of its own, so that the SDK and its dependencies are only
// pulled into programs that serve MCP. Import it under another name than the
// SDK package, which is called mcp too:
//
//	import (
//	    "github.com/modelcontextprotocol/go-sdk/mcp"
//	    mxmcp "github.com/ungerik/go-mx/mcp"
//	)
package mcp
