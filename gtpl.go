/*****************************************************************/
/* gtpl.go -- A simplified templating system that makes          */
/* separation of HTML and application logic easy.                */
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

package gtpl

import (
	"errors"
	"fmt"
	"io/ioutil"
	"regexp"
	"strings"
)

// Template handler functions that can be called template files
var handlers = make(map[string]func() string)

// Globally assigned variables.
var globalassignments = make(map[string]string)

// Simple structure to house our blocks and local assignments.
type TPL struct {
	LocalAssignments map[string]string
	blocks           map[string]string
}

// Open(variadic <filePath string | fileStream []bytes>) -- Processes a GTPL file from the file at filePath or contained in []bytes fileStream
// Input:
//        filePath   string			-- MANDATORY if fileStream is not provided
//        fileStream []byte			-- MANDATORY if filePath is not provided
// Output:
//        TPL object				-- Contains TPL data about GTPL file
//        error                     -- Returned if parser fails to parse TPL data or paramaters are wrong
func Open(vArgs ...interface{}) (TPL, error) {
	filePath, fileStream, pErrs := openParams(vArgs...)

	if pErrs != nil {
		return TPL{}, pErrs
	}

	tpl := TPL{}

	var fErr error
	if filePath != "" {
		fileStream, fErr = ioutil.ReadFile(filePath)
	}

	if fErr != nil {
		return tpl, fErr
	}

	// Setup the struct
	tpl.blocks = make(map[string]string)
	tpl.LocalAssignments = make(map[string]string)

	// Store raw content into output for processing
	tpl.blocks["[_GTPL_ROOT_]"] = string(fileStream)

	err := tpl.preprocess("")
	if err != nil {
		return tpl, errors.New(fmt.Sprintf("gtpl parser failure: %s", err))
	}

	return tpl, nil
}

// openParams(vArgs ...interface{}) -- Validates variadic parameters for Open()
// Input:
//        vArgs ...interface{}		-- List of variables
// Output:
//        filePath   string			-- untouched
//     OR fileStream []byte			-- untouched
//        err         error			-- set if incorrect number of arguments are passed
func openParams(vArgs ...interface{}) (filePath string, fileStream []byte, err error) {
	filePath, fileStream = "", []byte{}
	
	// Verify enough parameters
	if 1 > len(vArgs) {
		err = errors.New("not enough parameters")
	} else if 1 < len(vArgs) {
		err = errors.New("too many parameters")
	}

	// Validate and unload arguments
	var check1 bool
	var check2 bool
	for _,param := range vArgs {
		fileStream, check1 = param.([]byte)
		filePath, check2 = param.(string)
		if !check1 && !check2 {
			err = errors.New(fmt.Sprintf("invalid type: %T", param))
			return
		}
	}

	return
}

// Add a new handler
func AddHandler(name string, fn func() string) {
	handlers[name] = fn
}

// Assign a new global variable's value
func (tpl *TPL) AssignGlobal(variable string, value string) {
	globalassignments[variable] = sanitize(value)
}

// Assign a new local variable's value
func (tpl *TPL) Assign(variable string, value string) {
	tpl.LocalAssignments[variable] = sanitize(value)
}

// Parse a block. Blocks of code need to be parsed from most inner, to outter.
func (tpl *TPL) Parse(block_name string) {
	// Add the root block
	block_name = "[_GTPL_ROOT_]." + block_name

	// Cut off the last block name to get the parent block name
	cut_index := strings.LastIndex(block_name, ".")
	parent_block_name := block_name[:cut_index]

	// Store raw content
	content_results := tpl.blocks[block_name] + parent_block_name

	content_results = tpl.assignments(content_results)

	// Run handlers
	content_results = tpl.handlers(content_results)

	// Update the block in the map
	tpl.blocks[parent_block_name] = strings.Replace(tpl.blocks[parent_block_name], parent_block_name, content_results, 1)
}

