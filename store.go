package archive

import (
	"log"
	// "fmt"
	"errors"
	"math/rand"
	"time"
	"github.com/cznic/kv"
	"crypto/sha1"
	"encoding/json"
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
	signature int64
	RevisionRefs [][]byte
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

	enum, err := db.SeekFirst()

	if err == nil {
		key := make([]byte, 0)
		value := make([]byte, 0)
		err = nil

		for ; err == nil; key, value, err = enum.Next() {
			log.Printf("%#v: %v", key, string(value))
		}
	}

	return s
}

func (s *Store) getNote (b []byte) ([]byte, error) {
	sub, _ = s.db.Get(buf, b)
	log.Printf("%#v", string(sub))

	if sub == nil {
		return nil, errors.New("Could not find not with specified sha1sum")
	}

	return sub, nil
}

func (s *Store) getRevision (b []byte) ([]byte, error) {
	log.Printf("%#v", b)

	sub, _ = s.db.Get(buf, b)
	log.Printf("%#v", string(sub))

	if sub == nil {
		return nil, errors.New("Could not get revision with specified blobref: " + string(b))
	}

	return sub, nil
}

func (s *Store) addNote (title string) {
	rand.Seed(time.Now().UTC().UnixNano())

	n := &Note{
		Title: title,
		signature: rand.Int63(),
		RevisionRefs: make([][]byte, 0),
	}

	b, _  := json.Marshal(n)

	log.Printf("new note json: %#v", string(b))

	h := sha1.New()
	h.Write(b)
	sha1sum := h.Sum(nil)

	log.Printf("new note sha1: %x", sha1sum)
	log.Printf("new note sha1: %#v", sha1sum)

	s.db.Set(sha1sum, b)
	log.Printf(string(b))
}

func (s *Store) addRevision (targetRef []byte, content []byte) {
	notebytes := s.getNote(target)
	var n note
	json.Unmarshal(notebytes, &n)

	h := sha1.New()
	h.write(content)
	sha1sum := h.Sum(nil)

	s.db.Set(sha1sum, content)

	n.RevisionRefs = append(n.RevisionRefs, sha1sum)

	b, _ := json.Marshal(n)

	s.db.Set(targetRef, b)
}
