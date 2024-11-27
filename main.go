package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
)

const (
	localIndexer = "http://localhost:8546/graphql/query"
	test4Indexer = "https://indexer.test4.gno.land:443/graphql/query"
	test5Indexer = "https://indexer.test5.gno.land:443/graphql/query"
)

func main() {
	res, err := fetchData(localIndexer, allTxsQuery)
	if err != nil {
		panic(err)
	}

	uniqueAddresses := make(map[string]bool)
	parseData(string(res), uniqueAddresses)

	fmt.Printf("Total unique addresses: %d\n", len(uniqueAddresses))
	for addr := range uniqueAddresses {
		fmt.Println(addr)
	}
}

func parseData(data string, addrs map[string]bool) {
	var re = []string{
		`"caller":"([a-z0-9]+)"`,
		`"from_address":"([a-z0-9]+)"`,
		`"to_address":"([a-z0-9]+)"`,
		`"creator":"([a-z0-9]+)"`,
	}

	for _, r := range re {
		reg := regexp.MustCompile(r)
		matches := reg.FindAllStringSubmatch(data, -1)

		for _, m := range matches {
			_, ok := addrs[m[1]]
			if !ok {
				addrs[m[1]] = true
			}
		}
	}
}

func fetchData(endpoint, query string) ([]byte, error) {
	payload := map[string]string{
		"query": query,
	}
	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal query: %v", err)
	}

	resp, err := http.Post(endpoint, "application/json", bytes.NewBuffer(jsonPayload))
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("received non-200 response: %s", resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %v", err)
	}

	return body, nil
}

const allTxsQuery = `query getAllTransactions {
  getTransactions(
    where: { _or: [
      {messages: {value: {BankMsgSend: {from_address: {exists: true}}}}},
      {messages: {value: {MsgRun: {caller: {exists: true}}}}},
      {messages: {value: {MsgAddPackage: {creator: {exists: true}}}}},
      {messages: {value: {MsgCall: {caller: {exists: true}}}}}]},
  ) {
    block_height
    messages {
      value {
        ... on BankMsgSend {
          __typename
          from_address
          to_address
          amount
        }
        ... on MsgCall {
          __typename
          caller
        }
         ... on MsgAddPackage {
          __typename
          creator
        }
        ... on MsgRun {
          __typename
          caller
        }
      }
    }
  }
}`
