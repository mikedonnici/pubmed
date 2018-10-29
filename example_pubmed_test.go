package pubmed_test

import (
	"fmt"
	"github.com/mikedonnici/pubmed"
)

const backDays = 365
const query = "bee allergies"


func Example() {

	// Run a query
	p := pubmed.NewQuery(query)
	p.BackDays = backDays
	err := p.Search()
	if err != nil {
		panic(err)
	}

	// Fetch the first 10 articles in the result set
	set, err := p.Articles(0, 10)
	if err != nil {
		panic(err)
	}

	for _, a := range set.Articles {
		if len(a.MeshHeadings) > 0 {
			fmt.Println(a.Title)
		}
	}
}
