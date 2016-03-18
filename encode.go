package headercsv

import (
	"io"
	"reflect"
	"sync"
)

type Encoder struct {
	header []string
	w      io.Writer
}

func NewEncoder(w io.Writer) *Encoder {
	return &Encoder{w: w}
}

func (enc *Encoder) Encode(v interface{}) error {
	return nil
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
