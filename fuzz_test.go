package headercsv

import (
	"strings"
	"testing"
)

func FuzzDecodeAll_map(f *testing.F) {
	f.Add("A\nb\n")
	f.Add("A,B\nb,c\n")

	f.Fuzz(func(t *testing.T, a string) {
		dec := NewDecoder(strings.NewReader(a))
		var v []map[string]any
		if err := dec.DecodeAll(&v); err != nil {
			t.Skip(err)
		}
	})
}

func FuzzDecodeAll_slice(f *testing.F) {
	f.Add("A\nb\n")
	f.Add("A,B\nb,c\n")

	f.Fuzz(func(t *testing.T, a string) {
		dec := NewDecoder(strings.NewReader(a))
		var v [][]any
		if err := dec.DecodeAll(&v); err != nil {
			t.Skip(err)
		}
	})
}

func FuzzDecodeAll_struct(f *testing.F) {
	f.Add("String,Int,Uint,Float64,Bool,Any\nHello,42,42,3.14,true,World\n")

	f.Fuzz(func(t *testing.T, a string) {
		dec := NewDecoder(strings.NewReader(a))
		var v []struct {
			String  string
			Int     int
			Uint    uint
			Float64 float64
			Bool    bool
			Any     any
		}
		if err := dec.DecodeAll(&v); err != nil {
			t.Skip(err)
		}
	})
}
