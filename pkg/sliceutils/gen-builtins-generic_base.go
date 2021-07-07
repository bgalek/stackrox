// Code generated by genny. DO NOT EDIT.
// This file was automatically generated by genny.
// Any changes will be lost if this file is regenerated.
// see https://github.com/mauricelam/genny

package sliceutils

import (
	"github.com/pkg/errors"
)

// BoolSelect returns a slice containing the elements at the given indices of the input slice.
// CAUTION: This function panics if any index is out of range.
func BoolSelect(a []bool, indices ...int) []bool {
	result := make([]bool, 0, len(indices))
	for _, idx := range indices {
		if idx < 0 || idx >= len(a) {
			panic(errors.Errorf("invalid index %d: outside of expected range [0, %d)", idx, len(a)))
		}
		result = append(result, a[idx])
	}
	return result
}

// BoolClone clones a slice, creating a new slice
// and copying the contents of the underlying array.
// If `in` is a nil slice, a nil slice is returned.
// If `in` is an empty slice, an empty slice is returned.
func BoolClone(in []bool) []bool {
	if in == nil {
		return nil
	}
	if len(in) == 0 {
		return []bool{}
	}
	out := make([]bool, len(in))
	copy(out, in)
	return out
}

// ConcatBoolSlices concatenates slices, returning a slice with newly allocated backing storage of the exact
// size.
func ConcatBoolSlices(slices ...[]bool) []bool {
	length := 0
	for _, slice := range slices {
		length += len(slice)
	}
	result := make([]bool, length)
	i := 0
	for _, slice := range slices {
		nextI := i + len(slice)
		copy(result[i:nextI], slice)
		i = nextI
	}
	return result
}

// ByteSelect returns a slice containing the elements at the given indices of the input slice.
// CAUTION: This function panics if any index is out of range.
func ByteSelect(a []byte, indices ...int) []byte {
	result := make([]byte, 0, len(indices))
	for _, idx := range indices {
		if idx < 0 || idx >= len(a) {
			panic(errors.Errorf("invalid index %d: outside of expected range [0, %d)", idx, len(a)))
		}
		result = append(result, a[idx])
	}
	return result
}

// ByteClone clones a slice, creating a new slice
// and copying the contents of the underlying array.
// If `in` is a nil slice, a nil slice is returned.
// If `in` is an empty slice, an empty slice is returned.
func ByteClone(in []byte) []byte {
	if in == nil {
		return nil
	}
	if len(in) == 0 {
		return []byte{}
	}
	out := make([]byte, len(in))
	copy(out, in)
	return out
}

// ConcatByteSlices concatenates slices, returning a slice with newly allocated backing storage of the exact
// size.
func ConcatByteSlices(slices ...[]byte) []byte {
	length := 0
	for _, slice := range slices {
		length += len(slice)
	}
	result := make([]byte, length)
	i := 0
	for _, slice := range slices {
		nextI := i + len(slice)
		copy(result[i:nextI], slice)
		i = nextI
	}
	return result
}

// Complex128Select returns a slice containing the elements at the given indices of the input slice.
// CAUTION: This function panics if any index is out of range.
func Complex128Select(a []complex128, indices ...int) []complex128 {
	result := make([]complex128, 0, len(indices))
	for _, idx := range indices {
		if idx < 0 || idx >= len(a) {
			panic(errors.Errorf("invalid index %d: outside of expected range [0, %d)", idx, len(a)))
		}
		result = append(result, a[idx])
	}
	return result
}

// Complex128Clone clones a slice, creating a new slice
// and copying the contents of the underlying array.
// If `in` is a nil slice, a nil slice is returned.
// If `in` is an empty slice, an empty slice is returned.
func Complex128Clone(in []complex128) []complex128 {
	if in == nil {
		return nil
	}
	if len(in) == 0 {
		return []complex128{}
	}
	out := make([]complex128, len(in))
	copy(out, in)
	return out
}

