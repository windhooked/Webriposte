// To compile this program run:
//   go build -buildmode=c-archive exportgo.go

package main

import "C"
import "fmt"

//export GetIntFromDLL
func GetIntFromDLL() int32 {
	return 42
}

//export PrintHello
func PrintHello(name string) {
	fmt.Printf("From DLL: Hello, %s!\n", name)
}

//export PrintBye
func PrintBye() {
	fmt.Println("From DLL: Bye!")
}

func main() {
	// Need a main function to make CGO compile package as C shared library
}
