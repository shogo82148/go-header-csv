package headercsv

import (
	"bytes"
	"io"
	"reflect"
	"testing"
)

type AString struct {
	A string
}

type ABool struct {
	A bool
}

type AInt struct {
	A int
}

type AFloat struct {
	A float64
}

type ATag struct {
	A string `csv:"a"`
}

type AInterface struct {
	A interface{} `csv:"a"`
}

func TestDecode(t *testing.T) {
	testcases := []struct {
		in  string
		ptr interface{}
		out interface{}
	}{
		// struct
		{
			"A\nb\n",
			new(AString),
			&AString{A: "b"},
		},
		{
			"A\ntrue\n",
			new(ABool),
			&ABool{A: true},
		},
		{
			"A\n123\n",
			new(AInt),
			&AInt{A: 123},
		},
		{
			"A\n123\n",
			new(AFloat),
			&AFloat{A: 123},
		},
		{
			"a\nb\n",
			new(ATag),
			&ATag{A: "b"},
		},
		{
			"a\nhoge\n",
			new(AInterface),
			&AInterface{A: "hoge"},
		},

		// slice of struct
		{
			"A\nhoge\nfuga",
			new([]AString),
			&[]AString{{"hoge"}, {"fuga"}},
		},

		// array of struct
		{
			"A\nhoge\nfuga\n",
			new([3]AString),
			&[3]AString{{"hoge"}, {"fuga"}, {""}},
		},
		{
			"A\nhoge\nfuga\n",
			new([3]*AString),
			&[3]*AString{{"hoge"}, {"fuga"}, nil},
		},

		// map
		{
			"a\nb\n",
			map[string]string{},
			map[string]string{"a": "b"},
		},
		{
			"a\n123\n",
			map[string]int{},
			map[string]int{"a": 123},
		},
		{
			"a\nb\n",
			map[string]interface{}{},
			map[string]interface{}{"a": "b"},
		},
	}

	for _, tc := range testcases {
		d := NewDecoder(bytes.NewBufferString(tc.in))
		if err := d.Decode(tc.ptr); err != nil && err != io.EOF {
			t.Errorf("unexpected error: %v(%v)", err, tc)
		}
		if !reflect.DeepEqual(tc.ptr, tc.out) {
			t.Errorf("%#v: got %v, want %v", tc.in, tc.ptr, tc.out)
		}
	}
}
