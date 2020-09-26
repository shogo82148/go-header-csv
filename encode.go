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
	"sync"
)

type Encoder struct {
	MarshalField func(v interface{}) ([]byte, error)

	header []string
	w      *csv.Writer
}

func NewEncoder(w io.Writer) *Encoder {
	return &Encoder{w: csv.NewWriter(w)}
}

func (enc *Encoder) Encode(v interface{}) error {
	if enc.MarshalField == nil {
		enc.MarshalField = json.Marshal
	}

	rv := reflect.ValueOf(v)
	if rv.Kind() != reflect.Array && rv.Kind() != reflect.Slice {
		return enc.encodeRecord(rv)
	}

	for i := 0; i < rv.Len(); i++ {
		err := enc.encodeRecord(rv.Index(i))
		if err != nil {
			return err
		}
	}
	return nil
}

func (enc *Encoder) encodeRecord(v reflect.Value) error {
	if v.Kind() == reflect.Interface {
		v = v.Elem()
	}
	rt := recordType(v.Type())
	if enc.header == nil {
		// guess header
		header := rt.HeaderNames(v)
		if header == nil {
			return errors.New("cannot decide header")
		}
		if err := enc.SetHeader(header); err != nil {
			return err
		}
	}

	// fill record
	record := make([]string, len(enc.header))
	for i, k := range enc.header {
		v, opt := rt.Field(v, i, k)
		if opt != nil && opt.omitEmpty && isEmptyValue(v) {
			record[i] = ""
			continue
		}
		s, err := enc.encodeField(v, opt)
		if err != nil {
			return err
		}
		record[i] = s
	}
	return enc.w.Write(record)
}

// steel from https://golang.org/src/encoding/json/encode.go
func isEmptyValue(v reflect.Value) bool {
	switch v.Kind() {
	case reflect.Array, reflect.Map, reflect.Slice, reflect.String:
		return v.Len() == 0
	case reflect.Bool:
		return !v.Bool()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return v.Int() == 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return v.Uint() == 0
	case reflect.Float32, reflect.Float64:
		return v.Float() == 0
	case reflect.Interface, reflect.Ptr:
		return v.IsNil()
	}
	return false
}

func (enc *Encoder) encodeField(v reflect.Value, opt *field) (string, error) {
	if m, ok := v.Interface().(encoding.TextMarshaler); ok {
		text, err := m.MarshalText()
		if err != nil {
			return "", err
		}
		return string(text), nil
	}

	switch v.Kind() {
	case reflect.String:
		return v.String(), nil
	case reflect.Bool:
		return strconv.FormatBool(v.Bool()), nil
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return strconv.FormatInt(v.Int(), 10), nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return strconv.FormatUint(v.Uint(), 10), nil
	case reflect.Float32:
		return strconv.FormatFloat(v.Float(), 'g', -1, 32), nil
	case reflect.Float64:
		return strconv.FormatFloat(v.Float(), 'g', -1, 64), nil
	case reflect.Ptr:
		if v.IsNil() {
			return "null", nil
		}
		return enc.encodeField(v.Elem(), opt)
	case reflect.Array, reflect.Map, reflect.Slice, reflect.Interface, reflect.Struct:
		j, err := enc.MarshalField(v.Interface())
		if err != nil {
			return "", err
		}
		return string(j), nil
	}

	return "", fmt.Errorf("headercsv: unsupported type: %s", v.Type().String())
}

func (enc *Encoder) SetHeader(header []string) error {
	if enc.header != nil {
		return errors.New("headercsv: the header has been already set")
	}
	enc.header = header
	return enc.w.Write(header)
}

func (enc *Encoder) Flush() {
	enc.w.Flush()
}

type recordInterface interface {
	Field(v reflect.Value, i int, name string) (reflect.Value, *field)
	HeaderNames(v reflect.Value) []string
}

var recordTypeCache struct {
	sync.RWMutex
	m map[reflect.Type]recordInterface
}

