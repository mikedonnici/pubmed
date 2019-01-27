// Copyright 2018 Mike Donnici. All rights reserved.
// Use of this source code is governed by an MIT
// license that can be found in the LICENSE file.

// Package pubmed performs queries on the Pubmed database.
package pubmed

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"

	"strings"

	"github.com/pkg/errors"
)

const baseURL = "https://eutils.ncbi.nlm.nih.gov/entrez/eutils"
const searchURL = baseURL + "/esearch.fcgi?db=pubmed&retmode=json&usehistory=y"
const fetchURL = baseURL + "/efetch.fcgi?db=pubmed&retmode=xml&rettype=abstract"
const queryBackDays = "&reldate=%v&datetype=edat"
const querySearchTerm = "&term=%v"
const queryReturnMax = "&retmax=%v"
const queryStartIndex = "&retstart=%v"
const queryKey = "&query_key=%v"
const queryWebEnv = "&WebEnv=%s"

const defaultBackDays = 7

// Query represents a search request to the Pubmed esearch endpoint, the results of which are stored at Pubmed
// and subsequently referenced by Key and WebEnv
type Query struct {
	BackDays    int
	Term        string
	ResultCount int
	Key         string `json:"querykey"`
	WebEnv      string `json:"webenv"`
}

// NewQuery returns a pointer to a Query with some defaults set
func NewQuery(query string) *Query {
	return &Query{
		BackDays: defaultBackDays,
		Term:     query,
	}
}

// Search executes the query, results are stored at Pubmed and referenced by the Key and WenEnv values
func (ps *Query) Search() error {
	qURL := searchURL + fmt.Sprintf(queryBackDays, ps.BackDays) + fmt.Sprintf(querySearchTerm, ps.Term)
	xb, err := responseBody(qURL)
	if err != nil {
		return err
	}
	return ps.ReadSearchResponse(xb)
}

// ReadSearchResponse decodes the response body from a search request into the relevant Query fields
func (ps *Query) ReadSearchResponse(response []byte) error {

	var r = struct {
		Result struct {
			Count    string `json:"count"`
			QueryKey string `json:"querykey"`
			WebEnv   string `json:"webenv"`
		} `json:"esearchresult"`
	}{}

	err := json.Unmarshal(response, &r)
	if err != nil {
		return errors.Wrap(err, "Search, Unmarshal")
	}

	count, err := strconv.Atoi(r.Result.Count)
	if err != nil {
		return errors.Wrap(err, "Search, Atoi")
	}
	ps.ResultCount = count
	ps.Key = r.Result.QueryKey
	ps.WebEnv = r.Result.WebEnv

	return nil
}

// Articles fetches a set of articles from the Pubmed cache referenced by Key and WebEnv. The response is an xml
// payload that is unmarshaled into a PubMedSet.
// Ref: https://www.ncbi.nlm.nih.gov/books/NBK25499/#_chapter4_EFetch_
func (ps *Query) Articles(startIndex, retMax int) (ArticleSet, error) {

	qURL := fetchURL +
		fmt.Sprintf(queryKey, ps.Key) +
		fmt.Sprintf(queryWebEnv, ps.WebEnv) +
		fmt.Sprintf(queryStartIndex, startIndex) +
		fmt.Sprintf(queryReturnMax, retMax)
	xb, err := responseBody(qURL)
	if err != nil {
		return ArticleSet{}, errors.Wrap(err, "Articles could not get response body")
	}

	return ReadArticlesResponse(xb)
}

// ArticleByPMID fetches a single article by Pubmed ID
func ArticleByPMID(pmid string) (Article, error) {

	qURL := fetchURL + "&id=" + pmid
	xb, err := responseBody(qURL)
	if err != nil {
		return Article{}, errors.Wrap(err, "Could not fetch article with pmid "+pmid)
	}

	// Still returns an article set, we just want the first one
	xa, err := ReadArticlesResponse(xb)
	return xa.Articles[0], err
}

// ReadArticlesResponse decodes the xml response body from a request to fetch articles, into an ArticleSet
func ReadArticlesResponse(response []byte) (ArticleSet, error) {

	var set ArticleSet

	err := xml.Unmarshal(response, &set)
	if err != nil {
		return set, errors.Wrap(err, "ReadArticlesResponse")
	}

	for i, a := range set.Articles {
		set.Articles[i].Title = replaceQuotes(set.Articles[i].Title)
		set.Articles[i].Description = replaceQuotes(description(a))
		set.Articles[i].PubDate, _ = bestPubDate(a) // todo ... ignore the error?
		set.Articles[i].Keywords = mergeKeywords(a)
		set.Articles[i].URL = articleURL(a)
		set.Articles[i].Citation = citation(a)

	}

	return set, nil
}

// responseBody returns the response body from a GET request as a []byte
func responseBody(url string) ([]byte, error) {

	httpClient := &http.Client{Timeout: 90 * time.Second}
	r, err := httpClient.Get(url)
	if err != nil {
		return nil, errors.Wrap(err, "responseBody")
	}
	defer r.Body.Close()

	return ioutil.ReadAll(r.Body)
}

// replaceQuotes replaces double quotes with single quotes, to avoid breaking some JSON encoded/decode functions.
func replaceQuotes(s string) string {
	return strings.Replace(s, `"`, `'`, -1)
}
