package sophia

import (
	"fmt"
	"math"
	"math/rand"
	"os"
	"testing"
	"unsafe"

	"github.com/stretchr/testify/require"
)

// TODO write tests:
//   - using upsert
//     - success
//     - error
//   - test more settings for environment
//   - test more settings for database

const (
	KeyTemplate   = "key%v"
	ValueTemplate = "value%v"

	DBPath       = "sophia"
	DBName       = "test"
	RecordsCount = 500000

	RecordsCountBench = 5000000
)

func TestDatabaseCRUD(t *testing.T) {
	defer func() {
		require.Nil(t, os.RemoveAll(DBPath))
	}()
	var env *Environment
	var db *Database

	if !t.Run("New Environment", func(t *testing.T) { env = testNewEnvironment(t) }) {
		t.Fatal("Failed to create environment object")
	}
	defer func() { require.Nil(t, env.Close()) }()

	if !t.Run("New Database", func(t *testing.T) { db = testNewDatabase(t, env) }) {
		t.Fatalf("Failed to create database object: %v", env.Error())
	}

	if !t.Run("Set", func(t *testing.T) { testSet(t, db) }) {
		t.Fatalf("Set operations are failed: %v", env.Error())
	}
	if !t.Run("Get", func(t *testing.T) { testGet(t, db) }) {
		t.Fatalf("Get operations are failed: %v", env.Error())
	}
	if !t.Run("Detele", func(t *testing.T) { testDelete(t, db) }) {
		t.Fatalf("FDelete operations are failed: %v", env.Error())
	}
}

func testNewEnvironment(t *testing.T) *Environment {
	env, err := NewEnvironment()
	require.Nil(t, err)
	require.NotNil(t, env)
	return env
}

func testNewDatabase(t *testing.T, env *Environment) *Database {
	require.True(t, env.Set("sophia.path", DBPath))

	schema := &Schema{}
	require.Nil(t, schema.AddKey("key", FieldTypeString))
	require.Nil(t, schema.AddValue("value", FieldTypeString))

	db, err := env.NewDatabase(&DatabaseConfig{
		Name:   DBName,
		Schema: schema,
	})
	require.Nil(t, err)
	require.NotNil(t, db)
	require.Nil(t, env.Open())
	return db
}

func testSet(t *testing.T, db *Database) {
	for i := 0; i < RecordsCount; i++ {
		doc := db.Document()
		require.True(t, doc.Set("key", fmt.Sprintf(KeyTemplate, i)))
		require.True(t, doc.Set("value", fmt.Sprintf(ValueTemplate, i)))

		require.Nil(t, db.Set(doc))
		doc.Free()
	}
}

func testGet(t *testing.T, db *Database) {
	for i := 0; i < RecordsCount; i++ {
		doc := db.Document()
		require.NotNil(t, doc)
		require.True(t, doc.Set("key", fmt.Sprintf(KeyTemplate, i)))
		d, err := db.Get(doc)
		doc.Free()
		require.NotNil(t, d)
		require.Nil(t, err)
		var size int
		require.Equal(t, fmt.Sprintf(KeyTemplate, i), d.GetString("key", &size))
		require.Equal(t, fmt.Sprintf(ValueTemplate, i), d.GetString("value", &size))
		d.Destroy()
	}
}

func testDelete(t *testing.T, db *Database) {
	for i := 0; i < RecordsCount; i++ {
		doc := db.Document()
		require.NotNil(t, doc)
		require.True(t, doc.Set("key", fmt.Sprintf(KeyTemplate, i)))
		require.Nil(t, db.Delete(doc))
		doc.Free()
	}

	for i := 0; i < RecordsCount; i++ {
		doc := db.Document()
		require.NotNil(t, doc)
		require.True(t, doc.Set("key", fmt.Sprintf(KeyTemplate, i)))
		d, err := db.Get(doc)
		doc.Free()
		require.Nil(t, d)
		require.NotNil(t, err)
	}
}

