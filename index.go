package archive

import (
	"time"
	"log"
	"github.com/argusdusty/Ferret"
)

var (
	IndexCorrection = func(b []byte) [][]byte { return ferret.ErrorCorrect(b, ferret.LowercaseLetters) }
	IndexSorter = func(s string, v interface{}, l int, i int) float64 { return -float64(l + i) }
	IndexConverter = func(s string) []byte { return []byte(s) }
)

type Index struct {
	TitleIS *ferret.InvertedSuffix
	BodyIS *ferret.InvertedSuffix
}

func newIndex () *Index {
	return &Index{}
}

func (i *Index) RebuildWith(titles, hexes, bodies []string) {
	dummy := make([]interface{}, len(titles))
	t := time.Now();
        i.TitleIS = ferret.New(titles, hexes, dummy, IndexConverter)
        i.BodyIS = ferret.New(bodies, hexes, dummy, IndexConverter)
	log.Print("Created index in: ", time.Now().Sub(t))
}

func (i *Index) Query(term string) []string {
	t := time.Now();
	title_results, _ := i.TitleIS.ErrorCorrectingQuery(term, 10, IndexCorrection)
	body_results, _ := i.BodyIS.ErrorCorrectingQuery(term, 10, IndexCorrection)
	log.Print("Query completed in: ", time.Now().Sub(t))

	log.Printf("%#v", title_results)
	log.Printf("%#v", body_results)
	return append(title_results, body_results...)
}
