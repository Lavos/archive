package archive

import (

	"github.com/argusdusty/Ferret"
)

var (
	IndexCorrection = func(b []byte) [][]byte { return ferret.ErrorCorrect(b, ferret.LowercaseLetters) }
	IndexSorter = func(s string, v interface{}, l int, i int) float64 { return -float64(l + i) }
	IndexConverter = func(s string) []byte { return []byte(s) }
)

type Index struct {
	TitleIS *ferret.InvertedSuffix
	// BodyIS *ferret.InvertedSuffix
}

func newIndex () {

} 

func (i *Index) RebuildWith(titles, hexes []string) {
        i.TitleIS = ferret.New(ExampleTitles, ExampleTitles, ExampleData, ExampleConverter)
}

func (i *Index) Query(term []byte) []string {

}
