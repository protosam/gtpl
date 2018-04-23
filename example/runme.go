package main

import (
	"fmt"
	"github.com/protosam/gtpl"
	"log"
)

// Register handlers for specific tasks. These get ran on TPL.Parse() and TPL.Out()
func init() {
	gtpl.AddHandler("header", header_handler)
	gtpl.AddHandler("footer", footer_handler)
}

// Example of using template system!
func main() {
	log.Println("Hello TPL!")

	// Use the main.html template or die
	tpl, err := gtpl.Open("templates/main.html")
	if err != nil {
		log.Panic(err)
	}

	// Assign a global variable
	tpl.AssignGlobal("a_global_var", "Global Varaible Here")

	// Parse out the "top_body" block.
	tpl.Parse("top_body")

	// Assign a value to {foo}
	tpl.Assign("foo", "Something about foobar!")
	// Parse "some_row" which is nested in "content_body"
	tpl.Parse("content_body.some_row")

	// Assign a new value to {foo}
	tpl.Assign("foo", "Putting something else here...")
	// Parse "some_row" which is nested in "content_body"
	tpl.Parse("content_body.some_row")

	// Parse content_body
	tpl.Parse("content_body")

	// Spit out the parsed page content
	log.Println("Page Content is:")
	fmt.Print(tpl.Out(), "\n")
}

// Handler to parse out page headers
func header_handler() string {
	tpl, err := gtpl.Open("templates/overall.html")
	if err != nil {
		log.Println(err)
		return ""
	}

	tpl.Parse("header")
	return tpl.Out()
}

// Handler to parse out page footers
func footer_handler() string {
	tpl, err := gtpl.Open("templates/overall.html")
	if err != nil {
		log.Println(err)
		return ""
	}

	tpl.Parse("footer")
	return tpl.Out()
}
