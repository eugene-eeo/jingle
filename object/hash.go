package object

import (
	"hash/fnv"
	"math/rand"
	"unsafe"
)

const (
	hash_mask  = uint64(0xABBAABBA) // hash mask for integers...
	null_hash  = uint64(42)
	true_hash  = uint64(0xDEADBEEF)
	false_hash = uint64(0xBEEFDEAD)
)

// ================
// Hashable Objects
// ================

type Hashable interface {
	// A hashable object -- needs to implement the Hash()
	// method, as well as the HashEq() method (in case of)
	// collisions. The object passed to HashEq() can be
	// assumed to have the same type.
	Object
	HashEq(Object) bool
	Hash() uint64
}

func (b *Boolean) HashEq(other Object) bool { return b.Value == other.(*Boolean).Value }
func (b *Boolean) Hash() uint64 {
	if b.Value {
		return true_hash
	} else {
		return false_hash
	}
}

func (n *Null) HashEq(other Object) bool { return true }
func (n *Null) Hash() uint64             { return null_hash }

func (il *Integer) HashEq(other Object) bool { return il.Value == other.(*Integer).Value }
func (il *Integer) Hash() uint64             { return uint64(il.Value) ^ hash_mask }

func (s *String) HashEq(other Object) bool { return s.Value == other.(*String).Value }
func (s *String) Hash() uint64 {
	h := fnv.New64a()
	h.Write([]byte(s.Value))
	return h.Sum64()
}

func (bo *Builtin) HashEq(other Object) bool { return bo == other }
func (bo *Builtin) Hash() uint64 {
	return uint64(uintptr(unsafe.Pointer(bo)))
}

func (fn *Function) HashEq(other Object) bool { return fn == other }
func (fn *Function) Hash() uint64 {
	return uint64(uintptr(unsafe.Pointer(fn)))
}

// ================================
// Actual hash table implementation
// ================================

const (
	hashTableMinSize = 8
	// When deleting values, if the size / tableSize ratio
	// falls below this then we rehash the table.
	hashTableRehashRatio = 0.25
)

type HashTableEntry struct {
	Key   Hashable
	Value Object
}

func (hte *HashTableEntry) empty() bool {
	return hte.Key == nil && hte.Value == nil
}

func (hte *HashTableEntry) keyEquals(obj Hashable) bool {
	if hte.empty() {
		return false
	}
	// Fast case: both objects are pointer-equivalent.
	// This can speed up cases where e.g. the user uses a
	// builtin (null, true, false).
	if hte.Key == obj {
		return true
	}
	return hte.Key.Type() == obj.Type() && hte.Key.HashEq(obj)
}

type HashTable struct {
	// We are using cuckoo hashing.
	// The tableSize is initially 16, and is doubled every
	// time we fail to allocate :'(. tableSize is _always_
	// a power of two. The size field is the _actual_ size
	// of the table.
	tableSize uint64
	size      uint64
	// seeds for the two hash functions
	seed1 uint64
	seed2 uint64
	// len(array) == tableSize
	array []HashTableEntry
}

func (ht *HashTable) Size() uint64 {
	return ht.size
}

func (ht *HashTable) h1(h uint64) uint64 {
	return (h ^ ht.seed1) & (ht.tableSize - 1)
}

func (ht *HashTable) h2(h uint64) uint64 {
	return (h ^ ht.seed2) & (ht.tableSize - 1)
}

// getEntryIndex returns the index of the HashTableEntry in
// the array that matches the given key, if any.
func (ht *HashTable) getEntryIndex(key Hashable) (uint64, bool) {
	h := key.Hash()
	h1 := ht.h1(h)
	if entry := ht.array[h1]; entry.keyEquals(key) {
		return h1, true
	}
	h2 := ht.h2(h)
	if entry := ht.array[h2]; entry.keyEquals(key) {
		return h2, true
	}
	return 0, false
}

func (ht *HashTable) Get(key Hashable) (Object, bool) {
	idx, ok := ht.getEntryIndex(key)
	if ok {
		return ht.array[idx].Value, true
	}
	return nil, false
}

func (ht *HashTable) Delete(key Hashable) bool {
	idx, ok := ht.getEntryIndex(key)
	if ok {
		ht.array[idx].Key = nil
		ht.array[idx].Value = nil
		ht.size -= 1
		if ht.tableSize > hashTableMinSize && float64(ht.size)/float64(ht.tableSize) < hashTableRehashRatio {
			ht.rehash(false)
		}
		return true
	}
	return false
}

func (ht *HashTable) Set(key Hashable, val Object) {
	if idx, ok := ht.getEntryIndex(key); ok {
		// If we're already in the hash table, then we are done.
		ht.array[idx].Value = val
		return
	}
	// Otherwise, try to insert the new pair.
	toInsert := HashTableEntry{key, val}
	maxTries := ht.tableSize / 2
	for tries := uint64(1); tries <= maxTries; tries++ {
		var tmp HashTableEntry
		var idx uint64
		// swap x and array[h1]
		idx = ht.h1(toInsert.Key.Hash())
		tmp = toInsert
		toInsert = ht.array[idx]
		ht.array[idx] = tmp
		if toInsert.empty() {
			ht.size++
			return
		}
		// swap x and array[h2]
		idx = ht.h2(toInsert.Key.Hash())
		tmp = toInsert
		toInsert = ht.array[idx]
		ht.array[idx] = tmp
		if toInsert.empty() {
			ht.size++
			return
		}
	}
	ht.rehash(true)
	ht.Set(toInsert.Key, toInsert.Value)
}

func (ht *HashTable) rehash(grow bool) {
	ht.size = 0
	ht.seed1 = rand.Uint64()
	ht.seed2 = rand.Uint64()
	if grow {
		ht.tableSize *= 2
	} else {
		ht.tableSize /= 2
	}
	oldEntries := ht.array
	ht.array = make([]HashTableEntry, ht.tableSize, ht.tableSize)
	for _, entry := range oldEntries {
		// In general, this may trigger another rehash(); in that case
		// we are safe since we first change the seeds and table size
		// before re-insertion; when we come back from _that_ rehash(),
		// the subsequent ht.Set(...) calls will use the new seeds and
		// sizes.
		if !entry.empty() {
			ht.Set(entry.Key, entry.Value)
		}
	}
}

func (ht *HashTable) Iter(iterator func(Hashable, Object) bool) {
	for _, entry := range ht.array {
		if !entry.empty() {
			if !iterator(entry.Key, entry.Value) {
				break
			}
		}
	}
}

func NewHashTable() *HashTable {
	ht := &HashTable{}
	ht.tableSize = hashTableMinSize / 2
	ht.rehash(true) // this sets seeds and the correct tableSize
	return ht
}
