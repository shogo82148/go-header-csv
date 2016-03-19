package headercsv

import (
	"encoding/csv"
	"encoding/json"
	"errors"
	"io"
	"reflect"
	"strconv"
)

type Decoder struct {
	header []string
	r      *csv.Reader
}

func NewDecoder(r io.Reader) *Decoder {
	return &Decoder{r: csv.NewReader(r)}
}

func (dec *Decoder) Decode(v interface{}) error {
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

func (dec *Decoder) SetHeader(header []string) error {
	if dec.header != nil {
		return errors.New("header has been alread set")
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
			return errors.New("unsupported type")
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

func (dec *Decoder) decodeField(v reflect.Value, field string) error {
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
			json.Unmarshal([]byte(field), v.Addr().Interface())
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
