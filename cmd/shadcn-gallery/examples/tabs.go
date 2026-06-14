package examples

import (
	"github.com/ungerik/go-mx"
	"github.com/ungerik/go-mx/html"
	"github.com/ungerik/go-mx/shadcn"
)

func TabsDemo() mx.Component {
	return shadcn.Tabs("account-tabs", html.Class("w-full max-w-md"),
		shadcn.TabsList(html.Class("grid w-full grid-cols-2"),
			shadcn.TabsTrigger("account-tabs", "account", true, "Account"),
			shadcn.TabsTrigger("account-tabs", "password", false, "Password"),
		),
		shadcn.TabsContent("account-tabs", "account", true,
			html.PClass("p-2 text-sm text-muted-foreground",
				"Make changes to your account here. Click save when you're done."),
		),
		shadcn.TabsContent("account-tabs", "password", false,
			html.PClass("p-2 text-sm text-muted-foreground",
				"Change your password here. After saving, you'll be logged out."),
		),
	)
}
