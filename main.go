package main

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"log"
	"math/big"
	"net/http"
	"os"
	"time"
)

type TranslationResponse struct {
	Translations []struct {
		Text string `json:"text"`
	} `json:"translations"`
}

type TranslationRequest struct {
	Text string `json:"text"`
}

type Translation struct {
	Source      string `xml:"source"`
	Translation string `xml:"translation"`
}

type Context struct {
	Name      string        `xml:"name"`
	TransList []Translation `xml:"message"`
}

type TS struct {
	ContextList []Context `xml:"context"`
}

func main() {
	start := time.Now()
	r := new(big.Int)
	fmt.Println(r.Binomial(1000, 10))
	// Set your subscription key and endpoint.
	subscriptionKey := os.Getenv("AZURE_TRANSLATION_KEY")
	endpoint := "https://api.cognitive.microsofttranslator.com"

	// Set the target language.
	targetLanguage := "pt"

	// Read the XML file.
	xmlFile, err := os.Open("pt_PT-ui.xml")
	if err != nil {
		fmt.Println("Error opening file:", err)
		return
	}
	defer xmlFile.Close()

	// Parse the XML file.
	byteValue, _ := io.ReadAll(xmlFile)
	var ts TS
	xml.Unmarshal(byteValue, &ts)

	// Iterate through the translations and translate the source text.
	for i := range ts.ContextList {
		for j := range ts.ContextList[i].TransList {
			// Translate the text.
			m := ts.ContextList[i].TransList[j]
			// if len(m.Source) == 0 {
			// 	println("EEORROEOEOEO")
			// }
			//fmt.Printf("m %s", m)
			//fmt.Printf("text %s", m.Source)
			requestBody := []TranslationRequest{{Text: m.Source}}
			//fmt.Printf("requestBody %s", requestBody)

			requestBodyBytes := new(bytes.Buffer)
			json.NewEncoder(requestBodyBytes).Encode(requestBody)
			request, _ := http.NewRequest("POST", fmt.Sprintf("%s/translate?api-version=3.0&to=%s", endpoint, targetLanguage), requestBodyBytes)
			request.Header.Set("Content-Type", "application/json")
			request.Header.Set("Ocp-Apim-Subscription-Key", subscriptionKey)
			request.Header.Set("Ocp-Apim-Subscription-Region", "westeurope")

			client := &http.Client{}
			response, err := client.Do(request)
			if err != nil {
				fmt.Println("Error translating text:", err)
				return
			}
			defer response.Body.Close()

			// Read the response.
			responseBody, _ := io.ReadAll(response.Body)
			var translationResponse []TranslationResponse
			json.Unmarshal(responseBody, &translationResponse)
			if len(translationResponse) > 0 && len(translationResponse[0].Translations) > 0 {
				ts.ContextList[i].TransList[j].Translation = translationResponse[0].Translations[0].Text
			} else if len(translationResponse) == 0 {
				fmt.Println("Error: empty translation response")
			} else {
				fmt.Println("Error: no translations found")
			}

			// Update the translation.
			ts.ContextList[i].TransList[j].Translation = translationResponse[0].Translations[0].Text
		}
	}

	// Write the updated XML file.
	output, _ := xml.MarshalIndent(ts, "", "    ")
	output = []byte(xml.Header + string(output))
	err = os.WriteFile("file_pt_PT_translated.xml", output, 0644)
	if err != nil {
		fmt.Println("Error writing file:", err)
		return
	}

	fmt.Println("Translation complete.")
	elapsed := time.Since(start)
	log.Printf("Translation took %s seconds", elapsed)
}
