# draft-content-suggestions

[![Circle CI](https://circleci.com/gh/Financial-Times/draft-content-suggestions/tree/master.png?style=shield)](https://circleci.com/gh/Financial-Times/draft-content-suggestions/tree/master)[![Go Report Card](https://goreportcard.com/badge/github.com/Financial-Times/draft-content-suggestions)](https://goreportcard.com/report/github.com/Financial-Times/draft-content-suggestions) [![Coverage Status](https://coveralls.io/repos/github/Financial-Times/draft-content-suggestions/badge.svg)](https://coveralls.io/github/Financial-Times/draft-content-suggestions)

## Introduction

Draft Content Suggestions as a microservice, provides consolidated suggestions via fetching draft content
from Draft Content service and querying Suggestions Umbrella service.  

## Installation

Download the source code, dependencies and test dependencies:

        go get github.com:Financial-Times/draft-content-suggestions
        cd $GOPATH/src/github.com/Financial-Times/draft-content-suggestions
        go build .

## Running locally

1. Run the tests and install the binary:

        go test ./...
        go install

2. Run the binary (using the `help` flag to see the available optional arguments):

        $GOPATH/bin/draft-content-suggestions [--help]

Options:

        --app-system-code="draft-content-suggestions"            System Code of the application ($APP_SYSTEM_CODE)
        --app-name="Annotation Suggestions API"                   Application name ($APP_NAME)
        --port="8080"                                           Port to listen on ($APP_PORT)
        --draft-content-endpoint="http://localhost:9000/drafts/content" Draft Content Service
        --draft-content-gtg-endpoint="http://localhost:9000/__gtg" Draft Content Health Service
        --suggestions-umbrella-endpoint="http://test.api.ft.com/content/suggest" Suggestions Umbrella Service
        --suggestions-api-key="" Suggestions service apiKey

3. Test:

    1. Either using curl:

            curl http://localhost:8080/drafts/content/143ba45c-2fb3-35bc-b227-a6ed80b5c517/suggestions | json_pp

    1. Or using [httpie](https://github.com/jkbrzt/httpie):

            http GET http://localhost:8080/drafts/content/143ba45c-2fb3-35bc-b227-a6ed80b5c517/suggestions

## Build and deployment

_How can I build and deploy it (lots of this will be links out as the steps will be common)_

* Built by Docker Hub on merge to master: [coco/draft-content-suggestions](https://hub.docker.com/r/coco/draft-content-suggestions/)
* CI provided by CircleCI: [draft-content-suggestions](https://circleci.com/gh/Financial-Times/draft-content-suggestions)

## API Endpoints

### POST `/drafts/content/suggestions` - Returns a list of suggestions for a given content piece

* Staging endpoint: <https://api-t.ft.com/drafts/content/suggestions>
* Production endpoint: <https://api.ft.com/drafts/content/suggestions>

Note: An API key containing the policy `PAC Platform` is necessary to access this API.

The endpoint expects one of four eligible `Content-Type` header values and a body

Here are examples for each `Content-Type`:

<details><summary>Example with HTTP Header `Content-Type` - `application/vnd.ft-upp-article+json` and body:</summary>

```json
{
    "_internal": {},
    "accessLevel": "free",
    "alternativeTitles": {
        "promotionalTitle": "Invesco launches first ‘green building’ ETF"
    },
    "bodyXML": "<body xmlns:opaque=\"http://www.ft.com/upp/namespaces/internal-content/opaque\"><content data-embedded=\"true\" type=\"http://www.ft.com/ontology/content/ImageSet\" id=\"2f8947e5-df52-4fb1-a83f-fc285d8cb06a\"></content><p>Invesco is launching what is believed to be the world’s first “green building” exchange traded fund, aiming to fill a gap in portfolios in a world increasingly focused on climate change.</p><p>The Invesco MSCI Green Building ETF (GBLD), due to list on the New York stock exchange today will target the buildings sector, estimated by the UN Environment Programme to account for 38 per cent of global carbon emissions.</p><p>“We can’t talk about decarbonisation without talking about buildings and infrastructure,” said John Hoffman, head of ETFs and indexed strategies, Americas, at Invesco.</p><p>The launch comes amid a surge in demand for ESG investment, both in equities and fixed income, with total assets in the sector rising 50 per cent to a record $1.7tn last year, according to Morningstar.</p><p>While GBLD is also equity-based, it is designed to invest in real estate companies whose estates boast relatively high energy efficiency, have a healthier indoor environmental quality and make use of environmentally friendlier construction materials.</p><p>It will also hold companies involved in the design, construction, redevelopment, retrofitting or third-party certification of green-certified properties to effect climate change mitigation and adaptation.</p><p>“[The ETF] will be the first to focus specifically on the entire green building ecosystem,” said Hoffman.</p><p>GBLD is likely to receive a mixed reaction. Ben Johnson, director of global ETF research at Morningstar, said: “this is truly a first-of-its-kind product.”</p><experimental><div data-layout-name=\"card\" class=\"n-content-layout\" data-layout-width=\"inset-left\"><div class=\"n-content-layout__container\"><h3>Twice weekly newsletter</h3><div class=\"n-content-layout__slot\" data-slot-width=\"true\"><img src=\"https://d1e00ek4ebabms.cloudfront.net/production/2ebc1113-cce8-4a21-ab04-ea9d70cddebe.jpg\" data-image-type=\"image\" data-copyright=\"\" longdesc=\"\" alt=\"\"></img><p>Energy is the world’s indispensable business and Energy Source is its newsletter. Every Tuesday and Thursday, direct to your inbox, Energy Source brings you essential news, forward-thinking analysis and insider intelligence.&#160;<a href=\"https://ep.ft.com/newsletters/subscribe?emailId=5ef959a50301a30004e74bc9&amp;segmentId=22011ee7-896a-8c4c-22a0-7603348b7f22&amp;newsletterIds=5655d099e4b01077e911d60f\">Sign up here</a>.</p></div></div></div></experimental><p>However, Peter Sleep, senior portfolio manager at 7 Investment Management, was more dismissive, labelling it “another way to package up property companies in an exciting thematic wrapper”.</p><p>Invesco, the world’s fourth-largest ETF manager with $400bn in assets, declined to name potential holdings ahead of launch. However, the fund is likely to be heavily exposed to the commercial real estate investment trust (Reit) industry.</p><p>It will track the <a href=\"https://www.msci.com/documents/10199/2befec3a-e178-460d-a5b1-e79555dee387\">MSCI Global Green Building Index</a>, whose largest holdings at the end of March included Reits such as Boston Properties, Nippon Building Fund, Japan Real Estate Investment and Vornado Realty Trust, as well as other large property companies like Unibail-Rodamco-Westfield and Berkeley Group Holdings.</p><p>At least 50 per cent of a real estate company’s estate must be certified as “green” by Leadership in Energy and Environmental Design in the US or equivalent bodies in other countries in order to be included in the index.</p><p>Certification typically involves conserving natural resources, being constructed with recycled waste, avoiding toxic emissions, limiting water and energy use or contributing to a “safe, healthy built environment”.</p><p>As a result the index’s top 10 holdings are very different to those of existing global real estate ETFs, not least in their geographical diversification.</p><p>As of March 31, the index had about 26 per cent exposure to each of the US and Japan, with 11 per cent in both France and Singapore.</p><p>In contrast the industry-leading global property ETF, the $3bn iShares Global Reit ETF (<a href=\"https://etf.ft.com/funds/6625/\">REET</a>), has a 66 per cent weighting to the US and only 9 per cent to Japan.</p><experimental><div data-layout-name=\"card\" class=\"n-content-layout\" data-layout-width=\"inset-left\"><div class=\"n-content-layout__container\"><h3>Climate Capital</h3><div class=\"n-content-layout__slot\" data-slot-width=\"true\"><img src=\"https://d1e00ek4ebabms.cloudfront.net/production/8634607b-3b27-4b19-aa33-d58f9e846b26.jpg\" data-image-type=\"image\" data-copyright=\"\" longdesc=\"\" alt=\"\"></img><p>Where climate change meets business, markets and politics. <a href=\"http://www.ft.com/climate-capital\">Explore the FT’s coverage here</a>&#160;</p></div></div></div></experimental><p>“While GBLD’s underlying index is mostly made up of Reits, its portfolio has a much different complexion than a broad, cap-weighted global real estate index,” said Johnson “This speaks to its ESG remit as well as the fact that it includes building material suppliers, home builders, and property managers that fit this remit.”</p><p>The MSCI Global Green Building Index has also been more volatile, falling more than REET in 2020 (particularly during the market sell-off in the first quarter), 2018 and 2016 (when REET rose) but returning more in 2019 and 2017 — massively so in the latter case, by 29.2 per cent vs 6.8 per cent.</p><p>“One of the issues with thematic ETFs is that you tend to get unrewarded volatility, and fees,” said Sleep. “This seems a case in point.” GBLD will charge 39 basis points a year, compared to 14bp for the iShares ETF.</p><p>Todd Rosenbluth, head of ETF and mutual fund research at CFRA Research, said the MSCI index appeared to be focused on more economically sensitive sub-sectors such as offices and retail, compared to the greater weighting in everything from logistics and data centres to healthcare facilities and self-storage in other real estate ETFs such as REET.</p><p>Despite this, Rosenbluth believed GBLD had a place.</p><p>“Many investors view Reits as their own investment style or asset class [as opposed to equities] so I think it makes sense for people who are more ESG-focused that there is a portfolio of companies where the buildings themselves are green energy oriented,” he said.</p><p>“I can see this ETF making sense and fitting in for investors.”</p><p>One possible downside for the ETF is that it is launching in an environment when working from home may become a permanent option for many workers, reducing demand for space.</p><p>However Invesco said it saw an “increased desirability” for good air filtration systems, “which are a key element of green building ratings”.</p><p>Rene Reyna, head of thematic and specialty product strategy at Invesco, said he believed rising urban populations would continue to support demand for office space.</p><content data-embedded=\"true\" type=\"http://www.ft.com/ontology/content/Video\" id=\"8e10dbc3-1973-4ed6-987f-227b2a16b28d\"></content></body>",
    "byline": "Steve Johnson",
    "canBeDistributed": "yes",
    "canBeSyndicated": "yes",
    "comments": {
        "enabled": true
    },
    "editorialDesk": "/FT/MarketNews/ETFs",
    "firstPublishedDate": "2021-04-22T04:00:50.146Z",
    "identifiers": [
        {
            "authority": "http://api.ft.com/system/cct",
            "identifierValue": "88db6314-45e1-45c9-898f-d98e2ff60967"
        }
    ],
    "lastModified": "2024-02-18T22:51:35Z",
    "mainImage": "2f8947e5-df52-4fb1-a83f-fc285d8cb06a",
    "publishReference": "tid_search_reingest_carousel_0000379784_gentx",
    "publishedDate": "2021-04-22T04:00:50.146Z",
    "standfirst": "UN estimates the building sector accounts for 38% of global carbon emissions",
    "standout": {
        "editorsChoice": false,
        "exclusive": false,
        "scoop": false
    },
    "title": "Invesco launches first ‘green building’ ETF",
    "type": "Article",
    "uuid": "88db6314-45e1-45c9-898f-d98e2ff60967",
    "webUrl": "https://www.ft.com/content/88db6314-45e1-45c9-898f-d98e2ff60967"
}
```

</details>

<details><summary>Example with HTTP Header `Content-Type` - `application/vnd.ft-upp-live-blog-package+json` and body:</summary>

```json
{
    "_internal": {
        "leadImages": [
            {
                "id": "42ac5b0f-d934-44eb-ab1e-14801249cebe",
                "type": "standard"
            },
            {
                "id": "38e0eb2b-d1d8-4e82-83c1-25e4e1794f41",
                "type": "square"
            },
            {
                "id": "8751c053-9ed1-4b86-9764-7bce6906ba96",
                "type": "wide"
            }
        ],
        "summary": {
            "bodyXML": "<body><p><strong>Today’s main headlines:</strong>&#160;</p><ul><li><p>US jobless claims fall to fresh pandemic-era low</p></li><li><p>Alzheimer’s and heart disease overtake Covid as England’s top cause of death </p></li><li><p>India’s health infrastructure unravels as record 315,000 daily cases reported</p></li><li><p>US airlines say recovery is speeding up as passengers return</p></li><li><p>ECB vows to persist with faster bond purchases to prop up recovery</p></li></ul></body>"
        },
        "topper": {
            "backgroundColour": "auto",
            "headline": "",
            "layout": "full-bleed-offset",
            "standfirst": ""
        }
    },
    "accessLevel": "free",
    "alternativeTitles": {
        "promotionalTitle": "Coronavirus: Intel, Mattel and Snap optimistic about outlooks - as it happened"
    },
    "byline": "Mamta Badkar, Matthew Rocco and Peter Wells in New York, Sarah Provan, Oliver Ralph, Alistair Gray, Leke Oso Alabi and George Steer in London, and Gary Jones and Alice Woodhouse in Hong Kong",
    "canBeDistributed": "yes",
    "canBeSyndicated": "verify",
    "comments": {
        "enabled": true
    },
    "contentPackage": "38a73eb0-f58e-49c6-b546-34081d4c4779",
    "firstPublishedDate": "2021-04-21T22:32:46.234Z",
    "identifiers": [
        {
            "authority": "http://api.ft.com/system/cct",
            "identifierValue": "6ba73186-f94d-4763-9c44-9a8a4ade8788"
        }
    ],
    "lastModified": "2021-05-04T00:57:08.381Z",
    "mainImage": "0330df98-c9a4-4562-968a-07fcad9fbc3e",
    "publishReference": "tid_cct_6ba73186-f94d-4763-9c44-9a8a4ade8788_1620089824510",
    "publishedDate": "2021-04-22T16:41:14.548Z",
    "realtime": false,
    "standfirst": "",
    "standout": {
        "editorsChoice": false,
        "exclusive": false,
        "scoop": false
    },
    "title": "Coronavirus: Intel, Mattel and Snap optimistic about outlooks - as it happened",
    "type": "LiveBlogPackage",
    "uuid": "6ba73186-f94d-4763-9c44-9a8a4ade8788",
    "webUrl": "https://www.ft.com/content/6ba73186-f94d-4763-9c44-9a8a4ade8788"
}
```

</details>

<details><summary>Example with HTTP Header `Content-Type` - `application/vnd.ft-upp-live-blog-post+json` and body:</summary>

```json
{
    "_internal": {
        "publishCount": 1
    },
    "accessLevel": "subscribed",
    "alternativeTitles": {
        "promotionalTitle": "Promotion Title"
    },
    "bodyXML": "<body xmlns:opaque=\"http://www.ft.com/upp/namespaces/internal-content/opaque\"><p>Body Here</p></body>",
    "byline": "John Doe in Paris ",
    "canBeDistributed": "yes",
    "canBeSyndicated": "verify",
    "comments": {
        "enabled": true
    },
    "firstPublishedDate": "2021-07-13T07:32:19.308Z",
    "identifiers": [
        {
            "authority": "http://api.ft.com/system/cct",
            "identifierValue": "a2645da0b-8282-423f-8709-73db20c7fa5b"
        }
    ],
    "lastModified": "2024-02-19T00:02:07Z",
    "publishReference": "tid_...",
    "publishedDate": "2021-07-13T07:32:19.308Z",
    "standfirst": "",
    "standout": {
        "breakingNews": false,
        "editorsChoice": false,
        "exclusive": false,
        "scoop": false
    },
    "title": "Test Title",
    "type": "LiveBlogPost",
    "uuid": "a2645da0b-8282-423f-8709-73db20c7fa5b",
    "webUrl": "https://www.ft.com/content/a2645da0b-8282-423f-8709-73db20c7fa5b"
}
```

</details>

<details><summary>Example with HTTP Header `Content-Type` - `application/vnd.ft-upp-content-placeholder+json` and body:</summary>

```json
{
    "_internal": {
        "publishCount": 7
    },
    "accessLevel": "free",
    "canBeDistributed": "yes",
    "canBeSyndicated": "yes",
    "canonicalWebUrl": "https://www.ft.com/content/dca43692-2a6a-4d99-bf35-1d032452bbfb",
    "editorialDesk": "/FT/SpecialReports",
    "identifiers": [
        {
            "authority": "http://api.ft.com/system/cct",
            "identifierValue": "dca43692-2a6a-4d99-bf35-1d032452bbfb"
        }
    ],
    "lastModified": "2023-11-23T10:19:03Z",
    "mainImage": "025c7e34-b03d-4d3a-977b-79df59099578",
    "publication": [
        "88fdde6c-2aa4-4f78-af02-9f680097cfd6"
    ],
    "publishReference": "tid_cct_dca43692-2a6a-4d99-bf35-1d032452bbfb_1700734742990",
    "publishedDate": "2023-11-07T12:46:15.074Z",
    "standfirst": "Mainland Chinese are the most interested, but even they prefer Singapore",
    "title": "Hong Kong’s family office push falls flat with global billionaires",
    "type": "Content",
    "uuid": "dca43692-2a6a-4d99-bf35-1d032452bbfb",
    "webUrl": "https://asia.nikkei.com/Business/Finance/Hong-Kong-s-family-office-push-falls-flat-with-global-billionaires"
}
```

</details>

### Logging

* The application uses [go-logger/v2](https://github.com/Financial-Times/go-logger/tree/v2); the log library is initialised in [main.go](main.go).
* NOTE: `/__build-info` and `/__gtg` endpoints are not logged as they are called every second from varnish/vulcand and this information is not needed in logs/splunk.

## Change/Rotate sealed secrets

Please reffer to documentation in [pac-global-sealed-secrets-eks](https://github.com/Financial-Times/pac-global-sealed-secrets-eks/blob/master/README.md). Here are explained details how to create new, change existing sealed secrets.
