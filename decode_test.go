package headercsv

import (
	"bytes"
	"io"
	"reflect"
	"strconv"
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

type AStruct struct {
	A AString `csv:"a"`
}

type AMap struct {
	A map[string]string `csv:"a"`
}

type APtr struct {
	A *string
	B string
}

func ptrstr(s string) *string { return &s }

type testUnmarshal int

func (t *testUnmarshal) UnmarshalText(data []byte) error {
	// decode hex number
	i, err := strconv.ParseInt(string(data), 16, 0)
	if err != nil {
		return err
	}
	*t = testUnmarshal(i)
	return nil
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
		//AString doesn't have member B
		{
			"A,B\nb,c\n",
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
		{
			"A\nhoge\nfuga",
			new([]*AString),
			&[]*AString{{"hoge"}, {"fuga"}},
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

		// pointer
		{
			"A,B\na,b\n,b\na,\n",
			new([]*APtr),
			&[]*APtr{{ptrstr("a"), "b"}, {nil, "b"}, {ptrstr("a"), ""}},
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

		// nested struct
		{
			"a\n" + `"{""a"":""hoge""}"` + "\n",
			new(AStruct),
			&AStruct{A: AString{A: "hoge"}},
		},
		{
			"a\n" + `"{""a"":""hoge""}"` + "\n",
			map[string]AString{},
			map[string]AString{"a": AString{A: "hoge"}},
		},
		{
			"a\n" + `"{""a"":""hoge""}"` + "\n",
			new(AMap),
			&AMap{A: map[string]string{"a": "hoge"}},
		},

		// slice of slice
		{
			"a,b,c\n1,2,3\n4,5,6\n",
			new([][]string),
			&[][]string{{"1", "2", "3"}, {"4", "5", "6"}},
		},
		{
			"a,b,c\n1,2,3\n4,5,6\n",
			new([][3]string),
			&[][3]string{{"1", "2", "3"}, {"4", "5", "6"}},
		},

		// TextUnmarshaler
		{
			"a\nA\n",
			map[string]testUnmarshal{},
			map[string]testUnmarshal{"a": 10},
		},
	}

	for _, tc := range testcases {
		d := NewDecoder(bytes.NewBufferString(tc.in))
		if err := d.Decode(tc.ptr); err != nil && err != io.EOF {
			t.Errorf("unexpected error: %v(%v)", err, tc)
		}
		if !reflect.DeepEqual(tc.ptr, tc.out) {
			t.Errorf("%#v: got %#v, want %#v", tc.in, tc.ptr, tc.out)
		}
	}
}
