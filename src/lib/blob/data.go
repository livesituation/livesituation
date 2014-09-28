package blob

import (
	"crypto/sha512"
	"encoding/base64"
	"encoding/json"
	"strconv"
	"time"

	"appengine/datastore"
)

type Primitive primitive

type primitive string

// NewPrimitive creates a new Primitive by
// normalizing the json string j.
func NewPrimitive(j string) (Primitive, error) {
	m := make(map[string]interface{})
	err := json.Unmarshal([]byte(j), &m)
	if err != nil {
		return Primitive(""), err
	}
	str, err := json.Marshal(m)
	return Primitive(str), err
}

type TimePrimitive struct {
	Timestamp time.Time
	Primitive
}

// NewTimePrimitive creates a new TimePrimitive
// by normalizing the json string j.
func NewTimePrimitive(timestamp time.Time, j string) (TimePrimitive, error) {
	var t TimePrimitive
	t.Timestamp = timestamp
	var err error
	t.Primitive, err = NewPrimitive(j)
	return t, err
}

func (t TimePrimitive) Hash() string {
	h := sha512.New()
	h.Write([]byte(strconv.FormatInt(t.Timestamp.UnixNano(), 10)))
	h.Write([]byte(t.Primitive))

	var buf [sha512.Size]byte
	h.Sum(buf[:0])

	str := base64.URLEncoding.EncodeToString(buf[:])
	return str
}

type History struct {
	Revisions []*datastore.Key
}

type Blob struct {
	History *datastore.Key
}
