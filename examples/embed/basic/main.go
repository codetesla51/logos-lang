package main

import (
	"fmt"

	"github.com/codetesla51/logos/logos"
)

func main() {
	vm := logos.NewWithConfig(logos.SandboxConfig{
		AllowFileIO:  false,
		AllowNetwork: false,
		AllowShell:   false,
		AllowExit:    false,
	})

	vm.Register("greet", func(args ...logos.Object) logos.Object {
		name := args[0].(*logos.String).Value
		return &logos.String{Value: "hello from go " + name + "!"}

	})
	vm.Register("add", func(args ...logos.Object) logos.Object {
		a := args[0].(*logos.Integer).Value
		b := args[1].(*logos.Integer).Value
		return &logos.Integer{Value: a + b}
	})
	script := `
	let name = "uthma"
	print(greet(name))
	let sum = add(5, 10)
	print("sum:", sum)
	let final = "done"
	`
	err := vm.Run(script)
	if err != nil {
		fmt.Println(err)
		return
	}
	finalValue := vm.GetVar("final")
	fmt.Println("final value:", finalValue)

	err = vm.Run(`
        let res = fileRead("secret.txt")
        print(res)
    `)

	if err != nil {
		fmt.Println("blocked:", err)
	}
}
