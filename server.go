package archive

import (
	// "fmt"
	"log"
	"os"
	"github.com/hoisie/web"
	"encoding/json"
	"strconv"
	"errors"
)

type PostJSON struct {
	Title string `json:"title"`
}

type Server struct {
	server *web.Server
	store *Store
}

func NewServer() *Server {
	w := web.NewServer()
	s := &Server {
		server: w,
	}

	w.Get("/note/([0-9a-f]{40})/?", s.getNoteRevisions)
	w.Post("/note/", s.postNote)
	w.Post("/note/([0-9a-f]{40})", s.postRevision)

	return s
}

func getSumFromString (hex string) ([]byte, error) {
	sha1sum := make([]byte, 0)
	var pair string
	var val uint64
	var err error

	for len(hex) > 0 {
		pair = hex[:2]
		val, err = strconv.ParseUint(pair, 16, 8)

		if err != nil {
			return nil, errors.New("Could not parse int.")
		}

		sha1sum = append(sha1sum, uint8(val))
		hex = hex[2:]
	}

	return sha1sum, nil
}

func (s *Server) getNoteRevisions (ctx *web.Context, hex string) string {
	log.Printf("string: %v", hex);

	sha1sum, err := getSumFromString(hex)

	if err != nil {
		ctx.Abort(500, "Could not get sha1sum from provided string.")
		return ""
	}

	notebytes, err := s.store.getNote(sha1sum)

	if err != nil {
		ctx.Abort(500, "Could not get note with that sha1.")
		return ""
	}

	return string(notebytes)
}

func (s *Server) postNote (ctx *web.Context) string {
	var postjson PostJSON
	json.Unmarshal([]byte(ctx.Params["json"]), &postjson)

	log.Printf("params: %#v", ctx.Params)
	log.Printf("postjson obj: %#v", postjson)

	s.store.addNote(postjson.Title)
	return "done"
}

func (s *Server) postRevision (val string) string {
	
}

func (s *Server) Run () {
	log.Print("started server")

	s.store = newStore()
	go s.server.Run(":8000")
	awaitQuitKey()
	err := s.store.db.Close()
	log.Printf("%#v", err)
}

func awaitQuitKey() {
	var buf [1]byte
	for {
		_, err := os.Stdin.Read(buf[:])
		if err != nil || buf[0] == 'q' {
			return
		}
	}
}
