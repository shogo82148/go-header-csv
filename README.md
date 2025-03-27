![test](https://github.com/shogo82148/go-header-csv/workflows/test/badge.svg)
[![PkgGoDev](https://pkg.go.dev/badge/github.com/shogo82148/go-header-csv)](https://pkg.go.dev/github.com/shogo82148/go-header-csv)

# SYNOPSIS

go-header-csv is encoder/decoder csv with a header.

The following is an example of csv.

```go
func ExampleDecoder_DecodeAll() {
	in := `name,text
Ed,Knock knock.
Sam,Who's there?
Ed,Go fmt.
Sam,Go fmt who?
Ed,Go fmt yourself!
`
	out := []struct {
		Name string `csv:"name"`
		Text string `csv:"text"`
	}{}

	buf := bytes.NewBufferString(in)
	dec := headercsv.NewDecoder(buf)
	dec.DecodeAll(&out)

	for _, v := range out {
		fmt.Printf("%3s: %s\n", v.Name, v.Text)
	}
	// Output:
	//  Ed: Knock knock.
	// Sam: Who's there?
	//  Ed: Go fmt.
	// Sam: Go fmt who?
	//  Ed: Go fmt yourself!
}
```

```go
func ExampleEncoder_EncodeAll() {
	in := []struct {
		Name string `csv:"name"`
		Text string `csv:"text"`
	}{
		{"Ed", "Knock knock."},
		{"Sam", "Who's there?"},
		{"Ed", "Go fmt."},
		{"Sam", "Go fmt who?"},
		{"Ed", "Go fmt yourself!"},
	}

	enc := headercsv.NewEncoder(os.Stdout)
	enc.EncodeAll(in)
	enc.Flush()

	// Output:
	// name,text
	// Ed,Knock knock.
	// Sam,Who's there?
	// Ed,Go fmt.
	// Sam,Go fmt who?
	// Ed,Go fmt yourself!
}
```

## Related Works

- [gocarina/gocsv](https://github.com/gocarina/gocsv)
- [jszwec/csvutil](https://github.com/jszwec/csvutil)
- [yunabe/easycsv](https://github.com/yunabe/easycsv)