func TestSchemaDupKey(t *testing.T) {
	schema := Schema{}
	keyName := "key"
	require.Nil(t, schema.AddKey(keyName, FieldTypeString))

	require.Len(t, schema.keys, 1)
	require.Len(t, schema.keysNames, 1)

	require.Len(t, schema.values, 0)
	require.Len(t, schema.valuesNames, 0)

	require.Equal(t, FieldTypeString, schema.keys[keyName])
	require.Equal(t, keyName, schema.keysNames[0])

	require.NotNil(t, schema.AddKey(keyName, FieldTypeString))

	require.Len(t, schema.keys, 1)
	require.Len(t, schema.keysNames, 1)

	require.Len(t, schema.values, 0)
	require.Len(t, schema.valuesNames, 0)

	require.Equal(t, FieldTypeString, schema.keys[keyName])
	require.Equal(t, keyName, schema.keysNames[0])
}

func TestSchemaDupValue(t *testing.T) {
	schema := Schema{}
	valueName := "key"
	require.Nil(t, schema.AddValue(valueName, FieldTypeString))

	require.Len(t, schema.keys, 0)
	require.Len(t, schema.keysNames, 0)

	require.Len(t, schema.values, 1)
	require.Len(t, schema.valuesNames, 1)

	require.Equal(t, FieldTypeString, schema.values[valueName])
	require.Equal(t, valueName, schema.valuesNames[0])

	require.NotNil(t, schema.AddValue(valueName, FieldTypeString))

	require.Len(t, schema.keys, 0)
	require.Len(t, schema.keysNames, 0)

	require.Len(t, schema.values, 1)
	require.Len(t, schema.valuesNames, 1)

	require.Equal(t, FieldTypeString, schema.values[valueName])
	require.Equal(t, valueName, schema.valuesNames[0])
}

func TestSetIntKV(t *testing.T) {
	defer func() {
		require.Nil(t, os.RemoveAll(DBPath))
	}()
	env, err := NewEnvironment()
	require.Nil(t, err)
	require.NotNil(t, env)
	defer func() {
		require.Nil(t, env.Close())
	}()

	require.True(t, env.Set("sophia.path", DBPath))

	schema := &Schema{}
	require.Nil(t, schema.AddKey("key", FieldTypeUInt32))
	require.Nil(t, schema.AddValue("value", FieldTypeUInt32))

	db, err := env.NewDatabase(&DatabaseConfig{
		Name:   DBName,
		Schema: schema,
	})
	require.Nil(t, err)
	require.NotNil(t, db)
	require.Nil(t, env.Open())

	for i := 0; i < RecordsCount; i++ {
		doc := db.Document()
		require.NotNil(t, doc)
		require.True(t, doc.Set("key", int64(i)))
		require.True(t, doc.Set("value", int64(i)))

		require.Nil(t, db.Set(doc))
		doc.Free()
	}
	for i := 0; i < RecordsCount; i++ {
		doc := db.Document()
		require.NotNil(t, doc)
		require.True(t, doc.Set("key", int64(i)))
		d, err := db.Get(doc)
		doc.Free()
		require.Nil(t, err)
		require.NotNil(t, d)
		require.Equal(t, int64(i), d.GetInt("key"))
		require.Equal(t, int64(i), d.GetInt("value"))
		d.Destroy()
		d.Free()
	}
}

func TestSetMultiKey(t *testing.T) {
	defer func() { require.Nil(t, os.RemoveAll(DBPath)) }()
	env, err := NewEnvironment()
	require.Nil(t, err)
	require.NotNil(t, env)
	defer func() { require.Nil(t, env.Close()) }()

	require.True(t, env.Set("sophia.path", DBPath))

	schema := &Schema{}
	require.Nil(t, schema.AddKey("key", FieldTypeUInt32))
	require.Nil(t, schema.AddKey("key_j", FieldTypeUInt32))
	require.Nil(t, schema.AddKey("key_k", FieldTypeUInt32))
	require.Nil(t, schema.AddValue("value", FieldTypeUInt64))

	db, err := env.NewDatabase(&DatabaseConfig{
		Name:   DBName,
		Schema: schema,
	})
	require.Nil(t, err)
	require.NotNil(t, db)
	require.Nil(t, env.Open())

	count := int(math.Pow(RecordsCount, 1/3))

	for i := 0; i < count; i++ {
		for j := 0; j < count; j++ {
			for k := 0; k < count; k++ {
				doc := db.Document()
				require.True(t, doc.Set("key", i))
				require.True(t, doc.Set("key_j", uint64(j)))
				require.True(t, doc.Set("key_k", uint32(k)))
				require.True(t, doc.Set("value", i))

				require.Nil(t, db.Set(doc))
				doc.Free()
			}
		}
	}
	for i := 0; i < count; i++ {
		for j := 0; j < count; j++ {
			for k := 0; k < count; k++ {
				doc := db.Document()
				require.NotNil(t, doc)
				require.True(t, doc.Set("key", int64(i)))
				require.True(t, doc.Set("key_j", int64(j)))
				require.True(t, doc.Set("key_k", int64(k)))
				d, err := db.Get(doc)
				doc.Free()
				require.Nil(t, err)
				require.NotNil(t, d)
				require.Equal(t, int64(i), d.GetInt("key"))
				require.Equal(t, int64(j), d.GetInt("key_j"))
				require.Equal(t, int64(k), d.GetInt("key_k"))
				require.Equal(t, int64(i), d.GetInt("value"))
				d.Destroy()
				d.Free()
			}
		}
	}
}

