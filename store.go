package archive

import (
	"log"
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
	RevisionRefs [][]byte
}

type Revision struct {
	Ref BlobRef
	Blob []byte
}

type BlobRef []byte

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

	return s
}

func (s *Store) getRevision (title string) *Revision {
	sub, _ = s.db.Get(buf, []byte(title))
	log.Printf("%#v", string(sub))

	var nn Note
	json.Unmarshal(sub, &nn)

	log.Printf("note nn: %#v", nn)

	sub, _ = s.db.Get(buf, nn.RevisionRefs[0])
	log.Printf("%#v", string(sub))

	r := &Revision{
		Ref: nn.RevisionRefs[0],
		Blob: sub,
	}

	return r
}

func (s *Store) addRevision () {
	data := []byte("Lorem ipsum dolor sit amet, consectetuer adipiscing elit.")
	h := sha1.New()
	h.Write(data)
	sha1sum := h.Sum(nil)

	r := Revision{
		Ref: sha1sum,
		Blob: data,
	}

	n := Note{
		RevisionRefs: make([][]byte, 0),
	}

	n.RevisionRefs = append(n.RevisionRefs, r.Ref)

	b, _  := json.Marshal(n)

	log.Printf("%#v", b)

	s.db.BeginTransaction()

	s.db.Set([]byte("aaaaa"), b)
	s.db.Set(r.Ref, r.Blob)

	s.db.Commit()

}
