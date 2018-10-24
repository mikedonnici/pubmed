# pubmed

Performs queries on the [Pubmed](https://www.ncbi.nlm.nih.gov/pubmed/) database.

> *PubMed comprises more than 28 million citations for biomedical literature from MEDLINE, 
life science journals, and online books. Citations may include links to full-text content 
from PubMed Central and publisher web sites.*

The [Pubmed API](https://www.ncbi.nlm.nih.gov/books/NBK25497/) is powerful but can be a little curly to use. 
This package aims to make life a bit easier when performing basic queries against the Pubmed database.

## Installation

```bash
$ go get github.com/mikedonnici/pubmed
``` 

## Usage

```go
package main

import (
	"fmt"
	"net/url"
	
	"github.com/mikedonnici/pubmed"
)

// batch size
const retMax = 100

func main() {
	
        // query pubmed
        t := "quadricuspid aortic valve"
        p := pubmed.NewQuery(url.PathEscape(t))
        p.BackDays = 365 // default is 7
    	err := p.Search()
    	if err != nil {
    		// handle error
    	} 
    		
        for i := 0; i < p.ResultCount; i++ {
            xa, err := p.Articles(i, retMax)
            if err != nil {
                // handle error
            }
            for _, a := range xa.Articles {
                // do something fancy with the article
                fmt.Println(a)
            }
        }	
}
```

## Intro to E-utilities (Pubmed API)

For example, the query below searched for asthma articles with a publish date within the past 7 days:

```http
https://eutils.ncbi.nlm.nih.gov/entrez/eutils/esearch.fcgi?db=pubmed&retmode=json \ 
&datetype=pdat&reldate=7&retstart=0&retmax=20&term=asthma
```

**URL Params**

* `datetype=pdate` - query the date the article was *published*
* `reldate=7` - include articles *published* within last 7 days
* `retstart=0` - return a subset of the total results, starting with the first record
* `retmax=20` - return a maximum of 20 results

This will return a maximum of 20 article IDs as below:

```json
{
    "header": {},
    "esearchresult": {
        "count": "90",
        "retmax": "20",
        "retstart": "0",
        "idlist": ["29783109", "29782937", "...up to 20"]
    }
}
```

Here we only have a list of IDs and so must perform further queries to fetch 
subsequent sets of records and additional queries to fetch the actual article summaries. 

Luckily, the initial query can be altered to tell Pubmed to cache the entire result set for us.

The url can be simplified a bit by removing `retstart` and `retmax` -
they are irrelevant for the next step and will just default to 0 and 20
respectively.

A new param is added to the request url - `userhistory=y`. This will tell
 Pubmed to store the results for subsequent retrieval.

```http
https://eutils.ncbi.nlm.nih.gov/entrez/eutils/esearch.fcgi?db=pubmed&retmode=json& \ 
datetype=pdat&reldate=7&usehistory=y&term=asthma
```

The response now contains two additional fields, **`querykey`** and **`webenv`**.
These fields are used to reference the stored results in subsequent requests.

```json
{
    "header": {},
    "esearchresult": {
        "count": "90",
        "retmax": "20",
        "retstart": "0",
        "querykey": "1",
        "webenv": "NCID_1_95852708_130.14.18.34_9001_1527045975_1203341711_0MetA0_S_MegaStore",
        "idlist": ["29783109", "29782937", "+18 more"]
    }
}
```

**Important** - although the `idlist` only displays the 20 results limited
by the default value for `retmax`, the stored search results contain *all* 90 of articles found.

The stored search results can then be retrieved by adding the url params:
 * `query_key=1`
 * `WebEnv=NCID_1_95852708_130.14.18.34_9001_1527045975_1203341711_0MetA0_S_MegaStore`

```http
https://eutils.ncbi.nlm.nih.gov/entrez/eutils/efetch.cgi?db=pubmed& \ 
retmode=xml&rettype=abstract&query_key=1&WebEnv=NCID_..._MegaStore
```

...and voila! All 90 results are there, regardless of `retmax` or `retstart`
values in the previous query.

**Caution** - If the initial search returned a large result set then the `efetch` query (as above) will 
return a large file.

To fetch subsets of the articles in the result set use `retmax` and `retstart` url params. 

```http
https://eutils.ncbi.nlm.nih.gov/entrez/eutils/efetch.cgi?db=pubmed&retmode=xml& \ 
rettype=abstract&retstart=0&retmax=10&query_key=1&WebEnv=NCID_..._MegaStore
```

The first 10 results: `retstart=0&retmax=10`

The next 10: `retstart=10&retmax=10`

...and so on.

For large result sets processing in batches will be more efficient but leveraging the Pubmed cache feature 
reduces some of the burden.