// Provide output from the most parent blocks
func (tpl *TPL) Out() string {
	// Prepwork for cleanup
	place_holder_pattern := regexp.MustCompile(regexp.QuoteMeta("[_GTPL_ROOT_].") + "[A-Za-z0-9_\\-\\.]+")

	// Run handlers
	tpl.blocks["[_GTPL_ROOT_]"] = tpl.handlers(tpl.blocks["[_GTPL_ROOT_]"])

	// Remove all the position place holders
	tpl.blocks["[_GTPL_ROOT_]"] = string(place_holder_pattern.ReplaceAll([]byte(tpl.blocks["[_GTPL_ROOT_]"]), []byte("")))

	// Clean up random whitespacing
	re := regexp.MustCompile(`(?m)^\s*$[\r\n]*|[\r\n]+\s+\z`)
	tpl.blocks["[_GTPL_ROOT_]"] = re.ReplaceAllString(tpl.blocks["[_GTPL_ROOT_]"], "")

	return desanitize(tpl.blocks["[_GTPL_ROOT_]"])
}

// Preprocesses the entire tree of blocks
func (tpl *TPL) preprocess(parent_block_name string) error {
	// Begin processing the blocks
	begin_pattern := regexp.MustCompile("<!-- block: ([A-Za-z0-9_-]+) -->")
	var raw_block_name []string

	// Replace the block with placeholders
	if parent_block_name == "" {
		// Generate a root block name
		parent_block_name = "[_GTPL_ROOT_]"
	}

	raw_block_name = begin_pattern.FindStringSubmatch(tpl.blocks[parent_block_name])

	// No blocks found
	if raw_block_name == nil {
		return nil
	}

	for raw_block_name != nil {

		// Get the block's content
		block_pattern := regexp.MustCompile("<!-- block: " + raw_block_name[1] + " -->(?ms:(.*?))<!-- /block: " + raw_block_name[1] + " -->")
		block_content := block_pattern.FindStringSubmatch(tpl.blocks[parent_block_name])

		// No match was found, throw an error!
		if block_content == nil {
			return errors.New("Failed to find a match for block: " + raw_block_name[1])
		}

		// active block name
		active_block_name := parent_block_name + "." + raw_block_name[1]

		// Store found new block in the hashtable
		tpl.blocks[active_block_name] = block_content[1]

		// Tokenize the newly stored block as a reference in the parent
		tpl.blocks[parent_block_name] = string(block_pattern.ReplaceAll([]byte(tpl.blocks[parent_block_name]), []byte(active_block_name)))

		// parse sub blocks
		tpl.preprocess(active_block_name)

		// Next search
		raw_block_name = begin_pattern.FindStringSubmatch(tpl.blocks[parent_block_name])
	}

	return nil
}

// Replace variable tokens with values
func (tpl *TPL) assignments(content_results string) string {
	// Parse global variables in the content
	for variable, value := range globalassignments {
		content_results = strings.Replace(content_results, "{"+variable+"}", value, -1)
	}

	// Parse local variables in the content
	for variable, value := range tpl.LocalAssignments {
		content_results = strings.Replace(content_results, "{"+variable+"}", value, 1)
		delete(tpl.LocalAssignments, variable)
	}
	return content_results
}

// Replace handler tokens with handler results
func (tpl *TPL) handlers(content_results string) string {
	// Run handlers against the content
	handler_pattern := regexp.MustCompile("<!-- handler: ([A-Za-z0-9_-]+) -->")
	handler_search := handler_pattern.FindStringSubmatch(content_results)

	// Loop and do the handler functions
	for handler_search != nil {
		handler_comment := handler_search[0]
		handler_name := handler_search[1]
		handler_result := ""

		if _, ok := handlers[handler_name]; ok {
			handler_result = handlers[handler_name]()
		}

		content_results = strings.Replace(content_results, handler_comment, handler_result, -1)
		handler_search = handler_pattern.FindStringSubmatch(content_results)
	}
	return content_results
}

// Prevent template injection
func sanitize(content string) string {
	content = strings.Replace(content, "[_GTPL_ROOT_]", "[\\_GTPL_ROOT_]", -1)
	content = strings.Replace(content, "<!--", "<!--\\", -1)
	content = strings.Replace(content, "{", "{\\", -1)
	return content
}

// Remove sanitizations...
func desanitize(content string) string {
	content = strings.Replace(content, "[\\_GTPL_ROOT_]", "[_GTPL_ROOT_]", -1)
	content = strings.Replace(content, "<!--\\", "<!--", -1)
	content = strings.Replace(content, "{\\", "{", -1)
	return content
}
