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

// Encoder writes CSV records to an output stream.
type Encoder struct {
	MarshalField func(v any) ([]byte, error)

	header []string
	w      *csv.Writer
}

// NewEncoder returns a new encoder that writes to w.
func NewEncoder(w io.Writer) *Encoder {
	return &Encoder{w: csv.NewWriter(w)}
}

// NewEncoder returns a new encoder that writes to w.
func NewEncoderCSV(w *csv.Writer) *Encoder {
	return &Encoder{w: w}
}

// Encode writes a CSV record to the stream.
//
// Deprecated: use EncodeRecord or EncodeAll instead.
func (enc *Encoder) Encode(v any) error {
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

// EncodeRecord writes a CSV record to the stream.
func (enc *Encoder) EncodeRecord(v any) error {
	if enc.MarshalField == nil {
		enc.MarshalField = json.Marshal
	}

	rv := reflect.ValueOf(v)
	return enc.encodeRecord(rv)
}

// EncodeAll writes all CSV records to the stream.
// v must be a slice or an array.
func (enc *Encoder) EncodeAll(v any) error {
	if enc.MarshalField == nil {
		enc.MarshalField = json.Marshal
	}

	rv := reflect.ValueOf(v)
	kind := rv.Kind()
	if kind != reflect.Array && kind != reflect.Slice {
		return errors.New("headercsv: v is neither a slice nor an array")
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

// steel from https://github.com/golang/go/blob/1763ee199d33d2592332a29cfc3da7811718a4fd/src/encoding/json/encode.go#L318-L330
func isEmptyValue(v reflect.Value) bool {
	switch v.Kind() {
	case reflect.Array, reflect.Map, reflect.Slice, reflect.String:
		return v.Len() == 0
	case reflect.Bool,
		reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr,
		reflect.Float32, reflect.Float64,
		reflect.Interface, reflect.Pointer:
		return v.IsZero()
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

// SetHeader sets the header.
func (enc *Encoder) SetHeader(header []string) error {
	if enc.header != nil {
		return errors.New("headercsv: the header has been already set")
	}
	enc.header = header
	return enc.w.Write(header)
}

// Flush writes any buffered data to the underlying io.Writer.
// To check if an error occurred during the Flush, call Error.
func (enc *Encoder) Flush() {
	enc.w.Flush()
}

// Error reports any error that has occurred during a previous Write or Flush.
func (enc *Encoder) Error() error {
	return enc.w.Error()
}

type recordInterface interface {
	Field(v reflect.Value, i int, name string) (reflect.Value, *field)
	HeaderNames(v reflect.Value) []string
}

var recordTypeCache sync.Map

func recordType(t reflect.Type) recordInterface {
	f, ok := recordTypeCache.Load(t)
	if ok {
		return f.(recordInterface)
	}

	newType := newRecordType(t)
	recordTypeCache.Store(t, newType)
	return newType
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
