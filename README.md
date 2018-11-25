# pubmed

Performs queries on the [Pubmed](https://www.ncbi.nlm.nih.gov/pubmed/) database.

> _PubMed comprises more than 28 million citations for biomedical literature from MEDLINE, life science journals, and online books. Citations may include links to full-text content from PubMed Central and publisher web sites._

The [Pubmed API](https://www.ncbi.nlm.nih.gov/books/NBK25497/) is powerful but can be a little curly to use. This package aims to make life a bit easier when performing basic queries against the Pubmed database.

## Installation

```bash
go get github.com/mikedonnici/pubmed
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

There is tonnes of info on the [Pubmed API](https://www.ncbi.nlm.nih.gov/books/NBK25497/). This is a basic overview of the `esearch` and `efetch` queries.

By way of example, the query below will fetch articles relating to _asthma_ that have been published in the last 7 days:

```sh
https://eutils.ncbi.nlm.nih.gov/entrez/eutils/esearch.fcgi?db=pubmed&retmode=json& \
datetype=pdat&reldate=7&retstart=0&retmax=20&term=asthma
```

### URL Params

- `datetype=pdate` - specify the _published_ date (`pdate`) as the filter
- `reldate=7` - include articles _published_ within last 7 days
- `retstart=0` - return a subset of the total results, starting with the first record
- `retmax=20` - return a maximum of 20 results

This will return a maximum of 20 article IDs, as below:

```json
{
    "header": {},
    "esearchresult": {
        "count": "90",
        "retmax": "20",
        "retstart": "0",
        "idlist": ["29783109", "29782937", "...up to max 20"]
    }
}
```

This `esearch` query returns a limited set of IDs so further queries are required to fetch the rest. Article summaries can then be fetched (in batches) by performing an `efetch` query.

Pubmed can also cache the entire result set allowing subsets to be retrieved on subsequent requests. This reduces a bit of batch wrangling and only required one additional param on the `esearch` query:

- `userhistory=y`

`retstart` and `retmax` can be removed as they are irrelevant for the next step, and will just default to 0 and 20 respectively.

```sh
https://eutils.ncbi.nlm.nih.gov/entrez/eutils/esearch.fcgi?db=pubmed&retmode=json& \
datetype=pdat&reldate=7&usehistory=y&term=asthma
```

The response now contains two additional fields, **`querykey`** and **`webenv`**, which are used to reference the stored results in subsequent requests.

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

As mentioned, `retmax` will default to a value of 20 which will limit `idlist` to 20 values. Of course, this value can be overridden if required. However, the _stored_ search results actually contain _all_ of the 90 articles indicated by the `count` value and can be retrieved using an `efetch` query (see below). The `esearch` query only has to be executed once.

Stored search results are retrieved by including the following url params on the `efetch` query:

- `query_key=1`
- `WebEnv=NCID_1_95852708_130.14.18.34_9001_1527045975_1203341711_0MetA0_S_MegaStore`

`WebEnv` is unique for each stored search.

The `efetch` query becomes:

```sh
https://eutils.ncbi.nlm.nih.gov/entrez/eutils/efetch.cgi?db=pubmed& \
retmode=xml&rettype=abstract&query_key=1&WebEnv=NCID_..._MegaStore
```

...and voila! All 90 results are there, regardless of `retmax` / `retstart` params in the previous search query.

_Note: at this time there is no option for JSON response (`retmode=json`) for `efetch` queries)_

If the initial `esearch` query returns a large result set then the `efetch` query will return a significantly larger file. Use `retmax` and `retstart` params on the `efetch` query to limit the articles in the response.

```sh
https://eutils.ncbi.nlm.nih.gov/entrez/eutils/efetch.cgi?db=pubmed&retmode=xml& \
rettype=abstract&retstart=0&retmax=10&query_key=1&WebEnv=NCID_..._MegaStore
```

The first 10 results: `retstart=0&retmax=10`, the next 10: `retstart=10&retmax=10`, and so on.