// ConcatComplex128Slices concatenates slices, returning a slice with newly allocated backing storage of the exact
// size.
func ConcatComplex128Slices(slices ...[]complex128) []complex128 {
	length := 0
	for _, slice := range slices {
		length += len(slice)
	}
	result := make([]complex128, length)
	i := 0
	for _, slice := range slices {
		nextI := i + len(slice)
		copy(result[i:nextI], slice)
		i = nextI
	}
	return result
}

// Complex64Select returns a slice containing the elements at the given indices of the input slice.
// CAUTION: This function panics if any index is out of range.
func Complex64Select(a []complex64, indices ...int) []complex64 {
	result := make([]complex64, 0, len(indices))
	for _, idx := range indices {
		if idx < 0 || idx >= len(a) {
			panic(errors.Errorf("invalid index %d: outside of expected range [0, %d)", idx, len(a)))
		}
		result = append(result, a[idx])
	}
	return result
}

// Complex64Clone clones a slice, creating a new slice
// and copying the contents of the underlying array.
// If `in` is a nil slice, a nil slice is returned.
// If `in` is an empty slice, an empty slice is returned.
func Complex64Clone(in []complex64) []complex64 {
	if in == nil {
		return nil
	}
	if len(in) == 0 {
		return []complex64{}
	}
	out := make([]complex64, len(in))
	copy(out, in)
	return out
}

// ConcatComplex64Slices concatenates slices, returning a slice with newly allocated backing storage of the exact
// size.
func ConcatComplex64Slices(slices ...[]complex64) []complex64 {
	length := 0
	for _, slice := range slices {
		length += len(slice)
	}
	result := make([]complex64, length)
	i := 0
	for _, slice := range slices {
		nextI := i + len(slice)
		copy(result[i:nextI], slice)
		i = nextI
	}
	return result
}

// ErrorSelect returns a slice containing the elements at the given indices of the input slice.
// CAUTION: This function panics if any index is out of range.
func ErrorSelect(a []error, indices ...int) []error {
	result := make([]error, 0, len(indices))
	for _, idx := range indices {
		if idx < 0 || idx >= len(a) {
			panic(errors.Errorf("invalid index %d: outside of expected range [0, %d)", idx, len(a)))
		}
		result = append(result, a[idx])
	}
	return result
}

// ErrorClone clones a slice, creating a new slice
// and copying the contents of the underlying array.
// If `in` is a nil slice, a nil slice is returned.
// If `in` is an empty slice, an empty slice is returned.
func ErrorClone(in []error) []error {
	if in == nil {
		return nil
	}
	if len(in) == 0 {
		return []error{}
	}
	out := make([]error, len(in))
	copy(out, in)
	return out
}

// ConcatErrorSlices concatenates slices, returning a slice with newly allocated backing storage of the exact
// size.
func ConcatErrorSlices(slices ...[]error) []error {
	length := 0
	for _, slice := range slices {
		length += len(slice)
	}
	result := make([]error, length)
	i := 0
	for _, slice := range slices {
		nextI := i + len(slice)
		copy(result[i:nextI], slice)
		i = nextI
	}
	return result
}

// Float32Select returns a slice containing the elements at the given indices of the input slice.
// CAUTION: This function panics if any index is out of range.
func Float32Select(a []float32, indices ...int) []float32 {
	result := make([]float32, 0, len(indices))
	for _, idx := range indices {
		if idx < 0 || idx >= len(a) {
			panic(errors.Errorf("invalid index %d: outside of expected range [0, %d)", idx, len(a)))
		}
		result = append(result, a[idx])
	}
	return result
}

// Float32Clone clones a slice, creating a new slice
// and copying the contents of the underlying array.
// If `in` is a nil slice, a nil slice is returned.
// If `in` is an empty slice, an empty slice is returned.
func Float32Clone(in []float32) []float32 {
	if in == nil {
		return nil
	}
	if len(in) == 0 {
		return []float32{}
	}
	out := make([]float32, len(in))
	copy(out, in)
	return out
}

