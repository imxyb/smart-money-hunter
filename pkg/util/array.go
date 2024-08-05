package util

import (
	"reflect"
)

// ObjectEqual checks if objects are equal
func ObjectEqual(a, b interface{}) bool {
	if a == nil || b == nil {
		return a == b
	}
	return reflect.DeepEqual(a, b)
}

// ArrayContains checks if an array contains the element
func ArrayContains(l, elem interface{}) bool {
	vl := reflect.ValueOf(l)
	for i := 0; i < vl.Len(); i++ {
		if ObjectEqual(vl.Index(i).Interface(), elem) {
			return true
		}
	}
	return false
}

// Equal checks if two arrays are equal without the order
func Equal(a, b interface{}) bool {
	va := reflect.ValueOf(a)
	vb := reflect.ValueOf(b)
	for i := 0; i < va.Len(); i++ {
		if !ArrayContains(b, va.Index(i).Interface()) {
			return false
		}
	}
	for i := 0; i < vb.Len(); i++ {
		if !ArrayContains(a, vb.Index(i).Interface()) {
			return false
		}
	}
	return true
}

func Product(sets ...[]interface{}) [][]interface{} {
	lens := func(i int) int { return len(sets[i]) }
	var product [][]interface{}
	for ix := make([]int, len(sets)); ix[0] < lens(0); nextIndex(ix, lens) {
		var r []interface{}
		for j, k := range ix {
			r = append(r, sets[j][k])
		}
		product = append(product, r)
	}
	return product
}

func nextIndex(ix []int, lens func(i int) int) {
	for j := len(ix) - 1; j >= 0; j-- {
		ix[j]++
		if j == 0 || ix[j] < lens(j) {
			return
		}
		ix[j] = 0
	}
}

// DeleteSliceElms delete int from slice
func DeleteSliceElms(sl []int, elms ...int) []int {
	if len(sl) == 0 || len(elms) == 0 {
		return sl
	}
	m := make(map[int]struct{})
	for _, v := range elms {
		m[v] = struct{}{}
	}
	res := make([]int, 0, len(sl))
	for _, v := range sl {
		if _, ok := m[v]; !ok {
			res = append(res, v)
		}
	}
	return res
}

func HasDuplicatesInStrings(strings []string) bool {
	seen := make(map[string]bool)
	for _, s := range strings {
		if _, ok := seen[s]; ok {
			return true
		}
		seen[s] = true
	}
	return false
}
