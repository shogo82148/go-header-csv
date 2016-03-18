package headercsv

import (
	"reflect"
	"sort"
	"testing"
)

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
