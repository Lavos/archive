package archive

import (
	"fmt"
	"log"
	"os"
	"github.com/hoisie/web"
)

type Server struct {
	server *web.Server
	store *Store
}

func NewServer() *Server {
	w := web.NewServer()
	s := &Server {
		server: w,
	}

	w.Get("/note/(.*)", s.noteHandler)

	return s
}

func (s *Server) (val string) string {
	return fmt.Sprintf("hello! %#v - %#v", val, s.store.getRevision(val))
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
