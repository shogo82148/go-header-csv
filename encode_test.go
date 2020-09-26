package headercsv

import (
	"bytes"
	"reflect"
	"sort"
	"testing"
)

type testMarshal int

func (i testMarshal) MarshalText() ([]byte, error) {
	return []byte("<testMarshal>"), nil
}

func TestEncode(t *testing.T) {
	type FooBar struct {
		Foo string `json:"foo"`
		Bar string `json:"bar"`
	}

	testcases := []struct {
		in  interface{}
		out string
	}{
		{
			map[string]string{"a": "b"},
			"a\nb\n",
		},
		{
			struct{ A string }{"b"},
			"A\nb\n",
		},
		{
			struct {
				A string `csv:"a"`
			}{"b"},
			"a\nb\n",
		},
		{
			struct {
				A string `csv:"-"`
				B string `csv:"b"`
			}{"a", "b"},
			"b\nb\n",
		},
		{
			struct {
				A int `csv:"a"`
			}{10},
			"a\n10\n",
		},
		{
			struct {
				A uint `csv:"a"`
			}{10},
			"a\n10\n",
		},
		{
			struct {
				A float64 `csv:"a"`
			}{10},
			"a\n10\n",
		},
		{
			[]struct {
				A int `csv:"a,omitempty"`
				B int `csv:"b"`
			}{{1, 1}, {1, 0}, {0, 1}, {0, 0}},
			"a,b\n1,1\n1,0\n,1\n,0\n",
		},
		{
			struct {
				A FooBar
			}{A: FooBar{"f", "b"}},
			"A\n" + `"{""foo"":""f"",""bar"":""b""}"` + "\n",
		},
		{
			struct {
				A testMarshal
			}{A: 0},
			"A\n<testMarshal>\n",
		},
		{
			struct {
				A *string `csv:"a"`
			}{new(string)},
			"a\n\n",
		},
		{
			&struct {
				A string `csv:"a"`
			}{"b"},
			"a\nb\n",
		},

		// struct and (slice or array)
		{
			[]interface{}{
				struct {
					A string `csv:"a"`
					B string `csv:"b"`
					C string `csv:"c"`
				}{"this", "is", "struct"},
				[]string{"this", "is", "slice"},
			},
			"a,b,c\nthis,is,struct\nthis,is,slice\n",
		},
		{
			[]interface{}{
				struct {
					A string `csv:"a"`
					B string `csv:"b"`
					C string `csv:"c"`
				}{"this", "is", "struct"},
				&[]string{"this", "is", "slice"},
			},
			"a,b,c\nthis,is,struct\nthis,is,slice\n",
		},
		{
			[]interface{}{
				struct {
					A string `csv:"a"`
					B string `csv:"b"`
					C string `csv:"c"`
				}{"this", "is", "struct"},
				[3]string{"this", "is", "slice"},
			},
			"a,b,c\nthis,is,struct\nthis,is,slice\n",
		},
		{
			[]interface{}{
				struct {
					A string `csv:"a"`
					B string `csv:"b"`
					C string `csv:"c"`
				}{"this", "is", "struct"},
				&[3]string{"this", "is", "slice"},
			},
			"a,b,c\nthis,is,struct\nthis,is,slice\n",
		},
	}

	for _, tc := range testcases {
		buf := &bytes.Buffer{}
		enc := NewEncoder(buf)
		if err := enc.Encode(tc.in); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		enc.Flush()
		if buf.String() != tc.out {
			t.Errorf("got %v, expected %v", buf.String(), tc.out)
		}
	}
}

func TestNamedRecordType(t *testing.T) {
	testcases := []struct {
		in      interface{}
		out     map[string]string
		headers []string
		ordered bool
	}{
		{
			in:      map[string]string{"a": "b"},
			out:     map[string]string{"a": "b"},
			headers: []string{"a"},
			ordered: false,
		},
		{
			in:      struct{ A string }{"b"},
			out:     map[string]string{"A": "b"},
			headers: []string{"A"},
			ordered: false,
		},
		{
			in: struct {
				A string `csv:"a"`
			}{"b"},
			out:     map[string]string{"a": "b"},
			headers: []string{"a"},
			ordered: false,
		},
	}

	for _, tc := range testcases {
		in := reflect.ValueOf(tc.in)
		rt := recordType(reflect.TypeOf(tc.in))

		// Test FieldByName
		for k, v := range tc.out {
			got, _ := rt.Field(in, 0, k)
			if got.String() != v {
				t.Errorf("incorrect value: got %v, expected %v", got.String(), v)
			}
		}

		// Test HeaderNames
		headers := rt.HeaderNames(in)
		if !tc.ordered {
			sort.Sort(sort.StringSlice(headers))
			sort.Sort(sort.StringSlice(tc.headers))
		}
		if !reflect.DeepEqual(headers, tc.headers) {
			t.Errorf("incorrect header: got %s, expected %v", headers, tc.headers)
		}
	}
}