// ConcatFloat32Slices concatenates slices, returning a slice with newly allocated backing storage of the exact
// size.
func ConcatFloat32Slices(slices ...[]float32) []float32 {
	length := 0
	for _, slice := range slices {
		length += len(slice)
	}
	result := make([]float32, length)
	i := 0
	for _, slice := range slices {
		nextI := i + len(slice)
		copy(result[i:nextI], slice)
		i = nextI
	}
	return result
}

// Float64Select returns a slice containing the elements at the given indices of the input slice.
// CAUTION: This function panics if any index is out of range.
func Float64Select(a []float64, indices ...int) []float64 {
	result := make([]float64, 0, len(indices))
	for _, idx := range indices {
		if idx < 0 || idx >= len(a) {
			panic(errors.Errorf("invalid index %d: outside of expected range [0, %d)", idx, len(a)))
		}
		result = append(result, a[idx])
	}
	return result
}

// Float64Clone clones a slice, creating a new slice
// and copying the contents of the underlying array.
// If `in` is a nil slice, a nil slice is returned.
// If `in` is an empty slice, an empty slice is returned.
func Float64Clone(in []float64) []float64 {
	if in == nil {
		return nil
	}
	if len(in) == 0 {
		return []float64{}
	}
	out := make([]float64, len(in))
	copy(out, in)
	return out
}

// ConcatFloat64Slices concatenates slices, returning a slice with newly allocated backing storage of the exact
// size.
func ConcatFloat64Slices(slices ...[]float64) []float64 {
	length := 0
	for _, slice := range slices {
		length += len(slice)
	}
	result := make([]float64, length)
	i := 0
	for _, slice := range slices {
		nextI := i + len(slice)
		copy(result[i:nextI], slice)
		i = nextI
	}
	return result
}

// IntSelect returns a slice containing the elements at the given indices of the input slice.
// CAUTION: This function panics if any index is out of range.
func IntSelect(a []int, indices ...int) []int {
	result := make([]int, 0, len(indices))
	for _, idx := range indices {
		if idx < 0 || idx >= len(a) {
			panic(errors.Errorf("invalid index %d: outside of expected range [0, %d)", idx, len(a)))
		}
		result = append(result, a[idx])
	}
	return result
}

// IntClone clones a slice, creating a new slice
// and copying the contents of the underlying array.
// If `in` is a nil slice, a nil slice is returned.
// If `in` is an empty slice, an empty slice is returned.
func IntClone(in []int) []int {
	if in == nil {
		return nil
	}
	if len(in) == 0 {
		return []int{}
	}
	out := make([]int, len(in))
	copy(out, in)
	return out
}

// ConcatIntSlices concatenates slices, returning a slice with newly allocated backing storage of the exact
// size.
func ConcatIntSlices(slices ...[]int) []int {
	length := 0
	for _, slice := range slices {
		length += len(slice)
	}
	result := make([]int, length)
	i := 0
	for _, slice := range slices {
		nextI := i + len(slice)
		copy(result[i:nextI], slice)
		i = nextI
	}
	return result
}

// Int16Select returns a slice containing the elements at the given indices of the input slice.
// CAUTION: This function panics if any index is out of range.
func Int16Select(a []int16, indices ...int) []int16 {
	result := make([]int16, 0, len(indices))
	for _, idx := range indices {
		if idx < 0 || idx >= len(a) {
			panic(errors.Errorf("invalid index %d: outside of expected range [0, %d)", idx, len(a)))
		}
		result = append(result, a[idx])
	}
	return result
}

// Int16Clone clones a slice, creating a new slice
// and copying the contents of the underlying array.
// If `in` is a nil slice, a nil slice is returned.
// If `in` is an empty slice, an empty slice is returned.
func Int16Clone(in []int16) []int16 {
	if in == nil {
		return nil
	}
	if len(in) == 0 {
		return []int16{}
	}
	out := make([]int16, len(in))
	copy(out, in)
	return out
}

// ConcatInt16Slices concatenates slices, returning a slice with newly allocated backing storage of the exact
// size.
func ConcatInt16Slices(slices ...[]int16) []int16 {
	length := 0
	for _, slice := range slices {
		length += len(slice)
	}
	result := make([]int16, length)
	i := 0
	for _, slice := range slices {
		nextI := i + len(slice)
		copy(result[i:nextI], slice)
		i = nextI
	}
	return result
}

