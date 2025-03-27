package headercsv_test

import (
	"bytes"
	"fmt"
	"io"
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
	if err := enc.Encode(in); err != nil {
		panic(err)
	}
	enc.Flush()
	if err := enc.Error(); err != nil {
		panic(err)
	}

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
	if err := enc.EncodeRecord(in); err != nil {
		panic(err)
	}
	enc.Flush()
	if err := enc.Error(); err != nil {
		panic(err)
	}

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
	if err := enc.EncodeAll(in); err != nil {
		panic(err)
	}
	enc.Flush()
	if err := enc.Error(); err != nil {
		panic(err)
	}

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
	if err := dec.Decode(&out); err != nil && err != io.EOF {
		panic(err)
	}

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
	if err := dec.DecodeRecord(&out); err != nil {
		panic(err)
	}

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
	if err := dec.DecodeAll(&out); err != nil {
		panic(err)
	}

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
