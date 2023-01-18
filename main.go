package main

import (
    "encoding/xml"
    "fmt"
    "net/http"
    "strconv"
    "time"
    "os"
    "encoding/json"
)

type RollcallVote struct {
    XMLName xml.Name `xml:"rollcall-vote"`
    Metadata struct {
        Majority       string    `xml:"majority"`
        Congress       int       `xml:"congress"`
        Session        string    `xml:"session"`
        Chamber        string    `xml:"chamber"`
        RollcallNum    int       `xml:"rollcall-num"`
        LegisNum       string    `xml:"legis-num"`
        VoteQuestion   string    `xml:"vote-question"`
        VoteType       string    `xml:"vote-type"`
        VoteResult     string    `xml:"vote-result"`
        ActionDate     string    `xml:"action-date"` // convert date and time to time.Time
        ActionTime     string    `xml:"action-time"`
        VoteDesc       string    `xml:"vote-desc"`
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
                YeaTotal         int    `xml:"yea-total"`
                NayTotal         int    `xml:"nay-total"`
                PresentTotal     int    `xml:"present-total"`
                NotVotingTotal   int    `xml:"not-voting-total"`
            } `xml:"totals-by-party"`
            TotalsByVote struct {
                TotalStub       string `xml:"total-stub"`
                YeaTotal         int    `xml:"yea-total"`
                NayTotal         int    `xml:"nay-total"`
                PresentTotal     int    `xml:"present-total"`
                NotVotingTotal   int    `xml:"not-voting-total"`
            } `xml:"totals-by-vote"`
        } `xml:"vote-totals"`
    } `xml:"vote-metadata"`
}


func printToLog(source string, output string) {
  fmt.Printf("%v: %v", source, output)
}

/*
The code uses the os.OpenFile function to open the file with the following flags :
    os.O_APPEND: opens the file in append mode
    os.O_CREATE: creates the file if it doesn't exist
    os.O_WRONLY: opens the file for writing only
*/
func writeStructToJSON(fileName string, data RollcallVote) error {
    file, err := os.OpenFile(fileName, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
    if err != nil {
        return err
    }
    defer file.Close()

    bytes, err := json.Marshal(data)
    if err != nil {
        return err
    }

    _, err = file.Write(bytes)
    if err != nil {
        return err
    }

    return nil
}

// getResponseHead Sends a head HTTP request to send url
// (faster then get for finding highest possible value)
func getResponseHead(url string) (*http.Response, error) {
  resp, err := http.Head(url)
	if err != nil {
		return nil, fmt.Errorf("Error making HTTP request: %v", err)
	}
  defer resp.Body.Close()
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

func appendResultToJson(rollCall RollcallVote) {
  
}

func getVoteResults(year int, rollCall int, appendToFile bool) (RollcallVote, error) {
  url := "https://clerk.house.gov/evs/2023/roll031.xml"
    rollcall, err := ParseXML(url)
    if err != nil {
        fmt.Println(err)
        return nil, err
    }
    if appendToFile {
      appendResultToJson(rollcall)
    }
    return rollcall, nil
}

/* getYearMaximum Attempts to find the highest roll in the year passed
 returns the roll call maximum possible value
*/
func GetYearMaximum(year int) (int, error) {
  roll := 0
  failedAttempts := 0

  for failedAttempts < 3 {

    rollStr := strconv.Itoa(roll)
    rollStr = fmt.Sprintf("%03d", roll)
    url := fmt.Sprintf("https://clerk.house.gov/evs/%d/roll%s.xml", year, rollStr)
    resp, err := getResponseHead(url)
    if err != nil {
      fmt.Println(err)
    }

    // If content body is html, mark as failed request (html means roll call doesn't exist)
    if resp.Header.Get("Content-Type") == "text/html" {
      failedAttempts++
    }
    roll++
  }

  // Remove 3 results because non existent results were counted to total
  return roll-3, nil
}

func parseResultsForYear(year int, fileName) error {
  highestRollCall := GetYearMaximum(year)
  for i := 1; i <= highestRollCall; i++ {
    
  }
}

/* getPossibleResultTotal takes a before and after year, gets all possible
results and returns an int or if an error; 0 int and error code
*/
func getPossibleResultTotal(beforeYear int, afterYear int) (int, error) {
  totalPossibleRequests := 0
  for selectedYear := afterYear; selectedYear >= beforeYear; selectedYear -= 1 {
    err := GetYearMaximum(selectedYear)
    if err != nil {
      return 0, err
    }
    totalPossibleRequests += result
  }
  return totalPossibleRequests, nil
}

/*
getResultsBetweenYears takes two different years, and gets all the results
in between those years including the years passed
IT IS ADVISED NOT TO PARSE MORE THEN 5 YEARS UNLESS YOU ARE SAVING IT TO A FILE
*/
func getResultsBetweenYears(beforeYear int, afterYear int, fileName string) (string, error) {
  totalPossibleRequests := 0
  
  for selectedYear := afterYear; selectedYear >= beforeYear; selectedYear -= 1 {
    err := parseResultsForYear(selectedYear)
    if err != nil {
      return "", err
    }
    totalPossibleRequests += result
  }
  return fmt.Sprintf("%v.json", fileName) nil
}

func main() {
    startTime := time.Now()
    fmt.Println(getPossibleResultTotal(2020,2025))
    endTime := time.Now()
    fmt.Println(endTime.Sub(startTime))
}