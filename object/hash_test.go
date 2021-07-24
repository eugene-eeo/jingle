package object_test

import (
	"monkey/object"
	"testing"
)

func TestHashTableBasic(t *testing.T) {
	type keyFactory func() object.Hashable
	factories := []keyFactory{
		keyFactory(func() object.Hashable { return &object.Null{} }),
		keyFactory(func() object.Hashable { return &object.Boolean{Value: true} }),
		keyFactory(func() object.Hashable { return &object.Boolean{Value: false} }),
		keyFactory(func() object.Hashable { return &object.Integer{Value: 20} }),
		keyFactory(func() object.Hashable { return &object.String{Value: "abc"} }),
	}

	// This function exercises the HashEq() method of each
	// Hashable Object, by creating many (pointer-unequal)
	// objects.
	for _, keyFunc := range factories {
		v1 := &object.Null{}
		v2 := &object.Null{}
		ht := object.NewHashTable()
		ht.Set(keyFunc(), v1)
		value, ok := ht.Get(keyFunc())
		if ok != true || value != v1 {
			t.Errorf("invalid ht.Get() expected=(%T(%+v), %t), got=(%T(%+v), %t)",
				v1, v1, true,
				value, value, ok)
		}

		ht.Set(keyFunc(), v2)
		value, ok = ht.Get(keyFunc())
		if ok != true || value != v2 {
			t.Errorf("invalid ht.Get() expected=(%T(%+v), %t), got=(%T(%+v), %t)",
				v2, v2, true,
				value, value, ok)
		}

		// Delete existing key == true
		ok = ht.Delete(keyFunc())
		if !ok {
			t.Errorf("invalid ht.Delete() expected=%t, got=%t",
				true, ok)
		}

		// Delete non-existent key == false
		ok = ht.Delete(keyFunc())
		if ok {
			t.Errorf("invalid ht.Delete() expected=%t, got=%t",
				false, ok)
		}
	}
}

func TestHashTableFuncs(t *testing.T) {
	// These hashables rely on pointer equality...
	ht := object.NewHashTable()
	var (
		k1 = &object.Function{}
		k2 = &object.Function{}
		k3 = &object.Builtin{}
		k4 = &object.Builtin{}
		v1 = &object.Null{}
		v2 = &object.Null{}
		v3 = &object.Null{}
		v4 = &object.Null{}
	)
	keys := []object.Hashable{k1, k2, k3, k4}
	vals := []object.Object{v1, v2, v3, v4}

	// Set the keys...
	for i := range keys {
		ht.Set(keys[i], vals[i])
		if size := ht.Size(); size != uint64(i+1) {
			t.Errorf("expect ht.Size() == %d. got=%d", i+1, size)
		}
		// All should be set!
		for j := 0; j <= i; j++ {
			key := keys[j]
			expected := vals[j]
			got, ok := ht.Get(key)
			if !ok {
				t.Errorf("expect ht.Get(keys[%d]) to be true. got=%t",
					j, ok)
			}
			if got != expected {
				t.Errorf("invalid ht.Get(keys[%d]). expected=%T(%+v) got=%T(%+v)",
					j, expected, expected, got, got)
			}
		}
	}
}

func TestHashTable(t *testing.T) {
	ht := object.NewHashTable()
	var (
		k1 = &object.Boolean{Value: true}
		k2 = &object.Boolean{Value: false}
		k3 = &object.Null{}
		k4 = &object.Integer{Value: 1}
		k5 = &object.String{Value: "abc"}
		v1 = &object.Array{Elements: []object.Object{}}
		v2 = &object.Integer{Value: 100}
		v3 = &object.Integer{Value: 200}
		v4 = &object.Integer{Value: 300}
		v5 = k5
	)

	keys := []object.Hashable{k1, k2, k3, k4, k5}
	vals := []object.Object{v1, v2, v3, v4, v5}

	// Initially, hash map should be empty.
	if size := ht.Size(); size != uint64(0) {
		t.Errorf("expect ht.Size() == 0. got=%d", size)
	}
	for i, key := range keys {
		_, ok := ht.Get(key)
		if ok {
			t.Errorf("expect ht.Get(keys[%d]) to be false. got=%t", i, ok)
		}
	}

	// Set the keys...
	for i := range keys {
		ht.Set(keys[i], vals[i])
		if size := ht.Size(); size != uint64(i+1) {
			t.Errorf("expect ht.Size() == %d. got=%d", i+1, size)
		}
		// All should be set!
		for j := 0; j <= i; j++ {
			key := keys[j]
			expected := vals[j]
			got, ok := ht.Get(key)
			if !ok {
				t.Errorf("expect ht.Get(keys[%d]) to be true. got=%t",
					j, ok)
			}
			if got != expected {
				t.Errorf("invalid ht.Get(keys[%d]). expected=%T(%+v) got=%T(%+v)",
					j, expected, expected, got, got)
			}
		}
	}
}

