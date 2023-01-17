package main

import (
    "encoding/xml"
    "fmt"
    "net/http"
    "strconv"
)



type RollcallVote struct {
    XMLName xml.Name `xml:"rollcall-vote"`
    Metadata struct {
        Majority       string `xml:"majority"`
        Congress       string `xml:"congress"`
        Session        string `xml:"session"`
        Chamber        string `xml:"chamber"`
        RollcallNum    string `xml:"rollcall-num"`
        LegisNum       string `xml:"legis-num"`
        VoteQuestion   string `xml:"vote-question"`
        VoteType       string `xml:"vote-type"`
        VoteResult     string `xml:"vote-result"`
        ActionDate     string `xml:"action-date"`
        ActionTime     string `xml:"action-time"`
        VoteDesc       string `xml:"vote-desc"`
        VoteTotals     struct {
            TotalsByPartyHeader struct {
                PartyHeader    string `xml:"party-header"`
                YeaHeader      string `xml:"yea-header"`
                NayHeader      string `xml:"nay-header"`
                PresentHeader  string `xml:"present-header"`
                NotVotingHeader string `xml:"not-voting-header"`
            } `xml:"totals-by-party-header"`
            TotalsByParty []struct {
                Party            string `xml:"party"`
                YeaTotal         string `xml:"yea-total"`
                NayTotal         string `xml:"nay-total"`
                PresentTotal     string `xml:"present-total"`
                NotVotingTotal   string `xml:"not-voting-total"`
            } `xml:"totals-by-party"`
            TotalsByVote struct {
                TotalStub       string `xml:"total-stub"`
                YeaTotal         string `xml:"yea-total"`
                NayTotal         string `xml:"nay-total"`
                PresentTotal     string `xml:"present-total"`
                NotVotingTotal   string `xml:"not-voting-total"`
            } `xml:"totals-by-vote"`
        } `xml:"vote-totals"`
    } `xml:"vote-metadata"`
}


func getResponse(url string) (*http.Response, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("Error making HTTP request: %v", err)
	}
	return resp, nil
}

// ParseXML sends a request to the url and applies the returned xml to RollcallVote struct
// returns RollcallVote struct, returns nil if error
func ParseXML(url string) (*RollcallVote, error) {
    resp, err := http.Get(url)
    if err != nil {
        return nil, fmt.Errorf("Error downloading XML: %v", err)
    }
    defer resp.Body.Close()

    var rollcall RollcallVote
    err = xml.NewDecoder(resp.Body).Decode(&rollcall)
    if err != nil {
        return nil, fmt.Errorf("Error decoding XML: %v", err)
    }

    return &rollcall, nil
}

// getYearMaximum Attempts to find the highest roll in the year passed
// returns the roll call maximum possible value
func getYearMaximum(year int) (int, error) {
  roll := 0
  failedAttempts := 0

  for failedAttempts < 3 {

    rollStr := strconv.Itoa(roll)
    rollStr = fmt.Sprintf("%03d", roll)
    url := fmt.Sprintf("https://clerk.house.gov/evs/%d/roll%s.xml", year, rollStr)
    resp, err := getResponse(url)
    if err != nil {
      fmt.Println(err)
    }

    // If content body is html, mark as failed request (html means roll call doesn't exist)
    if resp.Header.Get("Content-Type") == "text/html" {
      failedAttempts++
    }
    roll++
  }

  return roll, nil
}

func main() {
    url := "https://clerk.house.gov/evs/2023/roll031.xml"
    rollcall, err := ParseXML(url)
    if err != nil {
        fmt.Println(err)
        return
    }
    fmt.Println("Parsed RollcallVote struct:", rollcall)
    fmt.Println(getYearMaximum(2023))
}