// Int32Select returns a slice containing the elements at the given indices of the input slice.
// CAUTION: This function panics if any index is out of range.
func Int32Select(a []int32, indices ...int) []int32 {
	result := make([]int32, 0, len(indices))
	for _, idx := range indices {
		if idx < 0 || idx >= len(a) {
			panic(errors.Errorf("invalid index %d: outside of expected range [0, %d)", idx, len(a)))
		}
		result = append(result, a[idx])
	}
	return result
}

// Int32Clone clones a slice, creating a new slice
// and copying the contents of the underlying array.
// If `in` is a nil slice, a nil slice is returned.
// If `in` is an empty slice, an empty slice is returned.
func Int32Clone(in []int32) []int32 {
	if in == nil {
		return nil
	}
	if len(in) == 0 {
		return []int32{}
	}
	out := make([]int32, len(in))
	copy(out, in)
	return out
}

// ConcatInt32Slices concatenates slices, returning a slice with newly allocated backing storage of the exact
// size.
func ConcatInt32Slices(slices ...[]int32) []int32 {
	length := 0
	for _, slice := range slices {
		length += len(slice)
	}
	result := make([]int32, length)
	i := 0
	for _, slice := range slices {
		nextI := i + len(slice)
		copy(result[i:nextI], slice)
		i = nextI
	}
	return result
}

// Int64Select returns a slice containing the elements at the given indices of the input slice.
// CAUTION: This function panics if any index is out of range.
func Int64Select(a []int64, indices ...int) []int64 {
	result := make([]int64, 0, len(indices))
	for _, idx := range indices {
		if idx < 0 || idx >= len(a) {
			panic(errors.Errorf("invalid index %d: outside of expected range [0, %d)", idx, len(a)))
		}
		result = append(result, a[idx])
	}
	return result
}

// Int64Clone clones a slice, creating a new slice
// and copying the contents of the underlying array.
// If `in` is a nil slice, a nil slice is returned.
// If `in` is an empty slice, an empty slice is returned.
func Int64Clone(in []int64) []int64 {
	if in == nil {
		return nil
	}
	if len(in) == 0 {
		return []int64{}
	}
	out := make([]int64, len(in))
	copy(out, in)
	return out
}

// ConcatInt64Slices concatenates slices, returning a slice with newly allocated backing storage of the exact
// size.
func ConcatInt64Slices(slices ...[]int64) []int64 {
	length := 0
	for _, slice := range slices {
		length += len(slice)
	}
	result := make([]int64, length)
	i := 0
	for _, slice := range slices {
		nextI := i + len(slice)
		copy(result[i:nextI], slice)
		i = nextI
	}
	return result
}

// Int8Select returns a slice containing the elements at the given indices of the input slice.
// CAUTION: This function panics if any index is out of range.
func Int8Select(a []int8, indices ...int) []int8 {
	result := make([]int8, 0, len(indices))
	for _, idx := range indices {
		if idx < 0 || idx >= len(a) {
			panic(errors.Errorf("invalid index %d: outside of expected range [0, %d)", idx, len(a)))
		}
		result = append(result, a[idx])
	}
	return result
}

// Int8Clone clones a slice, creating a new slice
// and copying the contents of the underlying array.
// If `in` is a nil slice, a nil slice is returned.
// If `in` is an empty slice, an empty slice is returned.
func Int8Clone(in []int8) []int8 {
	if in == nil {
		return nil
	}
	if len(in) == 0 {
		return []int8{}
	}
	out := make([]int8, len(in))
	copy(out, in)
	return out
}

// ConcatInt8Slices concatenates slices, returning a slice with newly allocated backing storage of the exact
// size.
func ConcatInt8Slices(slices ...[]int8) []int8 {
	length := 0
	for _, slice := range slices {
		length += len(slice)
	}
	result := make([]int8, length)
	i := 0
	for _, slice := range slices {
		nextI := i + len(slice)
		copy(result[i:nextI], slice)
		i = nextI
	}
	return result
}

