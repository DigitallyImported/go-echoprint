package echoprint

//
// import (
// 	"fmt"
//
// 	"github.com/vanng822/go-solr/solr"
// )
//
// type echoprintResultParser solr.StandardResultParser
//
// func (parser *echoprintResultParser) Parse(response *solr.SolrResponse) (*solr.SolrResult, error) {
// 	sr := &solr.SolrResult{}
// 	sr.Results = new(solr.Collection)
// 	sr.Status = response.Status
//
// 	parser.ParseResponseHeader(response, sr)
//
// 	if response.Status == 0 {
// 		err := parser.ParseResponse(response, sr)
// 		if err != nil {
// 			return nil, err
// 		}
//
//     parser.ParseStats(response, sr)
// 	} else {
// 		parser.ParseError(response, sr)
// 	}
//
// 	return sr, nil
// }
//
// func (parser *echoprintResultParser) ParseResponseHeader(response *solr.SolrResponse, sr *solr.SolrResult) {
// 	if responseHeader, ok := response.Response["responseHeader"].(map[string]interface{}); ok {
// 		sr.ResponseHeader = responseHeader
// 	}
// }
//
// func (parser *echoprintResultParser) ParseError(response *solr.SolrResponse, sr *solr.SolrResult) {
// 	if err, ok := response.Response["error"].(map[string]interface{}); ok {
// 		sr.Error = err
// 	}
// }
//
// func ParseDocResponse(docResponse map[string]interface{}, collection *solr.Collection) {
// 	collection.NumFound = int(docResponse["numFound"].(float64))
// 	collection.Start = int(docResponse["start"].(float64))
// 	if docs, ok := docResponse["docs"].([]interface{}); ok {
// 		collection.Docs = make([]Document, len(docs))
// 		for i, v := range docs {
// 			collection.Docs[i] = Document(v.(map[string]interface{}))
// 		}
// 	}
// }
//
// // ParseSolrResponse will assign result and build sr.docs if there is a response.
// // If there is no response or grouped property in response it will return error
// func (parser *StandardResultParser) ParseResponse(response *SolrResponse, sr *SolrResult) (err error) {
// 	if resp, ok := response.Response["response"].(map[string]interface{}); ok {
// 		ParseDocResponse(resp, sr.Results)
// 	} else if grouped, ok := response.Response["grouped"].(map[string]interface{}); ok {
// 		sr.Grouped = grouped
// 	} else {
// 		err = fmt.Errorf(`Standard parser can only parse solr response with response object,
// 					ie response.response and response.response.docs. Or grouped response
// 					Please use other parser or implement your own parser`)
// 	}
//
// 	return err
// }
