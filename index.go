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
	titleIS *ferret.InvertedSuffix
	bodyIS *ferret.InvertedSuffix

	querychan chan Query
	rebuildchan chan Bundle
}

type Query struct {
	term string
	returnchan chan []string
}

type Bundle struct {
	titles, hexes, bodies []string
}

func newIndex () *Index {
	i := &Index{
		querychan: make(chan Query),
		rebuildchan: make(chan Bundle),
	}

	go i.run()

	return i
}

func (i *Index) run() {
	for {
		select {
		case q := <-i.querychan:
			q.returnchan <- i.query(q.term)
		case b := <-i.rebuildchan:
			i.rebuildWith(b.titles, b.hexes, b.bodies)
		}
	}
}


func (i *Index) RebuildWith(titles, hexes, bodies []string) {
	b := Bundle{
		titles: titles,
		hexes: hexes,
		bodies: bodies,
	}

	i.rebuildchan <- b
}

func (i *Index) rebuildWith(titles, hexes, bodies []string) {
	dummy := make([]interface{}, len(titles))
	t := time.Now()
        i.titleIS = ferret.New(titles, hexes, dummy, IndexConverter)
        i.bodyIS = ferret.New(bodies, hexes, dummy, IndexConverter)
	log.Print("Created index in: ", time.Now().Sub(t))
}

func (i *Index) Query(term string) []string {
	q := Query{
		term: term,
		returnchan: make(chan []string),
	}

	i.querychan <- q
	return <-q.returnchan
}

func (i *Index) query(term string) []string {
	t := time.Now()
	title_results, _ := i.titleIS.ErrorCorrectingQuery(term, 10, IndexCorrection)
	body_results, _ := i.bodyIS.ErrorCorrectingQuery(term, 10, IndexCorrection)

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