// RuneSelect returns a slice containing the elements at the given indices of the input slice.
// CAUTION: This function panics if any index is out of range.
func RuneSelect(a []rune, indices ...int) []rune {
	result := make([]rune, 0, len(indices))
	for _, idx := range indices {
		if idx < 0 || idx >= len(a) {
			panic(errors.Errorf("invalid index %d: outside of expected range [0, %d)", idx, len(a)))
		}
		result = append(result, a[idx])
	}
	return result
}

// RuneClone clones a slice, creating a new slice
// and copying the contents of the underlying array.
// If `in` is a nil slice, a nil slice is returned.
// If `in` is an empty slice, an empty slice is returned.
func RuneClone(in []rune) []rune {
	if in == nil {
		return nil
	}
	if len(in) == 0 {
		return []rune{}
	}
	out := make([]rune, len(in))
	copy(out, in)
	return out
}

// ConcatRuneSlices concatenates slices, returning a slice with newly allocated backing storage of the exact
// size.
func ConcatRuneSlices(slices ...[]rune) []rune {
	length := 0
	for _, slice := range slices {
		length += len(slice)
	}
	result := make([]rune, length)
	i := 0
	for _, slice := range slices {
		nextI := i + len(slice)
		copy(result[i:nextI], slice)
		i = nextI
	}
	return result
}

// StringSelect returns a slice containing the elements at the given indices of the input slice.
// CAUTION: This function panics if any index is out of range.
func StringSelect(a []string, indices ...int) []string {
	result := make([]string, 0, len(indices))
	for _, idx := range indices {
		if idx < 0 || idx >= len(a) {
			panic(errors.Errorf("invalid index %d: outside of expected range [0, %d)", idx, len(a)))
		}
		result = append(result, a[idx])
	}
	return result
}

// StringClone clones a slice, creating a new slice
// and copying the contents of the underlying array.
// If `in` is a nil slice, a nil slice is returned.
// If `in` is an empty slice, an empty slice is returned.
func StringClone(in []string) []string {
	if in == nil {
		return nil
	}
	if len(in) == 0 {
		return []string{}
	}
	out := make([]string, len(in))
	copy(out, in)
	return out
}

// ConcatStringSlices concatenates slices, returning a slice with newly allocated backing storage of the exact
// size.
func ConcatStringSlices(slices ...[]string) []string {
	length := 0
	for _, slice := range slices {
		length += len(slice)
	}
	result := make([]string, length)
	i := 0
	for _, slice := range slices {
		nextI := i + len(slice)
		copy(result[i:nextI], slice)
		i = nextI
	}
	return result
}

// UintSelect returns a slice containing the elements at the given indices of the input slice.
// CAUTION: This function panics if any index is out of range.
func UintSelect(a []uint, indices ...int) []uint {
	result := make([]uint, 0, len(indices))
	for _, idx := range indices {
		if idx < 0 || idx >= len(a) {
			panic(errors.Errorf("invalid index %d: outside of expected range [0, %d)", idx, len(a)))
		}
		result = append(result, a[idx])
	}
	return result
}

// UintClone clones a slice, creating a new slice
// and copying the contents of the underlying array.
// If `in` is a nil slice, a nil slice is returned.
// If `in` is an empty slice, an empty slice is returned.
func UintClone(in []uint) []uint {
	if in == nil {
		return nil
	}
	if len(in) == 0 {
		return []uint{}
	}
	out := make([]uint, len(in))
	copy(out, in)
	return out
}

// ConcatUintSlices concatenates slices, returning a slice with newly allocated backing storage of the exact
// size.
func ConcatUintSlices(slices ...[]uint) []uint {
	length := 0
	for _, slice := range slices {
		length += len(slice)
	}
	result := make([]uint, length)
	i := 0
	for _, slice := range slices {
		nextI := i + len(slice)
		copy(result[i:nextI], slice)
		i = nextI
	}
	return result
}

