package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
)

func main() {
	f, err := ioutil.ReadFile("./examples/nodejs/function.zip")
	if err != nil {
		panic(err)
	}
	fmt.Println(http.DetectContentType(f))
	f, err = ioutil.ReadFile("./examples/nodejs/helloget.js")
	if err != nil {
		panic(err)
	}
	fmt.Println(http.DetectContentType(f))
}
