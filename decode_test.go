package headercsv

import (
	"bytes"
	"errors"
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
	A any `csv:"a"`
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
		ptr any
		out any
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
			map[string]any{},
			map[string]any{"a": "b"},
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
			map[string]AString{"a": {A: "hoge"}},
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

func TestDecodeRecord(t *testing.T) {
	testcases := []struct {
		in  string
		ptr any
		out any
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
			"A\n0x123\n",
			new(AInt),
			&AInt{A: 0x123},
		},
		{
			"A\n0123\n",
			new(AInt),
			&AInt{A: 0123},
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

		// pointer
		{
			"A,B\na,b\n",
			new(APtr),
			&APtr{ptrstr("a"), "b"},
		},

		// map
		{
			"a\nb\n",
			new(map[string]string),
			&map[string]string{"a": "b"},
		},
		{
			"a\n123\n",
			new(map[string]int),
			&map[string]int{"a": 123},
		},
		{
			"a\nb\n",
			new(map[string]any),
			&map[string]any{"a": "b"},
		},

		// nested struct
		{
			"a\n" + `"{""a"":""hoge""}"` + "\n",
			new(AStruct),
			&AStruct{A: AString{A: "hoge"}},
		},
		{
			"a\n" + `"{""a"":""hoge""}"` + "\n",
			new(map[string]AString),
			&map[string]AString{"a": {A: "hoge"}},
		},
		{
			"a\n" + `"{""a"":""hoge""}"` + "\n",
			new(AMap),
			&AMap{A: map[string]string{"a": "hoge"}},
		},

		// TextUnmarshaler
		{
			"a\nA\n",
			new(map[string]testUnmarshal),
			&map[string]testUnmarshal{"a": 10},
		},
	}

	for _, tc := range testcases {
		d := NewDecoder(bytes.NewBufferString(tc.in))
		if err := d.DecodeRecord(tc.ptr); err != nil {
			t.Errorf("%#v, %T: unexpected error: %v", tc.in, tc.out, err)
		}
		if !reflect.DeepEqual(tc.ptr, tc.out) {
			t.Errorf("%#v, %T: got %#v, want %#v", tc.in, tc.out, tc.ptr, tc.out)
		}
	}
}

func TestDecodeRecord_Error(t *testing.T) {
	t.Run("not a pointer", func(t *testing.T) {
		d := NewDecoder(bytes.NewBufferString("a\nb\n"))
		err := d.DecodeRecord(123)
		if err == nil {
			t.Error("want err, but none")
		}
	})

	t.Run("unsupported key type", func(t *testing.T) {
		d := NewDecoder(bytes.NewBufferString("a\nb\n"))
		err := d.DecodeRecord(&map[int]any{})
		if err == nil {
			t.Error("want err, but none")
		}
	})

	t.Run("overflow int8 in maps", func(t *testing.T) {
		d := NewDecoder(bytes.NewBufferString("a\n128\n"))
		err := d.DecodeRecord(&map[string]int8{})
		if err == nil {
			t.Error("want err, but none")
		}
		var decodeErr *DecodeError
		if !errors.As(err, &decodeErr) {
			t.Fatal("want DecodeError, but none")
		}
		if decodeErr.StartLine != 2 {
			t.Errorf("got %d, want 2", decodeErr.StartLine)
		}
		if decodeErr.Line != 2 {
			t.Errorf("got %d, want 2", decodeErr.Line)
		}
		if decodeErr.Column != 1 {
			t.Errorf("got %d, want 1", decodeErr.Column)
		}
		if decodeErr.Err == nil {
			t.Error("want err, but none")
		}
	})

	t.Run("overflow int8 in slice", func(t *testing.T) {
		d := NewDecoder(bytes.NewBufferString("a\n128\n"))
		err := d.DecodeRecord(&[]int8{})
		if err == nil {
			t.Error("want err, but none")
		}
		var decodeErr *DecodeError
		if !errors.As(err, &decodeErr) {
			t.Fatal("want DecodeError, but none")
		}
		if decodeErr.StartLine != 2 {
			t.Errorf("got %d, want 2", decodeErr.StartLine)
		}
		if decodeErr.Line != 2 {
			t.Errorf("got %d, want 2", decodeErr.Line)
		}
		if decodeErr.Column != 1 {
			t.Errorf("got %d, want 1", decodeErr.Column)
		}
		if decodeErr.Err == nil {
			t.Error("want err, but none")
		}
	})

	t.Run("overflow int8 in array", func(t *testing.T) {
		d := NewDecoder(bytes.NewBufferString("a\n128\n"))
		err := d.DecodeRecord(&[1]int8{})
		if err == nil {
			t.Error("want err, but none")
		}
		var decodeErr *DecodeError
		if !errors.As(err, &decodeErr) {
			t.Fatal("want DecodeError, but none")
		}
		if decodeErr.StartLine != 2 {
			t.Errorf("got %d, want 2", decodeErr.StartLine)
		}
		if decodeErr.Line != 2 {
			t.Errorf("got %d, want 2", decodeErr.Line)
		}
		if decodeErr.Column != 1 {
			t.Errorf("got %d, want 1", decodeErr.Column)
		}
		if decodeErr.Err == nil {
			t.Error("want err, but none")
		}
	})

	t.Run("overflow int8 in struct", func(t *testing.T) {
		d := NewDecoder(bytes.NewBufferString("a\n128\n"))
		err := d.DecodeRecord(&struct {
			A int8 `csv:"a"`
		}{})
		if err == nil {
			t.Error("want err, but none")
		}
		var decodeErr *DecodeError
		if !errors.As(err, &decodeErr) {
			t.Fatal("want DecodeError, but none")
		}
		if decodeErr.StartLine != 2 {
			t.Errorf("got %d, want 2", decodeErr.StartLine)
		}
		if decodeErr.Line != 2 {
			t.Errorf("got %d, want 2", decodeErr.Line)
		}
		if decodeErr.Column != 1 {
			t.Errorf("got %d, want 1", decodeErr.Column)
		}
		if decodeErr.Err == nil {
			t.Error("want err, but none")
		}
	})

	t.Run("overflow uint8 in maps", func(t *testing.T) {
		d := NewDecoder(bytes.NewBufferString("a\n256\n"))
		err := d.DecodeRecord(&map[string]uint8{})
		if err == nil {
			t.Error("want err, but none")
		}
		var decodeErr *DecodeError
		if !errors.As(err, &decodeErr) {
			t.Fatal("want DecodeError, but none")
		}
		if decodeErr.StartLine != 2 {
			t.Errorf("got %d, want 2", decodeErr.StartLine)
		}
		if decodeErr.Line != 2 {
			t.Errorf("got %d, want 2", decodeErr.Line)
		}
		if decodeErr.Column != 1 {
			t.Errorf("got %d, want 1", decodeErr.Column)
		}
		if decodeErr.Err == nil {
			t.Error("want err, but none")
		}
	})

	t.Run("parse error int8 in maps", func(t *testing.T) {
		d := NewDecoder(bytes.NewBufferString("a\nabc\n"))
		err := d.DecodeRecord(&map[string]int8{})
		if err == nil {
			t.Error("want err, but none")
		}
		var decodeErr *DecodeError
		if !errors.As(err, &decodeErr) {
			t.Fatal("want DecodeError, but none")
		}
		if decodeErr.StartLine != 2 {
			t.Errorf("got %d, want 2", decodeErr.StartLine)
		}
		if decodeErr.Line != 2 {
			t.Errorf("got %d, want 2", decodeErr.Line)
		}
		if decodeErr.Column != 1 {
			t.Errorf("got %d, want 1", decodeErr.Column)
		}
		if decodeErr.Err == nil {
			t.Error("want err, but none")
		}
	})

	t.Run("parse error uint8 in maps", func(t *testing.T) {
		d := NewDecoder(bytes.NewBufferString("a\nabc\n"))
		err := d.DecodeRecord(&map[string]uint8{})
		if err == nil {
			t.Error("want err, but none")
		}
		var decodeErr *DecodeError
		if !errors.As(err, &decodeErr) {
			t.Fatal("want DecodeError, but none")
		}
		if decodeErr.StartLine != 2 {
			t.Errorf("got %d, want 2", decodeErr.StartLine)
		}
		if decodeErr.Line != 2 {
			t.Errorf("got %d, want 2", decodeErr.Line)
		}
		if decodeErr.Column != 1 {
			t.Errorf("got %d, want 1", decodeErr.Column)
		}
		if decodeErr.Err == nil {
			t.Error("want err, but none")
		}
	})

	t.Run("overflow float32 in maps", func(t *testing.T) {
		d := NewDecoder(bytes.NewBufferString("a\n1e100\n"))
		err := d.DecodeRecord(&map[string]float32{})
		if err == nil {
			t.Error("want err, but none")
		}
		var decodeErr *DecodeError
		if !errors.As(err, &decodeErr) {
			t.Fatal("want DecodeError, but none")
		}
		if decodeErr.StartLine != 2 {
			t.Errorf("got %d, want 2", decodeErr.StartLine)
		}
		if decodeErr.Line != 2 {
			t.Errorf("got %d, want 2", decodeErr.Line)
		}
		if decodeErr.Column != 1 {
			t.Errorf("got %d, want 1", decodeErr.Column)
		}
		if decodeErr.Err == nil {
			t.Error("want err, but none")
		}
	})

	t.Run("parse error float32 in maps", func(t *testing.T) {
		d := NewDecoder(bytes.NewBufferString("a\nabc\n"))
		err := d.DecodeRecord(&map[string]float32{})
		if err == nil {
			t.Error("want err, but none")
		}
		var decodeErr *DecodeError
		if !errors.As(err, &decodeErr) {
			t.Fatal("want DecodeError, but none")
		}
		if decodeErr.StartLine != 2 {
			t.Errorf("got %d, want 2", decodeErr.StartLine)
		}
		if decodeErr.Line != 2 {
			t.Errorf("got %d, want 2", decodeErr.Line)
		}
		if decodeErr.Column != 1 {
			t.Errorf("got %d, want 1", decodeErr.Column)
		}
		if decodeErr.Err == nil {
			t.Error("want err, but none")
		}
	})
}

func TestDecodeAll(t *testing.T) {
	testcases := []struct {
		in  string
		ptr any
		out any
	}{
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
			new([]map[string]string),
			&[]map[string]string{{"a": "b"}},
		},
		{
			"a\n123\n",
			new([]map[string]int),
			&[]map[string]int{{"a": 123}},
		},
		{
			"a\nb\n",
			new([]map[string]any),
			&[]map[string]any{{"a": "b"}},
		},

		// nested struct
		{
			"a\n" + `"{""a"":""hoge""}"` + "\n",
			new([]*AStruct),
			&[]*AStruct{{A: AString{A: "hoge"}}},
		},
		{
			"a\n" + `"{""a"":""hoge""}"` + "\n",
			new([]map[string]AString),
			&[]map[string]AString{{"a": {A: "hoge"}}},
		},
		{
			"a\n" + `"{""a"":""hoge""}"` + "\n",
			new([]*AMap),
			&[]*AMap{{A: map[string]string{"a": "hoge"}}},
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
			new([]map[string]testUnmarshal),
			&[]map[string]testUnmarshal{{"a": 10}},
		},
	}

	for _, tc := range testcases {
		d := NewDecoder(bytes.NewBufferString(tc.in))
		if err := d.DecodeAll(tc.ptr); err != nil {
			t.Errorf("%#v, %T: unexpected error: %v", tc.in, tc.out, err)
		}
		if !reflect.DeepEqual(tc.ptr, tc.out) {
			t.Errorf("%#v, %T: got %#v, want %#v", tc.in, tc.out, tc.ptr, tc.out)
		}
	}
}