// Uint16Select returns a slice containing the elements at the given indices of the input slice.
// CAUTION: This function panics if any index is out of range.
func Uint16Select(a []uint16, indices ...int) []uint16 {
	result := make([]uint16, 0, len(indices))
	for _, idx := range indices {
		if idx < 0 || idx >= len(a) {
			panic(errors.Errorf("invalid index %d: outside of expected range [0, %d)", idx, len(a)))
		}
		result = append(result, a[idx])
	}
	return result
}

// Uint16Clone clones a slice, creating a new slice
// and copying the contents of the underlying array.
// If `in` is a nil slice, a nil slice is returned.
// If `in` is an empty slice, an empty slice is returned.
func Uint16Clone(in []uint16) []uint16 {
	if in == nil {
		return nil
	}
	if len(in) == 0 {
		return []uint16{}
	}
	out := make([]uint16, len(in))
	copy(out, in)
	return out
}

// ConcatUint16Slices concatenates slices, returning a slice with newly allocated backing storage of the exact
// size.
func ConcatUint16Slices(slices ...[]uint16) []uint16 {
	length := 0
	for _, slice := range slices {
		length += len(slice)
	}
	result := make([]uint16, length)
	i := 0
	for _, slice := range slices {
		nextI := i + len(slice)
		copy(result[i:nextI], slice)
		i = nextI
	}
	return result
}

// Uint32Select returns a slice containing the elements at the given indices of the input slice.
// CAUTION: This function panics if any index is out of range.
func Uint32Select(a []uint32, indices ...int) []uint32 {
	result := make([]uint32, 0, len(indices))
	for _, idx := range indices {
		if idx < 0 || idx >= len(a) {
			panic(errors.Errorf("invalid index %d: outside of expected range [0, %d)", idx, len(a)))
		}
		result = append(result, a[idx])
	}
	return result
}

// Uint32Clone clones a slice, creating a new slice
// and copying the contents of the underlying array.
// If `in` is a nil slice, a nil slice is returned.
// If `in` is an empty slice, an empty slice is returned.
func Uint32Clone(in []uint32) []uint32 {
	if in == nil {
		return nil
	}
	if len(in) == 0 {
		return []uint32{}
	}
	out := make([]uint32, len(in))
	copy(out, in)
	return out
}

// ConcatUint32Slices concatenates slices, returning a slice with newly allocated backing storage of the exact
// size.
func ConcatUint32Slices(slices ...[]uint32) []uint32 {
	length := 0
	for _, slice := range slices {
		length += len(slice)
	}
	result := make([]uint32, length)
	i := 0
	for _, slice := range slices {
		nextI := i + len(slice)
		copy(result[i:nextI], slice)
		i = nextI
	}
	return result
}

// Uint64Select returns a slice containing the elements at the given indices of the input slice.
// CAUTION: This function panics if any index is out of range.
func Uint64Select(a []uint64, indices ...int) []uint64 {
	result := make([]uint64, 0, len(indices))
	for _, idx := range indices {
		if idx < 0 || idx >= len(a) {
			panic(errors.Errorf("invalid index %d: outside of expected range [0, %d)", idx, len(a)))
		}
		result = append(result, a[idx])
	}
	return result
}

// Uint64Clone clones a slice, creating a new slice
// and copying the contents of the underlying array.
// If `in` is a nil slice, a nil slice is returned.
// If `in` is an empty slice, an empty slice is returned.
func Uint64Clone(in []uint64) []uint64 {
	if in == nil {
		return nil
	}
	if len(in) == 0 {
		return []uint64{}
	}
	out := make([]uint64, len(in))
	copy(out, in)
	return out
}

// ConcatUint64Slices concatenates slices, returning a slice with newly allocated backing storage of the exact
// size.
func ConcatUint64Slices(slices ...[]uint64) []uint64 {
	length := 0
	for _, slice := range slices {
		length += len(slice)
	}
	result := make([]uint64, length)
	i := 0
	for _, slice := range slices {
		nextI := i + len(slice)
		copy(result[i:nextI], slice)
		i = nextI
	}
	return result
}