func TestDatabaseUseSomeDocumentsAtTheSameTime(t *testing.T) {
	defer func() {
		require.Nil(t, os.RemoveAll(DBPath))
	}()
	env, err := NewEnvironment()
	require.Nil(t, err)
	require.NotNil(t, env)

	require.True(t, env.Set("sophia.path", DBPath))

	schema := &Schema{}
	require.Nil(t, schema.AddKey("key", FieldTypeString))
	require.Nil(t, schema.AddValue("value", FieldTypeString))

	db, err := env.NewDatabase(&DatabaseConfig{
		Name:   DBName,
		Schema: schema,
	})
	require.Nil(t, err)
	require.NotNil(t, db)
	require.Nil(t, env.Open())

	doc1 := db.Document()
	doc2 := db.Document()

	require.NotNil(t, doc1)
	require.NotNil(t, doc2)

	defer func() {
		doc1.Free()
		doc2.Free()
	}()

	require.True(t, doc1.Set("key", "key1"))
	require.True(t, doc1.Set("value", "value1"))
	require.True(t, doc2.Set("key", "key2"))
	require.True(t, doc2.Set("value", "value2"))

	db.Set(doc1)
	db.Set(doc2)

	doc := db.Document()
	require.NotNil(t, doc)

	doc.Set("key", "key1")
	d, err := db.Get(doc)
	require.NotNil(t, d)
	require.Nil(t, err)
	size := 0
	require.Equal(t, "value1", d.GetString("value", &size))
	require.Equal(t, 6, size)
	d.Destroy()

	doc = db.Document()
	require.NotNil(t, doc)

	doc.Set("key", "key2")
	d, err = db.Get(doc)
	require.NotNil(t, d)
	require.Nil(t, err)
	size = 0
	require.Equal(t, "value2", d.GetString("value", &size))
	require.Equal(t, 6, size)
	d.Destroy()
}

func TestDatabaseDeleteNotExistingKey(t *testing.T) {
	defer func() {
		require.Nil(t, os.RemoveAll(DBPath))
	}()
	env, err := NewEnvironment()
	require.Nil(t, err)
	require.NotNil(t, env)

	require.True(t, env.Set("sophia.path", DBPath))

	schema := &Schema{}
	require.Nil(t, schema.AddKey("key", FieldTypeString))
	require.Nil(t, schema.AddValue("value", FieldTypeString))

	db, err := env.NewDatabase(&DatabaseConfig{
		Name:   DBName,
		Schema: schema,
	})
	require.Nil(t, err)
	require.NotNil(t, db)
	require.Nil(t, env.Open())

	doc := db.Document()
	defer doc.Free()
	doc.Set("key", "key1")
	require.Nil(t, db.Delete(doc))
}

