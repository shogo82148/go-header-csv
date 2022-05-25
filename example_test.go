package headercsv_test

import (
	"bytes"
	"fmt"
	"os"

	headercsv "github.com/shogo82148/go-header-csv"
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

func ExampleEncoder_EncodeRecord() {
	in := struct {
		Name string `csv:"name"`
		Text string `csv:"text"`
	}{"Ed", "Knock knock."}

	enc := headercsv.NewEncoder(os.Stdout)
	enc.EncodeRecord(in)
	enc.Flush()

	// Output:
	// name,text
	// Ed,Knock knock.
}

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

func ExampleDecoder_DecodeRecord() {
	in := `name,text
Ed,Knock knock.
Sam,Who's there?
Ed,Go fmt.
Sam,Go fmt who?
Ed,Go fmt yourself!
`
	out := struct {
		Name string `csv:"name"`
		Text string `csv:"text"`
	}{}

	buf := bytes.NewBufferString(in)
	dec := headercsv.NewDecoder(buf)
	dec.DecodeRecord(&out)

	fmt.Printf("%3s: %s\n", out.Name, out.Text)
	// Output:
	//  Ed: Knock knock.
}

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