// Uint8Select returns a slice containing the elements at the given indices of the input slice.
// CAUTION: This function panics if any index is out of range.
func Uint8Select(a []uint8, indices ...int) []uint8 {
	result := make([]uint8, 0, len(indices))
	for _, idx := range indices {
		if idx < 0 || idx >= len(a) {
			panic(errors.Errorf("invalid index %d: outside of expected range [0, %d)", idx, len(a)))
		}
		result = append(result, a[idx])
	}
	return result
}

// Uint8Clone clones a slice, creating a new slice
// and copying the contents of the underlying array.
// If `in` is a nil slice, a nil slice is returned.
// If `in` is an empty slice, an empty slice is returned.
func Uint8Clone(in []uint8) []uint8 {
	if in == nil {
		return nil
	}
	if len(in) == 0 {
		return []uint8{}
	}
	out := make([]uint8, len(in))
	copy(out, in)
	return out
}

// ConcatUint8Slices concatenates slices, returning a slice with newly allocated backing storage of the exact
// size.
func ConcatUint8Slices(slices ...[]uint8) []uint8 {
	length := 0
	for _, slice := range slices {
		length += len(slice)
	}
	result := make([]uint8, length)
	i := 0
	for _, slice := range slices {
		nextI := i + len(slice)
		copy(result[i:nextI], slice)
		i = nextI
	}
	return result
}

// UintptrSelect returns a slice containing the elements at the given indices of the input slice.
// CAUTION: This function panics if any index is out of range.
func UintptrSelect(a []uintptr, indices ...int) []uintptr {
	result := make([]uintptr, 0, len(indices))
	for _, idx := range indices {
		if idx < 0 || idx >= len(a) {
			panic(errors.Errorf("invalid index %d: outside of expected range [0, %d)", idx, len(a)))
		}
		result = append(result, a[idx])
	}
	return result
}

// UintptrClone clones a slice, creating a new slice
// and copying the contents of the underlying array.
// If `in` is a nil slice, a nil slice is returned.
// If `in` is an empty slice, an empty slice is returned.
func UintptrClone(in []uintptr) []uintptr {
	if in == nil {
		return nil
	}
	if len(in) == 0 {
		return []uintptr{}
	}
	out := make([]uintptr, len(in))
	copy(out, in)
	return out
}

// ConcatUintptrSlices concatenates slices, returning a slice with newly allocated backing storage of the exact
// size.
func ConcatUintptrSlices(slices ...[]uintptr) []uintptr {
	length := 0
	for _, slice := range slices {
		length += len(slice)
	}
	result := make([]uintptr, length)
	i := 0
	for _, slice := range slices {
		nextI := i + len(slice)
		copy(result[i:nextI], slice)
		i = nextI
	}
	return result
}

// ByteSliceSelect returns a slice containing the elements at the given indices of the input slice.
// CAUTION: This function panics if any index is out of range.
func ByteSliceSelect(a []ByteSlice, indices ...int) []ByteSlice {
	result := make([]ByteSlice, 0, len(indices))
	for _, idx := range indices {
		if idx < 0 || idx >= len(a) {
			panic(errors.Errorf("invalid index %d: outside of expected range [0, %d)", idx, len(a)))
		}
		result = append(result, a[idx])
	}
	return result
}

// ByteSliceClone clones a slice, creating a new slice
// and copying the contents of the underlying array.
// If `in` is a nil slice, a nil slice is returned.
// If `in` is an empty slice, an empty slice is returned.
func ByteSliceClone(in []ByteSlice) []ByteSlice {
	if in == nil {
		return nil
	}
	if len(in) == 0 {
		return []ByteSlice{}
	}
	out := make([]ByteSlice, len(in))
	copy(out, in)
	return out
}

// ConcatByteSliceSlices concatenates slices, returning a slice with newly allocated backing storage of the exact
// size.
func ConcatByteSliceSlices(slices ...[]ByteSlice) []ByteSlice {
	length := 0
	for _, slice := range slices {
		length += len(slice)
	}
	result := make([]ByteSlice, length)
	i := 0
	for _, slice := range slices {
		nextI := i + len(slice)
		copy(result[i:nextI], slice)
		i = nextI
	}
	return result
}