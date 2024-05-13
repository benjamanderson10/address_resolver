package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
)

type ValidityResponse struct {
	ResultStatus string   `json:"resultStatus"`
	AddressList  []string `json:"title"`
}

type ValidityParams struct {
	CompanyName string
	Address1    string
	Address2    string
	City        string
	State       string
	UrbanCode   string
	Zip         string
}

func main() {
	addresses := []ValidityParams{}
	unfinished := []string{}
	incorrect := []ValidityParams{}

	file, err := os.Open("data.txt")
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	defer file.Close()

	wrfile, err := os.Create("out.txt")
	if err != nil {
		panic(err)
	}
	defer wrfile.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {

		parts := strings.Split(scanner.Text(), "\t")

		vp := ValidityParams{}

		var new_parts []string
		for _, a := range parts {
			if a == "" {
				continue
			}
			new_parts = append(new_parts, a)
		}
		if len(new_parts) < 4 {
			unfinished = append(unfinished, scanner.Text())
			continue
		}
		vp.Address1 = new_parts[0]
		vp.City = new_parts[1]
		vp.State = new_parts[2]
		vp.Zip = new_parts[3]

		addresses = append(addresses, vp)

		_, err := wrfile.WriteString(fmt.Sprintf("%s, %s, %s, %s\n", vp.Address1, vp.City, vp.State, vp.Zip))
		if err != nil {
			panic(err)
		}
	}

	// for _, a := range addresses {
	// 	fmt.Println(a.Address1)
	// }

	uri := "https://tools.usps.com/tools/app/ziplookup/zipByAddress"
	for i, a := range addresses {
		fmt.Printf("testing %d/%d ... ", i+1, len(addresses))
		res := test_address(uri, a)

		if res != "SUCCESS" {
			incorrect = append(incorrect, a)
			fmt.Println("FAILURE")
		} else {
			fmt.Println("SUCCESS")
		}

		if i > 600 {
			break
		}
	}

	fmt.Println("INCORRECT ADDRESSES:")
	for _, a := range incorrect {
		fmt.Printf("%s, %s, %s, %s\n", a.Address1, a.City, a.State, a.Zip)
	}

}

func test_address(uri string, params ValidityParams) string {
	values := url.Values{
		"companyName": {params.CompanyName},
		"address1":    {params.Address1},
		"address2":    {params.Address2},
		"city":        {params.City},
		"state":       {params.State},
		"urbanCode":   {params.UrbanCode},
		"zip":         {params.Zip},
	}
	//"companyName=&address1=1209+Merrill+Hall&address2=&city=Provo&state=UT&urbanCode=&zip=84602"
	payload := strings.NewReader(values.Encode())

	client := &http.Client{}
	req, err := http.NewRequest("POST", uri, payload)

	if err != nil {
		panic(err)
	}
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("Authorization", "Bearer yQ2AV4tUwXW7eGuNj3ZlahTM8EarMU-0WkzAx0qhzK4.ATl6WqD2c6v9mcP9g2VesSG51waJLwX08Zx6U-kWBBg")
	req.Header.Add("Cookie", "TLTSID=0491d9c3d8581663910a00e0ed96a2ca; NSC_uppmt-usvf-ofx=ffffffff3b2237bd45525d5f4f58455e445a4a4212d3")

	res, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		panic(err)
	}
	// fmt.Println(string(body))

	var parsed ValidityResponse
	err2 := json.Unmarshal(body, &parsed)
	if err2 != nil {
		panic(err)
	}
	fmt.Println(parsed.ResultStatus)
	return parsed.ResultStatus
}
