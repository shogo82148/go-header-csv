package headercsv_test

import (
	"bytes"
	"fmt"
	"os"

	"github.com/shogo82148/go-header-csv"
)

func ExampleEncoder_Encode() {
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
	enc.Encode(in)
	enc.Flush()

	// Output:
	// name,text
	// Ed,Knock knock.
	// Sam,Who's there?
	// Ed,Go fmt.
	// Sam,Go fmt who?
	// Ed,Go fmt yourself!
}

func ExampleDecoder_Decode() {
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
	dec.Decode(&out)

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
