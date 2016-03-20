[![Build Status](https://travis-ci.org/shogo82148/go-header-csv.svg?branch=master)](https://travis-ci.org/shogo82148/go-header-csv)

# go-header-csv

go-header-csv is encoder/decoder csv with header.

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

See [godoc](https://godoc.org/github.com/shogo82148/go-header-csv) for more detail.
