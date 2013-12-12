package archive

import (
	// "fmt"
	"log"
	"os"
	"github.com/hoisie/web"
	"encoding/json"
	"strconv"
	"errors"
	"io/ioutil"
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

	w.Get("/api/([0-9a-f]{40})", s.getBlob)
	w.Get("/api/search", s.search)

	w.Post("/api/new", s.postNote)
	w.Post("/api/([0-9a-f]{40})", s.postRevision)

	w.Match("PATCH", "/api/([0-9a-f]{40})", s.patchNote)

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

func (s *Server) getBlob (ctx *web.Context, hex string) string {
	log.Printf("string: %v", hex);

	sha1sum, err := getSumFromString(hex)

	if err != nil {
		ctx.Abort(500, "Could not get sha1sum from provided string.")
		return ""
	}

	b, err := s.store.getBlob(sha1sum)

	if err != nil {
		ctx.Abort(500, "Could not get blob with that sha1.")
		return ""
	}

	return string(b)
}

func (s *Server) search (ctx *web.Context) []byte {
	notes := s.store.query(ctx.Params["q"])
	b, _ := json.Marshal(notes)

	return b
}

func (s *Server) postNote (ctx *web.Context) string {
	body, err := ioutil.ReadAll(ctx.Request.Body)

	if err != nil {
		ctx.Abort(500, "No body for new note.")
		return ""
	}

	var postjson PostJSON
	err = json.Unmarshal(body, &postjson)

	if err != nil {
		ctx.Abort(500, "Could not parse body as JSON.")
		return ""
	}

	if len(postjson.Title) == 0 {
		ctx.Abort(500, "No title found.")
		return ""
	}

	return s.store.addNote(postjson.Title)
}

func (s *Server) patchNote (ctx *web.Context, hex string) string {
	body, err := ioutil.ReadAll(ctx.Request.Body)

	if err != nil {
		ctx.Abort(500, "No body for nore patch.")
		return ""
	}

	var postjson PostJSON
	err = json.Unmarshal(body, &postjson)

	if err != nil {
		ctx.Abort(500, "Could not parse body as JSON.")
		return ""
	}

	if len(postjson.Title) == 0 {
		ctx.Abort(500, "No title found.")
		return ""
	}

	sha1sum, err := getSumFromString(hex)

	if err != nil {
		ctx.Abort(500, "Could not get sha1sum from provided hex.")
		return ""
	}

	return s.store.patchNote(sha1sum, postjson.Title)
}

func (s *Server) postRevision (ctx *web.Context, hex string) string {
	body, err := ioutil.ReadAll(ctx.Request.Body)

	if err != nil {
		ctx.Abort(500, "No body supplied for revision.")
		return ""
	}

	sha1sum, err := getSumFromString(hex)

	if err != nil {
		ctx.Abort(500, "Could not get sha1sum from provided hex.")
		return ""
	}

	return s.store.addRevision(sha1sum, body)
}

func (s *Server) Run () {
	log.Print("started server")

	s.store = newStore()
	go s.server.Run(":9000")
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
