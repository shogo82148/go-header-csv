package headercsv

import (
	"encoding"
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"reflect"
	"strconv"
)

// Decoder reads and decodes CSV values from an input stream.
type Decoder struct {
	UnmarshalField func(in []byte, out interface{}) error

	header []string
	r      *csv.Reader
}

// NewDecoder returns a new decoder that reads from r.
func NewDecoder(r io.Reader) *Decoder {
	return &Decoder{r: csv.NewReader(r)}
}

// Decode reads the next CSV record from its input and stores it in the value pointed to by v.
func (dec *Decoder) Decode(v interface{}) error {
	if dec.UnmarshalField == nil {
		dec.UnmarshalField = json.Unmarshal
	}

	if dec.header == nil {
		record, err := dec.r.Read()
		if err != nil {
			return err
		}
		dec.SetHeader(record)
	}

	rv := reflect.ValueOf(v)

	if rv.Kind() == reflect.Ptr && rv.Elem().Kind() == reflect.Slice {
		elem := reflect.MakeSlice(rv.Elem().Type(), 0, 4)
		for {
			ev := reflect.New(elem.Type().Elem())
			if err := dec.decodeRecord(ev); err != nil {
				rv.Elem().Set(elem)
				return err
			}
			elem = reflect.Append(elem, ev.Elem())
		}
	}

	if rv.Kind() == reflect.Ptr && rv.Elem().Kind() == reflect.Array {
		elem := rv.Elem()
		l := elem.Len()
		for i := 0; i < l; i++ {
			ev := elem.Index(i)
			if err := dec.decodeRecord(ev); err != nil {
				ev.Set(reflect.Zero(ev.Type()))
				return err
			}
		}
	}

	return dec.decodeRecord(rv)
}

// SetHeader sets the header.
// If no header is set, first CSV record is used for the header.
func (dec *Decoder) SetHeader(header []string) error {
	if dec.header != nil {
		return errors.New("headercsv: the header has been already set")
	}
	dec.header = header
	return nil
}

func (dec *Decoder) decodeRecord(v reflect.Value) error {
	v = dec.indirect(v)
	record, err := dec.r.Read()
	if err != nil {
		return err
	}

	t := v.Type()
	if v.Kind() == reflect.Map {
		if t.Key().Kind() != reflect.String {
			return fmt.Errorf("headercsv: unsupported type: %s", t.Key().String())
		}
		elemType := v.Type().Elem()
		for i, k := range dec.header {
			if i >= len(record) {
				break
			}
			elem := reflect.New(elemType).Elem()
			if err := dec.decodeField(elem, record[i]); err != nil {
				return err
			}
			v.SetMapIndex(reflect.ValueOf(k), elem)
		}
		return nil
	}

	if v.Kind() == reflect.Slice {
		v.Set(reflect.MakeSlice(v.Type(), len(dec.header), len(dec.header)))
		rt := recordType(t)
		for i, k := range dec.header {
			if i >= len(record) {
				break
			}
			v, _ := rt.Field(v, i, k)
			if err := dec.decodeField(v, record[i]); err != nil {
				return err
			}
		}
		return nil
	}

	if v.Kind() == reflect.Array {
		rt := recordType(t)
		for i, k := range dec.header {
			if i >= len(record) {
				break
			}
			v, _ := rt.Field(v, i, k)
			if err := dec.decodeField(v, record[i]); err != nil {
				return err
			}
		}
		return nil
	}

	rt := recordType(t)
	for i, k := range dec.header {
		if i >= len(record) {
			break
		}
		v, f := rt.Field(v, i, k)
		if f != nil {
			if err := dec.decodeField(v, record[i]); err != nil {
				return err
			}
		}
	}

	return nil
}

func (dec *Decoder) decodeField(v reflect.Value, field string) error {
	if field == "" && v.Kind() == reflect.Ptr {
		v.Set(reflect.Zero(v.Type()))
		return nil
	}
	u, v := dec.indirectField(v)
	if u != nil {
		return u.UnmarshalText([]byte(field))
	}
	if !v.CanSet() {
		return nil
	}
	switch v.Kind() {
	case reflect.Bool:
		b, err := strconv.ParseBool(field)
		if err != nil {
			return err
		}
		v.SetBool(b)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		i, err := strconv.ParseInt(field, 10, 64)
		if err != nil {
			return err
		}
		if v.OverflowInt(i) {
			return errors.New("overflow")
		}
		v.SetInt(i)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		i, err := strconv.ParseUint(field, 10, 64)
		if err != nil {
			return err
		}
		if v.OverflowUint(i) {
			return errors.New("overflow")
		}
		v.SetUint(i)
	case reflect.Float32, reflect.Float64:
		n, err := strconv.ParseFloat(field, v.Type().Bits())
		if err != nil {
			return err
		}
		if v.OverflowFloat(n) {
			return errors.New("overflow")
		}
		v.SetFloat(n)
	case reflect.String:
		v.SetString(field)
	case reflect.Interface:
		v.Set(reflect.ValueOf(field))
	default:
		if v.CanAddr() {
			return dec.UnmarshalField([]byte(field), v.Addr().Interface())
		}
	}
	return nil
}

func (dec *Decoder) indirect(v reflect.Value) reflect.Value {
	if v.Kind() != reflect.Ptr && v.Type().Name() != "" && v.CanAddr() {
		v = v.Addr()
	}
	for {
		if v.Kind() != reflect.Ptr {
			break
		}
		if v.IsNil() {
			v.Set(reflect.New(v.Type().Elem()))
		}
		v = v.Elem()
	}
	return v
}

func (dec *Decoder) indirectField(v reflect.Value) (encoding.TextUnmarshaler, reflect.Value) {
	if v.Kind() != reflect.Ptr && v.Type().Name() != "" && v.CanAddr() {
		v = v.Addr()
	}
	for {
		if v.Kind() != reflect.Ptr {
			break
		}
		if v.IsNil() {
			v.Set(reflect.New(v.Type().Elem()))
		}
		if v.Type().NumMethod() > 0 {
			if u, ok := v.Interface().(encoding.TextUnmarshaler); ok {
				return u, reflect.Value{}
			}
		}
		v = v.Elem()
	}
	return nil, v
}
