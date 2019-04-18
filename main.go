package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/shopspring/decimal"
)

const exchangeRateAPI = "https://api.exchangeratesapi.io/latest"

type Response struct {
	Base  string                 `json:"base"`
	Rates map[string]json.Number `json:"rates"`
}

func main() {
	var amount = flag.String("amount", "1", "From currency amount")
	var srcSymbol = flag.String("src_symbol", "EUR", "Convert from currency")
	var destSymbol = flag.String("dst_symbol", "USD", "Convert to currency")
	flag.Parse()
	decimalAmount, err := decimal.NewFromString(*amount)
	if err != nil {
		log.Fatalf("error: amount is not a number:%s %s", decimalAmount, err)
	}
	if *srcSymbol == *destSymbol {
		fmt.Println(decimalAmount)
		return
	}
	resp, err := http.Get(exchangeRateAPI)
	if err != nil {
		log.Fatalf("error: performing http request:%s", err)
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("error: reading response body :%s", err)
	}
	if resp.StatusCode != http.StatusOK {
		log.Fatalf("error: incorrect status code while querying exchange api:%d %s", resp.StatusCode, string(body))
	}
	jsonResp := &Response{}
	if err := json.Unmarshal(body, jsonResp); err != nil {
		log.Fatalf("error: decoding received from exchange api json: %s", string(body))
	}
	ratesDecimal := map[string]decimal.Decimal{}
	for rateSym, rateNum := range jsonResp.Rates {
		rateDecNum, _ := decimal.NewFromString(rateNum.String())
		ratesDecimal[rateSym] = rateDecNum
	}
	if _, ok := ratesDecimal[*srcSymbol]; !ok && *srcSymbol != jsonResp.Base {
		log.Fatalf("error: src_symbol is not present in exchange api response, can't convert values")
	}
	if _, ok := ratesDecimal[*destSymbol]; !ok && *destSymbol != jsonResp.Base {
		log.Fatalf("error: dst_symbol is not present in exchange api response, can't convert values")
	}
	var res decimal.Decimal
	// one step convertation
	if *srcSymbol == jsonResp.Base {
		rateDecimal := ratesDecimal[*destSymbol]
		res = decimalAmount.Mul(rateDecimal)
	} else if *destSymbol == jsonResp.Base {
		rateDecimal := ratesDecimal[*srcSymbol]
		res = decimalAmount.Div(rateDecimal)
		// two step convertation
	} else {
		rateDecimalSrc := ratesDecimal[*srcSymbol]
		resInBase := decimalAmount.Div(rateDecimalSrc)
		rateDecimalDst := ratesDecimal[*destSymbol]
		res = resInBase.Mul(rateDecimalDst)
	}
	fmt.Println(res)
}
