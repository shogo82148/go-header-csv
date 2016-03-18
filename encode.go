package headercsv

import (
	"encoding/csv"
	"errors"
	"io"
	"reflect"
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
			record[i] = rt.FieldByName(v, k).String()
		}
	default:
		return errors.New("unsupported type")
	}
	return enc.w.Write(record)
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
	FieldByIndex(v reflect.Value, i int) reflect.Value
}

type namedRecordType interface {
	FieldByName(v reflect.Value, name string) reflect.Value
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
	}
	return nil
}

var unsupportedRecordType struct{}

type mapRecordType struct {
}

func (rt *mapRecordType) FieldByName(v reflect.Value, name string) reflect.Value {
	return v.MapIndex(reflect.ValueOf(name))
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

type structRecordType struct {
	headers []string
	index   map[string]int
}

func (rt *structRecordType) FieldByName(v reflect.Value, name string) reflect.Value {
	return v.Field(rt.index[name])
}

func (rt *structRecordType) HeaderNames(v reflect.Value) []string {
	return rt.headers
}

func newStructRecordType(t reflect.Type) interface{} {
	num := t.NumField()
	headers := make([]string, 0, num)
	index := make(map[string]int, num)
	for i := 0; i < num; i++ {
		name := t.Field(i).Tag.Get("csv")
		if name == "-" {
			continue
		}
		if name == "" {
			name = t.Field(i).Name
		}
		headers = append(headers, name)
		index[name] = i
	}
	return &structRecordType{
		headers: headers,
		index:   index,
	}
}