func TestHashTableIter(t *testing.T) {
	TIMES := int64(10)
	ht := object.NewHashTable()
	values := make([]*object.Integer, TIMES)
	for i := int64(0); i < TIMES; i++ {
		values[i] = &object.Integer{Value: i * 5}
		ht.Set(
			&object.Integer{Value: i},
			values[i],
		)
		seen := map[int64]bool{}
		ht.Iter(func(key object.Hashable, value object.Object) bool {
			intKey := key.(*object.Integer)
			intVal := value.(*object.Integer)
			if intKey.Value < 0 || intKey.Value > i {
				t.Fatalf("expect 0 <= intKey.Value <= i. got=%d", intKey.Value)
			}
			if intVal.Value != intKey.Value*5 {
				t.Fatalf("expect intVal.Value == intKey.Value * 5. expected=%d, got=%d",
					intKey.Value*5,
					intVal.Value)
			}
			seen[intKey.Value] = true
			return true
		})
		if len(seen) != int(i+1) {
			t.Errorf("expect to visit == %d values. got=%d", i+1, len(seen))
		}
	}

	stop := 5
	seen := map[int64]bool{}
	ht.Iter(func(key object.Hashable, value object.Object) bool {
		intKey := key.(*object.Integer)
		intVal := value.(*object.Integer)
		if intKey.Value < 0 || intKey.Value > TIMES {
			t.Fatalf("expect 0 <= intKey.Value <= i. got=%d", intKey.Value)
		}
		if intVal.Value != intKey.Value*5 {
			t.Fatalf("expect intVal.Value == intKey.Value * 5. expected=%d, got=%d",
				intKey.Value*5,
				intVal.Value)
		}
		seen[intKey.Value] = true
		if len(seen) == stop {
			return false
		}
		return true
	})
	if len(seen) != stop {
		t.Errorf("expect to visit == %d values. got=%d", stop, len(seen))
	}
}

func TestHashTableInsertMany(t *testing.T) {
	// Test inserting 1000000 items.
	TIMES := int64(1000 * 1000)
	ht := object.NewHashTable()
	values := make([]*object.Integer, TIMES)
	for i := int64(0); i < TIMES; i++ {
		values[i] = &object.Integer{Value: i * 5}
		ht.Set(
			&object.Integer{Value: i},
			values[i],
		)
		if size := ht.Size(); size != uint64(i+1) {
			t.Errorf("expect ht.Size() == %d. got=%d", i+1, size)
		}
	}
	for i := int64(0); i < TIMES; i++ {
		// Different objects, should still get the same thing back!
		v, ok := ht.Get(&object.Integer{Value: i})
		if !ok {
			t.Fatalf("key %d went missing", i)
		}
		if v != values[i] {
			t.Fatalf("ht.Get(%d) wrong. expected=%T(%+v), got=%T(%+v)",
				i, values[i], values[i], v, v)
		}
	}
	for i := int64(0); i < TIMES; i++ {
		// Delete all the values...
		ok := ht.Delete(&object.Integer{Value: i})
		if !ok {
			t.Fatalf("key %d went missing", i)
		}
		if size := ht.Size(); size != uint64(TIMES-i-1) {
			t.Fatalf("expect ht.Size() == %d. got=%d", TIMES-i-1, size)
		}
	}
	// Check that we can still insert something.
	// Regression where we forget to bound the limit where we can shrink
	// the hash map in ht.Delete() :(.
	ht.Set(
		&object.String{Value: "abracadabra"},
		&object.String{Value: "here"},
	)
}
