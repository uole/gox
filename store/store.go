package store
import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"sync"
	"time"
)

type (
	item struct {
		Count     int
		CreatedAt time.Time
		Value     interface{}
	}

	Store struct {
		filename string
		mutex    sync.RWMutex
		Items    map[string]*item
	}
)

func (i *item) Resolve(v interface{}) error {
	if buf, err := json.Marshal(i.Value); err == nil {
		return json.Unmarshal(buf, v)
	} else {
		return err
	}
}

func (s *Store) Get(key string) (item, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	if v, ok := s.Items[key]; !ok {
		return item{}, errors.New(key + " not exists")
	} else {
		v.Count++
		return *v, nil
	}
}

func (s *Store) Set(key string, value interface{}) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.Items[key] = &item{
		Value:     value,
		CreatedAt: time.Now(),
	}
}

func (s *Store) Del(key string) (item, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	if v, ok := s.Items[key]; !ok {
		return item{}, errors.New(key + " not exists")
	} else {
		delete(s.Items, key)
		return *v, nil
	}
}

func (s *Store) xor(buf []byte) []byte {
	length := len(buf)
	blockSize := length%8 + 1
	for i := blockSize - 1; i < length; i += blockSize {
		for j := 0; j < blockSize-1; j++ {
			buf[i-j] = buf[i-j] ^ buf[i-(blockSize-1)]
		}
	}
	return buf
}

func (s *Store) Load() error {
	if buf, err := ioutil.ReadFile(s.filename); err == nil {
		s.xor(buf)
		return json.Unmarshal(buf, &s.Items)
	} else {
		return err
	}
}

func (s *Store) Flush() error {
	if buf, err := json.Marshal(s.Items); err == nil {
		s.xor(buf)
		return ioutil.WriteFile(s.filename, buf, 0644)
	} else {
		return err
	}
}

func NewStore(filename string) *Store {
	return &Store{
		Items:    make(map[string]*item),
		filename: filename,
	}
}
