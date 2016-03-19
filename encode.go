package headercsv

import (
	"encoding"
	"encoding/csv"
	"encoding/json"
	"errors"
	"io"
	"reflect"
	"strconv"
	"sync"
)

type Encoder struct {
	header []string
	w      *csv.Writer
}

func NewEncoder(w io.Writer) *Encoder {
	return &Encoder{w: csv.NewWriter(w)}
}

func (enc *Encoder) Encode(v interface{}) error {
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
		named := rt.(namedRecordType)
		if named == nil {
			return errors.New("cannot decide header")
		}
		if err := enc.SetHeader(named.HeaderNames(v)); err != nil {
			return err
		}
	}

	// fill record
	record := make([]string, len(enc.header))
	switch rt := rt.(type) {
	case namedRecordType:
		for i, k := range enc.header {
			v, opt := rt.FieldByName(v, k)
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
	case orderedRecoredType:
		for i := range enc.header {
			v, _ := rt.FieldByIndex(v, i)
			s, err := enc.encodeField(v, nil)
			if err != nil {
				return err
			}
			record[i] = s
		}
	default:
		return errors.New("unsupported type")
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
		j, err := json.Marshal(v.Interface())
		if err != nil {
			return "", err
		}
		return string(j), nil
	}
	return "", errors.New("unsuported type")
}

func (enc *Encoder) SetHeader(header []string) error {
	if enc.header != nil {
		return errors.New("header has been alread set")
	}
	enc.header = header
	return enc.w.Write(header)
}

func (enc *Encoder) Flush() {
	enc.w.Flush()
}

type orderedRecoredType interface {
	FieldByIndex(v reflect.Value, i int) (reflect.Value, *field)
}

type namedRecordType interface {
	FieldByName(v reflect.Value, name string) (reflect.Value, *field)
	HeaderNames(v reflect.Value) []string
}

var recordTypeCache struct {
	sync.RWMutex
	m map[reflect.Type]interface{}
}

func recordType(t reflect.Type) interface{} {
	recordTypeCache.RLock()
	f := recordTypeCache.m[t]
	recordTypeCache.RUnlock()
	if f != nil {
		return f
	}

	f = newRecordType(t)
	recordTypeCache.Lock()
	if recordTypeCache.m == nil {
		recordTypeCache.m = map[reflect.Type]interface{}{}
	}
	recordTypeCache.m[t] = f
	recordTypeCache.Unlock()
	return f
}

func newRecordType(t reflect.Type) interface{} {
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

var unsupportedRecordType struct{}

type mapRecordType struct {
}

func (rt *mapRecordType) FieldByName(v reflect.Value, name string) (reflect.Value, *field) {
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

func newMapRecordType(t reflect.Type) interface{} {
	if t.Key().Kind() != reflect.String {
		return unsupportedRecordType
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

func (rt *structRecordType) FieldByName(v reflect.Value, name string) (reflect.Value, *field) {
	f := rt.fields[name]
	return v.Field(f.index), f
}

func (rt *structRecordType) HeaderNames(v reflect.Value) []string {
	return rt.headers
}

func newStructRecordType(t reflect.Type) interface{} {
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

type namedPtrRecordType struct {
	elem namedRecordType
}

func (rt *namedPtrRecordType) FieldByName(v reflect.Value, name string) (reflect.Value, *field) {
	return rt.elem.FieldByName(v.Elem(), name)
}

func (rt *namedPtrRecordType) HeaderNames(v reflect.Value) []string {
	return rt.elem.HeaderNames(v.Elem())
}

type orderedPtrRecordType struct {
	elem orderedRecoredType
}

func (rt *orderedPtrRecordType) FieldByIndex(v reflect.Value, i int) (reflect.Value, *field) {
	return rt.elem.FieldByIndex(v.Elem(), i)
}

func newPtrRecordType(t reflect.Type) interface{} {
	elem := recordType(t.Elem())
	switch elem := elem.(type) {
	case namedRecordType:
		return &namedPtrRecordType{
			elem: elem,
		}
	case orderedRecoredType:
		return &orderedPtrRecordType{
			elem: elem,
		}
	}
	return unsupportedRecordType
}

type sliceRecordType struct {
	zero reflect.Value
}

func (rt *sliceRecordType) FieldByIndex(v reflect.Value, i int) (reflect.Value, *field) {
	if i >= v.Len() {
		return rt.zero, nil
	}
	return v.Index(i), nil
}

func newSliceRecordType(t reflect.Type) interface{} {
	return &sliceRecordType{
		zero: reflect.Zero(t.Elem()),
	}
}
