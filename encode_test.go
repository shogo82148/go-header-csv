package headercsv

import (
	"bytes"
	"reflect"
	"sort"
	"testing"
)

func TestEncode(t *testing.T) {
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
	}

	for _, tc := range testcases {
		buf := &bytes.Buffer{}
		enc := NewEncoder(buf)
		if err := enc.Encode(tc.in); err != nil {
			t.Errorf("unexpcted error: %v", err)
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
	}

	for _, tc := range testcases {
		in := reflect.ValueOf(tc.in)
		rt := recordType(reflect.TypeOf(tc.in)).(namedRecordType)

		// Test FieldByName
		for k, v := range tc.out {
			got := rt.FieldByName(in, k).String()
			if got != v {
				t.Errorf("incorrect value: got %v, expected %v", got, v)
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
