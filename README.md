![test](https://github.com/shogo82148/go-header-csv/workflows/test/badge.svg)
[![PkgGoDev](https://pkg.go.dev/badge/github.com/shogo82148/go-header-csv)](https://pkg.go.dev/github.com/shogo82148/go-header-csv)

# go-header-csv

go-header-csv is encoder/decoder csv with a header.

The following is an example of csv.

``` plain
name,text
Ed,Knock knock.
Sam,Who's there?
Ed,Go fmt.
Sam,Go fmt who?
Ed,Go fmt yourself!
```

go-header-csv treats first line (`name,text`) as the names of fields,
and decode into a golang struct.

``` go
[]struct {
  Name string `csv:"name"`
  Text string `csv:"text"`
} {
  {"Ed", "Knock knock."},
  {"Sam", "Who's there?"},
  {"Ed", "Go fmt."},
  {"Sam", "Go fmt who?"},
  {"Ed", "Go fmt yourself!"},
}
```
