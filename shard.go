package influxdb

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"
	"unsafe"

	"github.com/boltdb/bolt"
)

// Shard represents the physical storage for a given time range.
type Shard struct {
	ID        uint64    `json:"id,omitempty"`
	StartTime time.Time `json:"startTime,omitempty"`
	EndTime   time.Time `json:"endTime,omitempty"`

	store *bolt.DB
}

// newShard returns a new initialized Shard instance.
func newShard() *Shard { return &Shard{} }

// Duration returns the duration between the shard's start and end time.
func (s *Shard) Duration() time.Duration { return s.EndTime.Sub(s.StartTime) }

// open initializes and opens the shard's store.
func (s *Shard) open(path string) error {
	// Return an error if the shard is already open.
	if s.store != nil {
		return errors.New("shard already open")
	}

	// Open store on shard.
	store, err := bolt.Open(path, 0600, &bolt.Options{Timeout: 1 * time.Second})
	if err != nil {
		return err
	}
	s.store = store

	// Initialize store.
	if err := s.init(); err != nil {
		_ = s.close()
		return fmt.Errorf("init: %s", err)
	}

	return nil
}

// init creates top-level buckets in the datastore.
func (s *Shard) init() error {
	return s.store.Update(func(tx *bolt.Tx) error {
		_, _ = tx.CreateBucketIfNotExists([]byte("values"))
		return nil
	})
}

// close shuts down the shard's store.
func (s *Shard) close() error {
	return s.store.Close()
}

// writeSeries writes series data to a shard.
func (s *Shard) writeSeries(overwrite bool, data []byte) error {
	id, timestamp, values, err := unmarshalPoint(data)
	if err != nil {
		return err
	}

	// TODO: make this work
	fmt.Println("writeSeries: ", id, timestamp, values)
	return s.store.Update(func(tx *bolt.Tx) error {
		return nil // TODO
	})
}

func (s *Shard) deleteSeries(name string) error {
	panic("not yet implemented") // TODO
}

// Shards represents a list of shards.
type Shards []*Shard

// IDs returns a slice of all shard ids.
func (p Shards) IDs() []uint64 {
	ids := make([]uint64, len(p))
	for i, s := range p {
		ids[i] = s.ID
	}
	return ids
}

func marshalPoint(seriesID uint32, timestamp time.Time, values map[string]interface{}) ([]byte, error) {
	b := make([]byte, 12)
	*(*uint32)(unsafe.Pointer(&b[0])) = seriesID
	*(*int64)(unsafe.Pointer(&b[4])) = timestamp.UnixNano()

	d, err := json.Marshal(values)
	if err != nil {
		return nil, err
	}
	return append(b, d...), err
}

func unmarshalPoint(data []byte) (uint32, time.Time, map[string]interface{}, error) {
	id := *(*uint32)(unsafe.Pointer(&data[0]))
	ts := *(*int64)(unsafe.Pointer(&data[4]))
	timestamp := time.Unix(0, ts)
	var v map[string]interface{}

	err := json.Unmarshal(data[12:], &v)
	return id, timestamp, v, err
}
