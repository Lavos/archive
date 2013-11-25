package main

import (
	"log"
	"github.com/cznic/kv"
	"crypto/sha1"
	"encoding/json"
)

type Note struct {
	Revisions [][]byte
}

func main () {
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

	defer db.Close()

	data := []byte("Lorem ipsum dolor sit amet, consectetuer adipiscing elit. Aenean commodo ligula eget dolor. Aenean massa. Cum sociis natoque penatibus et magnis dis parturient montes, nascetur ridiculus mus. Donec quam felis, ultricies nec, pellentesque eu, pretium quis, sem. Nulla consequat massa quis enim. Donec pede justo, fringilla vel, aliquet nec, vulputate eget, arcu. In enim justo, rhoncus ut, imperdiet a, venenatis vitae, justo. Nullam dictum felis eu pede mollis pretium. Integer tincidunt. Cras dapibus. Vivamus elementum semper nisi. Aenean vulputate eleifend tellus. Aenean leo ligula, porttitor eu, consequat vitae, eleifend ac, enim. Aliquam lorem ante, dapibus in, viverra quis, feugiat a, tellus. Phasellus viverra nulla ut metus varius laoreet. Quisque rutrum. Aenean imperdiet. Etiam ultricies nisi vel augue. Curabitur ullamcorper ultricies nisi. Nam eget dui. Etiam rhoncus.")
	h := sha1.New()
	h.Write(data)
	sha1sum := h.Sum(nil)

	log.Printf("sha1: %x", sha1sum)

	n := Note{
		Revisions: make([][]byte, 0),
	}

	n.Revisions = append(n.Revisions, sha1sum)

	b, _  := json.Marshal(n)

	db.Set([]byte("First Testing Title"), b)
	db.Set(sha1sum, data)

	db.Commit()

	buf := make([]byte, 20)
	sub := make([]byte, 0)

	sub, _ = db.Get(buf, []byte("First Testing Title"))
	log.Printf("%#v", string(sub))

	var nn Note
	json.Unmarshal(sub, &nn)

	log.Printf("note nn: %#v", nn)

	sub, _ = db.Get(buf, nn.Revisions[0])
	log.Printf("%#v", string(sub))
}
