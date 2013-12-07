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
	t := time.Now()
        i.TitleIS = ferret.New(titles, hexes, dummy, IndexConverter)
        i.BodyIS = ferret.New(bodies, hexes, dummy, IndexConverter)
	log.Print("Created index in: ", time.Now().Sub(t))
}

func (i *Index) Insert(title, hex string) {
	t := time.Now()
        i.TitleIS.Insert(title, hex, nil)
	log.Print("Insert into title index in: ", time.Now().Sub(t))
}

func (i *Index) Query(term string) []string {
	t := time.Now()
	title_results, _ := i.TitleIS.ErrorCorrectingQuery(term, 10, IndexCorrection)
	body_results, _ := i.BodyIS.ErrorCorrectingQuery(term, 10, IndexCorrection)

	log.Print("Query completed in: ", time.Now().Sub(t))

	log.Printf("%#v - %v", title_results, len(title_results))
	log.Printf("%#v - %v" , body_results, len(body_results))

	t = time.Now()

	merged_results := append(title_results, body_results...)
	unique_map := make(map[string]bool)
	unique_results := make([]string, 0)

	for _, hex := range merged_results {
		_, ok := unique_map[hex]

		if !ok {
			unique_map[hex] = true
			unique_results = append(unique_results, hex)
		}
	}

	log.Print("Unique Results found in: ", time.Now().Sub(t))
	log.Print("Unique Results: ", len(unique_results))

	return unique_results
}
