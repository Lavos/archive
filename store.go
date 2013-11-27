package archive

import (
	"log"
	"fmt"
	"errors"
	"math/rand"
	"time"
	"crypto/sha1"
	"bytes"
	"encoding/json"

	"github.com/cznic/kv"
)

var (
	buf = make([]byte, 20)
	sub = make([]byte, 0)
)

type Store struct {
	db *kv.DB
}

type Note struct {
	Title string
	Signature int64
	RevisionRefs []string
}

func newStore () *Store {
	db, err := kv.Open("db.kv", &kv.Options{})

	if err != nil {
		log.Printf("Error opening database: %#v", err)
		log.Print("Trying to create database.")
		db, err = kv.Create("db.kv", &kv.Options{})

		if err != nil {
			log.Fatalf("Could not create database: %#v", err)
		} else {
			log.Print("Database created successfully.")
		}
	}

	s := &Store{
		db: db,
	}

	titles, hexes := s.dump()

	for index, title := range titles {
		log.Printf("%v: %v", title, hexes[index])
	}

	return s
}

func (s *Store) dump () ([]string, []string) {
	t := time.Now()

	titles := make([]string, 0)
	hexes := make([]string, 0)

	enum, err := s.db.SeekFirst()

	if err == nil {
		key := make([]byte, 0)
		value := make([]byte, 0)
		var n Note
		var loop_err error

		for ; loop_err == nil; key, value, loop_err = enum.Next() {
			if bytes.HasPrefix(value, []byte(`{"Title":"`)) {
				json.Unmarshal(value, &n)

				titles = append(titles, n.Title)
				hexes = append(hexes, fmt.Sprintf("%x", key))
			}
		}
	}

	log.Print("Dumped store values index in: ", time.Now().Sub(t))

	return titles, hexes
}

func (s *Store) getBlob (b []byte) ([]byte, error) {
	sub, _ = s.db.Get(buf, b)
	log.Printf("%#v", string(sub))

	if sub == nil {
		return nil, errors.New("Could not find not with specified sha1sum")
	}

	return sub, nil
}

func (s *Store) addNote (title string) string {
	rand.Seed(time.Now().UTC().UnixNano())

	n := &Note{
		Title: title,
		Signature: rand.Int63(),
		RevisionRefs: make([]string, 0),
	}

	b, _  := json.Marshal(n)

	h := sha1.New()
	h.Write(b)
	sha1sum := h.Sum(nil)

	log.Printf("new note sha1: %x", sha1sum)

	s.db.Set(sha1sum, b)
	return fmt.Sprintf("%x", sha1sum)
}

func (s *Store) addRevision (targetRef []byte, content []byte) string {
	notebytes, _ := s.getBlob(targetRef)
	var n Note
	json.Unmarshal(notebytes, &n)

	h := sha1.New()
	h.Write(content)
	sha1sum := h.Sum(nil)
	hex := fmt.Sprintf("%x", sha1sum)

	log.Printf("add revision hex: %#v", hex)

	s.db.Set(sha1sum, content)

	n.RevisionRefs = append(n.RevisionRefs, hex)

	b, _ := json.Marshal(n)

	s.db.Set(targetRef, b)
	return hex
}
