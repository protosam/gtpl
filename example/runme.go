/*****************************************************************/
/* runme.go -- An example program using GTPL.                    */
/*                                                               */
/*---------------------------------------------------------------*/
/* Copyright (c) 2018 Sam                                        */
/* Copyright (c) 2022 Matt Rienzo                                */
/*                                                               */
/* MIT Licensed:                                                 */
/* Permission is hereby granted, free of charge, to any person   */
/* obtaining a copy of this software and associated documentation*/
/* files (the "Software"), to deal in the Software without       */
/* restriction, including without limitation the rights to use,  */
/* copy, modify, merge, publish, distribute, sublicense, and/or  */
/* sell copies of the Software, and to permit persons to whom the*/
/* Software is furnished to do so, subject to the following      */
/* conditions:                                                   */
/*                                                               */
/* The above copyright notice and this permission notice shall   */
/* be included in all copies or substantial portions of the      */
/* Software.                                                     */
/*                                                               */
/* THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY     */
/* KIND, EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE    */
/* WARRANTIES OF MERCHANTABILITY, FITNESS FOR A PARTICULAR       */
/* PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR */
/* COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER   */
/* LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR          */
/* OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE     */
/* SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.        */
/*****************************************************************/

package main

import (
	"fmt"
	"github.com/casnix/gtpl"
	"io/ioutil"
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
	// Pass file directly to gtpl.Open() as a byte slice
	// instead of having gtpl.Open() read the file
	mainData, bErr := ioutil.ReadFile("templates/main.html")
	if bErr != nil {
		log.Panic(bErr)
	}
	tpl, err := gtpl.Open(mainData)
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
	// Pass filename as string to gtpl.Open()
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