func TestDatabaseUpsert(t *testing.T) {
	defer func() {
		require.Nil(t, os.RemoveAll(DBPath))
	}()
	env, err := NewEnvironment()
	require.Nil(t, err)
	require.NotNil(t, env)

	require.True(t, env.Set("sophia.path", DBPath))

	schema := &Schema{}
	require.Nil(t, schema.AddKey("key", FieldTypeUInt32))
	require.Nil(t, schema.AddValue("id", FieldTypeUInt32))

	db, err := env.NewDatabase(&DatabaseConfig{
		Name:   DBName,
		Schema: schema,
		Upsert: upsertCallback,
	})
	require.Nil(t, err)
	require.NotNil(t, db)


	require.Nil(t, env.Open())

	/* increment key 10 times */
	const key uint32 = 1234
	const iterations = 10
	var increment int64 = 1
	for i := 0; i < iterations; i++ {
		doc := db.Document()
		doc.Set("key", key)
		doc.Set("id", increment)
		require.Nil(t, db.Upsert(doc))
	}

	/* get */
	doc := db.Document()
	doc.Set("key", key)

	result, err := db.Get(doc)
	require.Nil(t, err)
	require.NotNil(t, result)
	defer result.Destroy()

	require.Equal(t, iterations*increment, result.GetInt("id"))
}

func upsertCallback(count int,
	src []unsafe.Pointer, srcSize uint32,
	upsert []unsafe.Pointer, upsertSize uint32,
	result []unsafe.Pointer, resultSize uint32,
	arg unsafe.Pointer) int {
	var a uint32 = *(*uint32)(src[1])
	var b uint32 = *(*uint32)(upsert[1])
	ret := a + b
	resPtr := (*uint32)(result[1])
	*resPtr = ret
	return 0
}

// ATTN - This benchmark don't show real performance
// It is just a long running tests
func BenchmarkDatabaseSet(b *testing.B) {
	defer func() { require.Nil(b, os.RemoveAll(DBPath)) }()
	env, err := NewEnvironment()
	require.Nil(b, err)
	require.NotNil(b, env)

	require.True(b, env.Set("sophia.path", DBPath))

	schema := &Schema{}
	require.Nil(b, schema.AddKey("key", FieldTypeString))
	require.Nil(b, schema.AddValue("value", FieldTypeString))

	db, err := env.NewDatabase(&DatabaseConfig{
		Name:   DBName,
		Schema: schema,
	})
	require.Nil(b, err)
	require.NotNil(b, db)
	require.Nil(b, env.Open())

	indices := rand.Perm(b.N)
	keys := make(map[string]string)
	for _, i := range indices {
		keys[fmt.Sprintf(KeyTemplate, i)] = fmt.Sprintf(ValueTemplate, i)
	}
	b.ResetTimer()
	for key, value := range keys {
		doc := db.Document()
		require.True(b, doc.Set("key", key))
		require.True(b, doc.Set("value", value))
		require.Nil(b, db.Set(doc))
		doc.Free()
	}
}

// ATTN - This benchmark don't show real performance
// It is just a long running tests
func BenchmarkDatabaseGet(b *testing.B) {
	defer func() { require.Nil(b, os.RemoveAll(DBPath)) }()
	env, err := NewEnvironment()
	require.Nil(b, err)
	require.NotNil(b, env)

	require.True(b, env.Set("sophia.path", DBPath))

	schema := &Schema{}
	require.Nil(b, schema.AddKey("key", FieldTypeString))
	require.Nil(b, schema.AddValue("value", FieldTypeString))

	db, err := env.NewDatabase(&DatabaseConfig{
		Name:   DBName,
		Schema: schema,
	})
	require.Nil(b, err)
	require.NotNil(b, db)
	require.Nil(b, env.Open())

	for i := 0; i < RecordsCountBench; i++ {
		doc := db.Document()
		require.True(b, doc.Set("key", fmt.Sprintf(KeyTemplate, i)))
		require.True(b, doc.Set("value", fmt.Sprintf(ValueTemplate, i)))
		err = db.Set(doc)
		require.Nil(b, err)
		doc.Free()
	}

	indices := rand.Perm(b.N)
	keys := make(map[string]string)
	for _, i := range indices {
		keys[fmt.Sprintf(KeyTemplate, i)] = fmt.Sprintf(ValueTemplate, i)
	}
	var size int
	b.ResetTimer()
	for key, value := range keys {
		doc := db.Document()
		require.True(b, doc.Set("key", key))
		d, err := db.Get(doc)
		require.Nil(b, err)
		require.Equal(b, value, d.GetString("value", &size))
		doc.Free()
		d.Free()
		d.Destroy()
	}
}
