package parse

import (
	"slices"
)

func FindAll[T any](elements []interface{}) []T {
	var list []T
	find(elements, func(t T) bool {
		list = append(list, t)
		return false
	})
	return list
}

func FindFirst[T any](elements []interface{}) T {
	var found T
	find(elements, func(t T) bool {
		found = t
		return true
	})
	return found
}

func Search[T any](elements []interface{}, callback func(t T) bool) {
	find(elements, callback)
}

func find[T any](elements []interface{}, callback func(t T) bool) bool {
	for _, e := range elements {
		var shortCircuit bool
		if el, ok := e.(T); ok {
			shortCircuit = callback(el)
		} else if ae, ok := e.(HasBase); ok {
			be := ae.GetBase()
			if el, ok := be.(T); ok {
				shortCircuit = callback(el)
			}
		}
		if shortCircuit {
			return true
		}
		if he, ok := e.(HasElements); ok {
			shortCircuit = find(he.GetElements(), callback)
		} else if ae, ok := e.(HasBase); ok {
			be := ae.GetBase()
			if el, ok := be.(HasElements); ok {
				shortCircuit = find(el.GetElements(), callback)
			}
		}
		if shortCircuit {
			return true
		}
	}
	return false
}

func Skim[T any](elements []interface{}) []T {
	var list []T
	for _, e := range elements {
		if ae, ok := e.(HasBase); ok {
			e = ae.GetBase()
		}
		switch el := e.(type) {
		case T:
			list = append(list, el)
		}

	}
	return list
}

func SkimFunc[T any](elements []interface{}, callback func(t T) bool) bool {
	for _, e := range elements {
		var shortCircuit bool
		if el, ok := e.(T); ok {
			shortCircuit = callback(el)
		} else if ae, ok := e.(HasBase); ok {
			be := ae.GetBase()
			if el, ok := be.(T); ok {
				shortCircuit = callback(el)
			}
		}
		if shortCircuit {
			return true
		}
	}
	return false
}

func Filter(parent HasElements, callback func(i interface{}) (remove bool, shortCircuit bool)) (shortCircuit bool) {
	i := 0
	elements := parent.GetElements()
	var removed bool
	for i < len(elements) {
		e := elements[i]
		if ae, ok := e.(HasBase); ok {
			e = ae.GetBase()
		}
		switch el := e.(type) {
		case HasElements:
			shortCircuit = Filter(el, callback)
		}
		if shortCircuit {
			break
		}
		remove, shortCircuit := callback(e)
		var empty []interface{}
		if remove {
			elements = slices.Replace(elements, i, i+1, empty...)
			removed = true
		} else {
			i++
		}
		if shortCircuit {
			break
		}
	}
	if removed {
		_ = parent.SetElements(elements)
	}
	return
}

func Traverse[T any](parent HasElements, elements []interface{}, callback func(el T, parent HasElements, index int) bool) {
	traverse(parent, elements, callback)
}

func traverse[T any](parent HasElements, elements []interface{}, callback func(el T, parent HasElements, index int) bool) bool {
	for i, e := range elements {
		if v, ok := e.(T); ok && callback(v, parent, i) {
			return true
		}
	}
	for _, e := range elements {
		switch el := e.(type) {
		case HasElements:
			if traverse(el, el.GetElements(), callback) {
				return true
			}
		}
	}
	return false
}
