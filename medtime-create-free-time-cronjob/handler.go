package function

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"
)

// Handle a serverless request
func Handle(request []byte) string {
	url := "https://api.admin.u-code.io/v2/object/get-list/doctor"
	xApiKey := "9dfa92f1d53a"
	payload := map[string]interface{}{
		"data": map[string]interface{}{
			"row_view_id": "7fc98477-d729-4aac-bde1-9dfa92f1d53a",
			"offset":      0,
			"order":       map[string]interface{}{},
			"view_fields": []interface{}{},
			"search":      "",
			"limit":       20,
			"undefined":   "",
		},
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		log.Fatalf("Error marshaling payload: %v", err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		log.Fatalf("Error creating request: %v", err)
	}

	req.Header.Set("Authorization", xApiKey)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Fatalf("Error making request: %v", err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("Error reading response: %v", err)
	}

	var result struct {
		Data struct {
			Data struct {
				Response []struct {
					Guid string `json:"guid"`
				} `json:"response"`
			} `json:"data"`
		} `json:"data"`
	}

	if err := json.Unmarshal(body, &result); err != nil {
		log.Fatalf("Error unmarshaling response: %v", err)
	}

	var guids []string
	for _, item := range result.Data.Data.Response {
		guids = append(guids, item.Guid)
	}

	fmt.Println("GUIDs:", guids)

	bookingURL := "https://api.admin.u-code.io/v2/items/doctor_booking"

	for _, doctorID := range guids {
		for day := 0; day < 30; day++ {
			date := time.Now().AddDate(0, 0, day).Format("2006-01-02T15:04:05.000Z")
			for hour := 10; hour < 16; hour += 2 {
				fromTime := fmt.Sprintf("%02d:00", hour)
				toTime := fmt.Sprintf("%02d:00", hour+2)

				bookingPayload := map[string]interface{}{
					"data": map[string]interface{}{
						"doctor_id": doctorID,
						"finished":  false,
						"is_booked": false,
						"invite":    false,
						"date":      date,
						"from_time": fromTime,
						"to_time":   toTime,
					},
				}

				bookingData, err := json.Marshal(bookingPayload)
				if err != nil {
					log.Printf("Error marshaling booking payload: %v", err)
					continue
				}

				bookingReq, err := http.NewRequest("POST", bookingURL, bytes.NewBuffer(bookingData))
				if err != nil {
					log.Printf("Error creating booking request: %v", err)
					continue
				}

				bookingReq.Header.Set("Authorization", xApiKey)
				bookingReq.Header.Set("Content-Type", "application/json")

				bookingResp, err := client.Do(bookingReq)
				if err != nil {
					log.Printf("Error sending booking request: %v", err)
					continue
				}
				bookingResp.Body.Close()
			}
		}
	}
	return result.Data.Data.Response[0].Guid
}
