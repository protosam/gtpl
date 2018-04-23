# GTPL - Golang Templates
`GTPL` is a simplified templating system that makes separation of HTML and application logic easy. This small library was created as the successor of `vision` (https://github.com/protosam/vision/). `GTPL` takes HTML that is sliced into blocks with html comments, parses out blocks as needed, and can even run registered functions.

## Security
This package doesn't provide protection from malicious HTML, CSS, or even Javascript. For most things you should be sanitizing inputs anyway, but when you begin talking about comments on blogs or even forums, you need to provide some means of formating text. Consider using the `html` and `html/template` package for handling input sanitization for html input.  
  
## The Example
If you switch to the the `example` directory you will find a basic example of how to use `GTPL`. Running it is as simple as `go run runme.go`!
