package pubmed_test

import (
	"io/ioutil"
	"log"
	"testing"

	"github.com/matryer/is"
	"github.com/mikedonnici/pubmed"
)

var mockResponseJSON = map[string][]byte{
	"search": {},
}

var mockResponseXML = map[string][]byte{
	"articles": {},
}

// init sets up the mock responses
func init() {

	for i := range mockResponseJSON {
		f := i + ".json"
		xb, err := ioutil.ReadFile("testdata/" + f)
		if err != nil {
			log.Fatalf("Cannot read %s\n", f)
		}
		mockResponseJSON[i] = xb
	}

	for i := range mockResponseXML {
		f := i + ".xml"
		xb, err := ioutil.ReadFile("testdata/" + f)
		if err != nil {
			log.Fatalf("Cannot read %s\n", f)
		}
		mockResponseXML[i] = xb
	}
}

// TestReadSearchResponse tests the un-marshalling of a mocked Pubmed esearch response
func TestReadSearchResponse(t *testing.T) {
	is := is.New(t)
	ps := pubmed.NewQuery("not used for a mock")
	xb := []byte(mockResponseJSON["search"])
	ps.ReadSearchResponse(xb)
	is.Equal(ps.ResultCount, 36332)                                                                     // Incorrect result count
	is.Equal(ps.Key, "1")                                                                               // Incorrect query key
	is.Equal(ps.WebEnv, "NCID_1_243404818_130.14.22.215_9001_1527820979_1002563964_0MetA0_S_MegaStore") // Incorrect web env
}

// TestReadArticleResponse tests the un-marshalling of a mocked Pubmed efetch response
func TestReadArticleResponse(t *testing.T) {
	is := is.New(t)
	xb := []byte(mockResponseXML["articles"])
	xa, err := pubmed.ReadArticlesResponse(xb)
	is.NoErr(err)                 // ReadArticleResponse error
	is.Equal(len(xa.Articles), 2) // Should be 2 articles
	exp := "MRI with gadofosveset: A potential marker for permeability in myocardial infarction."
	got := xa.Articles[0].Title
	is.Equal(exp, got) // Article title

	// Trim description to first 17 chars
	exp = "Acute ischemia is"
	got = xa.Articles[0].Description[:17]
	is.Equal(exp, got) // Article description
}

// Real queries

//func TestRealSearch(t *testing.T) {
//	is := is.New(t)
//	ps := pubmed.NewQuery(query)
//	ps.BackDays = 100
//	err := ps.Search()
//	is.NoErr(err)               // Search error
//	is.True(ps.ResultCount > 0) // No results for last 100 days?
//	set, err := ps.Articles(0, 1000)
//	is.NoErr(err) // Fetch error
//
//	for _, a := range set.Articles {
//		if len(a.MeshHeadings) > 0 {
//			a.Print()
//		}
//	}
//}

//func TestFetchArticles(t *testing.T) {
//	is := is.New(t)
//	ps := pubmed.NewQuery(query)
//	ps.Name = "Cardiology"
//	xps, err := ps.Articles("29735362", "29730991")
//	is.NoErr(err) // Error fetching articles
//	fmt.Println(xps)
//}