func recordType(t reflect.Type) recordInterface {
	recordTypeCache.RLock()
	f := recordTypeCache.m[t]
	recordTypeCache.RUnlock()
	if f != nil {
		return f
	}

	f = newRecordType(t)
	recordTypeCache.Lock()
	if recordTypeCache.m == nil {
		recordTypeCache.m = map[reflect.Type]recordInterface{}
	}
	recordTypeCache.m[t] = f
	recordTypeCache.Unlock()
	return f
}

func newRecordType(t reflect.Type) recordInterface {
	switch t.Kind() {
	case reflect.Map:
		return newMapRecordType(t)
	case reflect.Struct:
		return newStructRecordType(t)
	case reflect.Ptr:
		return newPtrRecordType(t)
	case reflect.Slice, reflect.Array:
		return newSliceRecordType(t)
	}
	return nil
}

type unsupportedRecordType struct{}

func (rt *unsupportedRecordType) Field(v reflect.Value, i int, name string) (reflect.Value, *field) {
	return reflect.Zero(v.Type()), nil
}

func (rt *unsupportedRecordType) HeaderNames(v reflect.Value) []string {
	return nil
}

type mapRecordType struct {
}

func (rt *mapRecordType) Field(v reflect.Value, i int, name string) (reflect.Value, *field) {
	return v.MapIndex(reflect.ValueOf(name)), nil
}

func (rt *mapRecordType) HeaderNames(v reflect.Value) []string {
	vkeys := v.MapKeys()
	keys := make([]string, len(vkeys))
	for i, k := range vkeys {
		keys[i] = k.String()
	}
	return keys
}

func newMapRecordType(t reflect.Type) recordInterface {
	if t.Key().Kind() != reflect.String {
		return &unsupportedRecordType{}
	}
	return &mapRecordType{}
}

type field struct {
	index     int
	omitEmpty bool
}

type structRecordType struct {
	headers []string
	fields  map[string]*field
}

func (rt *structRecordType) Field(v reflect.Value, i int, name string) (reflect.Value, *field) {
	f, ok := rt.fields[name]
	if !ok {
		return reflect.Value{}, f
	}
	return v.Field(f.index), f
}

func (rt *structRecordType) HeaderNames(v reflect.Value) []string {
	return rt.headers
}

func newStructRecordType(t reflect.Type) recordInterface {
	num := t.NumField()
	headers := make([]string, 0, num)
	fields := make(map[string]*field, num)
	for i := 0; i < num; i++ {
		tag := t.Field(i).Tag.Get("csv")
		if tag == "-" {
			continue
		}
		name, opts := parseTag(tag)
		if name == "" {
			name = t.Field(i).Name
		}
		headers = append(headers, name)
		fields[name] = &field{
			index:     i,
			omitEmpty: opts.Contains("omitempty"),
		}
	}
	return &structRecordType{
		headers: headers,
		fields:  fields,
	}
}

type ptrRecordType struct {
	elem recordInterface
}

func (rt *ptrRecordType) Field(v reflect.Value, i int, name string) (reflect.Value, *field) {
	return rt.elem.Field(v.Elem(), i, name)
}

func (rt *ptrRecordType) HeaderNames(v reflect.Value) []string {
	return rt.elem.HeaderNames(v.Elem())
}

func newPtrRecordType(t reflect.Type) recordInterface {
	elem := recordType(t.Elem())
	return &ptrRecordType{
		elem: elem,
	}
}

type sliceRecordType struct {
}

func (rt *sliceRecordType) Field(v reflect.Value, i int, name string) (reflect.Value, *field) {
	if i >= v.Len() {
		return reflect.Zero(v.Type().Elem()), nil
	}
	return v.Index(i), nil
}

func (rt *sliceRecordType) HeaderNames(v reflect.Value) []string {
	return nil
}

func newSliceRecordType(t reflect.Type) recordInterface {
	return &sliceRecordType{}
}
