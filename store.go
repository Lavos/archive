package archive

import (
	"log"
	// "fmt"
	"errors"
	"math/rand"
	"time"
	"crypto/sha1"
	"bytes"
	"encoding/hex"
	"encoding/json"

	"github.com/cznic/kv"
)

var (
	buf = make([]byte, 20)
	sub = make([]byte, 0)
)

type Store struct {
	db *kv.DB
	index *Index
}

type Note struct {
	Hex string `json:"hex"`
	Title string `json:"title"`
	Signature int64 `json:"signature"`
	RevisionRefs []string `json:"revision_refs"`
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
		index: newIndex(),
	}

	go s.reindex()

	return s
}

func (s *Store) reindex () {
	titles, hexes, bodies := s.dump()

	s.index.RebuildWith(titles, hexes, bodies)

	/* for index, title := range titles {
		log.Printf("%v: %v", title, hexes[index])
	} */
}

func (s *Store) dump () ([]string, []string, []string) {
	t := time.Now()

	titles := make([]string, 0)
	hexes := make([]string, 0)
	bodies := make([]string, 0)

	enum, err := s.db.SeekFirst()

	if err == nil {
		key := make([]byte, 0)
		value := make([]byte, 0)
		latest_revision_bytes := make([]byte, 0)
		sha1sum := make([]byte, 0)
		var body string

		var n Note
		var loop_err error

		for ; loop_err == nil; key, value, loop_err = enum.Next() {
			if bytes.HasPrefix(value, []byte(`{"hex":"`)) {
				json.Unmarshal(value, &n)

				titles = append(titles, n.Title)
				hexes = append(hexes, hex.EncodeToString(key))

				if len(n.RevisionRefs) > 0 {
					sha1sum, _ = hex.DecodeString(n.RevisionRefs[len(n.RevisionRefs)-1])
					latest_revision_bytes, _ = s.getBlob(sha1sum)
					body = string(latest_revision_bytes)
				} else {
					body = ""
				}

				bodies = append(bodies, body)
			}
		}
	}

	log.Print("Dumped store values in: ", time.Now().Sub(t))

	return titles, hexes, bodies
}

func (s *Store) query (term string) []Note {
	hexes := s.index.Query(term)
	notes := make([]Note, 0)

	var sha1sum []byte
	var notebytes []byte

	for _, hexstr := range hexes {
		sha1sum, _ = hex.DecodeString(hexstr)
		notebytes, _ = s.getBlob(sha1sum)

		var n Note
		json.Unmarshal(notebytes, &n)

		notes = append(notes, n)
	}

	return notes
}

func (s *Store) getBlob (b []byte) ([]byte, error) {
	sub, _ = s.db.Get(buf, b)

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

	b_sign, _  := json.Marshal(n)

	h := sha1.New()
	h.Write(b_sign)
	sha1sum := h.Sum(nil)

	n.Hex = hex.EncodeToString(sha1sum)

	b_val, _ := json.Marshal(n)

	s.db.Set(sha1sum, b_val)

	go s.reindex()

	return n.Hex
}

func (s *Store) patchNote (targetRef []byte, title string) string {
	notebytes, _ := s.getBlob(targetRef)
	var n Note
	json.Unmarshal(notebytes, &n)

	n.Title = title;

	log.Printf("updated note: %#v", n)

	b_val, _ := json.Marshal(n)

	s.db.Set(targetRef, b_val)
	go s.reindex()

	return n.Hex
}

func (s *Store) addRevision (targetRef []byte, content []byte) string {
	notebytes, _ := s.getBlob(targetRef)
	var n Note
	json.Unmarshal(notebytes, &n)

	h := sha1.New()
	h.Write(content)
	sha1sum := h.Sum(nil)
	hexstr := hex.EncodeToString(sha1sum)

	s.db.Set(sha1sum, content)

	n.RevisionRefs = append(n.RevisionRefs, hexstr)

	b, _ := json.Marshal(n)

	s.db.Set(targetRef, b)
	go s.reindex()

	return hexstr
}
