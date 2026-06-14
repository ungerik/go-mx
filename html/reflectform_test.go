package html

import (
	"context"
	"os"

	"github.com/ungerik/go-mx"
)

func ExampleReflectFormComponents() {
	type UserDetails struct {
		Name          string `input:"name=name"`
		Email         string `input:"type=email"`
		Ingore        string `input:"-"`
		Age           int
		TermsAccepted bool `label:"Terms and conditions accepted"`
	}

	Form(Action("/submit"), MethodPOST,
		ReflectFormComponents(UserDetails{
			Name: "John Doe",
		}),
		InputTypeSubmit(Value("Submit")),
	).Render(context.Background(), mx.NewCheckedWriter(os.Stdout).WithIndent("", "  "))

	// Output:
	// <form action="/submit" method="post">
	//   <label>Name:
	//     <input name="name" value="John Doe"/>
	//   </label>
	//   <label>Email:
	//     <input type="email" name="Email"/>
	//   </label>
	//   <label>Age:
	//     <input name="Age" type="number"/>
	//   </label>
	//   <label>
	//     <input name="TermsAccepted" type="checkbox"/>Terms and conditions accepted
	//   </label>
	//   <input type="submit" value="Submit"/>
	// </form>
}
