package blob

import (
	"fmt"
	"time"

	"appengine"
	"appengine/datastore"
)

var (
	tpEntityType      = "TimePrimitive"
	historyEntityType = "History"
	blobEntityType    = "Blob"
)

func PutNewBlob(c appengine.Context, timestamp time.Time, j string) (ID string, err error) {
	tp, err := NewTimePrimitive(timestamp, j)
	if err != nil {
		return
	}

	key := datastore.NewKey(c, tpEntityType, tp.Hash(), 0, nil)
	_, err = datastore.Put(c, key, &tp)
	if err != nil {
		return
	}

	h := History{[]*datastore.Key{key}}
	key = datastore.NewIncompleteKey(c, historyEntityType, nil)
	key, err = datastore.Put(c, key, &h)
	if err != nil {
		// TODO(synful): attempt to delete primitive that's now orphaned
		return
	}

	b := Blob{key}
	key = datastore.NewIncompleteKey(c, blobEntityType, nil)
	key, err = datastore.Put(c, key, &b)
	if err != nil {
		// TODO(synful): attempt to delete primitive and history
		// that are now orphaned
		return
	}

	ID = key.Encode()
	return
}

func UpdateBlob(c appengine.Context, timestamp time.Time, ID, j string) error {
	tp, err := NewTimePrimitive(timestamp, j)
	if err != nil {
		return err
	}

	h, hkey, err := getHistory(c, ID)
	if err != nil {
		return err
	}

	key := datastore.NewKey(c, tpEntityType, tp.Hash(), 0, nil)
	_, err = datastore.Put(c, key, &tp)
	if err != nil {
		return err
	}

	h.Revisions = append(h.Revisions, key)
	_, err = datastore.Put(c, hkey, &h)
	if err != nil {
		return err
	}

	return nil
}

func GetCurrentBlob(c appengine.Context, ID string) (TimePrimitive, error) {
	var tp TimePrimitive
	h, _, err := getHistory(c, ID)
	if err != nil {
		return tp, err
	}

	err = datastore.Get(c, h.Revisions[len(h.Revisions)-1], &tp)
	return tp, err
}

func GetBlobRevision(c appengine.Context, ID string, revision int) (TimePrimitive, error) {
	h, _, err := getHistory(c, ID)
	if revision >= len(h.Revisions) {
		return TimePrimitive{}, fmt.Errorf("no such revision: %v", revision)
	}

	var tp TimePrimitive
	err = datastore.Get(c, h.Revisions[revision], &tp)
	return tp, err
}

func getHistory(c appengine.Context, ID string) (History, *datastore.Key, error) {
	var b Blob

	key, err := datastore.DecodeKey(ID)
	if err != nil {
		return History{}, key, err
	}
	err = datastore.Get(c, key, &b)
	if err != nil {
		return History{}, nil, err
	}

	var h History
	err = datastore.Get(c, b.History, &h)
	return h, b.History, err
}
