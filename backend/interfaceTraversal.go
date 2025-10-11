package main

import (
	"fmt"
	"strings"
)

// returns a value in m based on key. basically, key="x.y.z" will return m["x"]["y"]["z"].
func IT_RawValue(m map[string]any, key string) (any, error) {
	var interf any = m
	for _, k := range strings.Split(key, ".") {
		nested, ok := interf.(map[string]any)
		if !ok {
			return nil, fmt.Errorf("not a map at %q", k)
		}
		interf, ok = nested[k]
		if !ok {
			return nil, fmt.Errorf("couldn't find %q", key)
		}
	}

	return interf, nil
}

// sets a value in m to val based on key. basically, key="x.y.z" will set m["x"]["y"]["z"] to val.
func IT_Set(m map[string]any, key string, val any) error {
	var interf any = m
	for i, k := range strings.Split(key, ".") {
		nested, ok := interf.(map[string]any)
		if !ok {
			return fmt.Errorf("not a map at %s", key)
		}

		if i == len(strings.Split(key, "."))-1 {
			nested[k] = val
			return nil
		}

		if _, ok := nested[k]; !ok {
			nested[k] = make(map[string]any)
		}
		interf = nested[k]
	}

	return nil
}

// returns T in m based on key. basically, key="x.y.z" will return m["x"]["y"]["z"].
func IT_Must[T any](m map[string]any, key string, must T) T {
	interf, err := IT_RawValue(m, key)
	if err != nil {
		return must
	}

	val, ok := interf.(T)
	if !ok {
		return must
	}
	return val
}

// returns T in m based on key, using H_Cast[T]. basically, key="x.y.z" will return m["x"]["y"]["z"].
func IT_MustNumber[T Number](m map[string]any, key string, must T) T {
	interf, err := IT_RawValue(m, key)
	if err != nil {
		return must
	}

	casted, err := H_Cast[T](interf)
	if err != nil {
		return must
	}

	return casted
}
