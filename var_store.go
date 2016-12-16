package sophia

import (
	"reflect"
	"unsafe"
)

// varStore manages memory allocation and free for C variables
// Interface for C sophia object
// Only for internal usage
type varStore struct {
	// ptr Pointer to C sophia object
	ptr unsafe.Pointer

	// pointers slice of pointers to allocated C variables,
	// that must be freed after store usage
	pointers []unsafe.Pointer
}

func newVarStore(ptr unsafe.Pointer) *varStore {
	return &varStore{
		ptr:      ptr,
		pointers: make([]unsafe.Pointer, 0),
	}
}

// TODO :: implement another types
func (s *varStore) Set(path string, val interface{}) bool {
	v := reflect.ValueOf(val)

	switch v.Kind() {
	case reflect.String:
		return s.SetString(path, v.String())
	case reflect.Int, reflect.Int64, reflect.Int8, reflect.Int16, reflect.Int32:
		return s.SetInt(path, v.Int())
	case reflect.Uint, reflect.Uint64, reflect.Uint8, reflect.Uint16, reflect.Uint32:
		return s.SetInt(path, int64(v.Uint()))
	}

	cPath := cString(path)
	s.pointers = append(s.pointers, unsafe.Pointer(cPath))
	size := int(reflect.TypeOf(val).Size())
	return spSetString(s.ptr, cPath, (unsafe.Pointer)(reflect.ValueOf(val).Pointer()), size)
}

func (s *varStore) SetString(path, val string) bool {
	cPath := cString(path)
	cVal := cString(val)
	s.pointers = append(s.pointers, unsafe.Pointer(cPath), unsafe.Pointer(cVal))
	return spSetString(s.ptr, cPath, unsafe.Pointer(cVal), len(val))
}

func (s *varStore) SetInt(path string, val int64) bool {
	cPath := cString(path)
	s.pointers = append(s.pointers, unsafe.Pointer(cPath))
	return spSetInt(s.ptr, cPath, val)
}

func (s *varStore) Get(path string, size *int) unsafe.Pointer {
	return spGetString(s.ptr, path, size)
}

func (s *varStore) GetString(path string, size *int) string {
	ptr := spGetString(s.ptr, path, size)
	sh := reflect.StringHeader{Data: uintptr(ptr), Len: *size}
	return *(*string)(unsafe.Pointer(&sh))
}

func (s *varStore) GetObject(path string) unsafe.Pointer {
	return spGetObject(s.ptr, path)
}

func (s *varStore) GetInt(key string) int64 {
	return spGetInt(s.ptr, key)
}

// Free frees allocated memory for all C variables, that were in this store
func (s *varStore) Free() {
	for _, f := range s.pointers {
		free(f)
	}
	s.pointers = s.pointers[:0]
}