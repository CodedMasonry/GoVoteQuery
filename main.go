package main

import (
    "encoding/xml"
    "fmt"
    "net/http"
    "strconv"
    "time"
    "os"
    "io"
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



// getResponseHead: Sends a head HTTP request to the provided URL and returns the response and an error if one occurred. The function utilizes the http.Head method to send the request and fmt.Errorf to format an error message if one occurred. The response body is closed after reading.
func getResponseHead(url string) (*http.Response, error) {
  resp, err := http.Head(url)
	if err != nil {
		return nil, fmt.Errorf("Error making HTTP request: %v", err)
	}
  defer resp.Body.Close()
	return resp, nil
}

// GetRollcallVote: Sends an HTTP GET request to the provided URL, reads the response body, and parses the XML data. The function utilizes the http.Get method to send the request, io.ReadAll to read the response body, and xml.Unmarshal to parse the XML data. The response body is closed after reading. Returns a pointer to a RollcallVote struct and an error if one occurred.
func GetRollcallVote(url string) (*RollcallVote, error) {
    resp, err := http.Get(url)
    if err != nil {
        return nil, fmt.Errorf("Error sending HTTP request: %v", err)
    }
    defer resp.Body.Close()

    xmlData, err := io.ReadAll(resp.Body)
    if err != nil {
        return nil, fmt.Errorf("Error reading response body: %v", err)
    }

    var rollcall RollcallVote
    if err := xml.Unmarshal(xmlData, &rollcall); err != nil {
        return nil, fmt.Errorf("Error parsing XML: %v", err)
    }

    return &rollcall, nil
}

// AppendToJSONFile: Appends a JSON representation of the provided RollcallVote struct to the specified file path. The function utilizes the os.OpenFile method to open or create the file, json.NewEncoder to encode the struct as JSON and io.WriteString to write the encoded JSON to the file. The file is closed after writing. Returns an error if one occurred.
func AppendToJSONFile(filePath string, rollcall *RollcallVote) error {
    // Open the file or create it if it doesn't exist
    file, err := os.OpenFile(filePath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
    if err != nil {
        return fmt.Errorf("Error opening file: %v", err)
    }
    defer file.Close()

    // Append rollcall to results array
    if _, err := file.Seek(0, io.SeekEnd); err != nil {
        return fmt.Errorf("Error seeking end of file: %v", err)
    }

    // check if the file is empty
    fi, _ := file.Stat()
    if fi.Size() != 0 {
        if _, err := file.Write([]byte(",")); err != nil {
            return fmt.Errorf("Error writing to file: %v", err)
        }
    }

    // Encode rollcall as JSON
    encoder := json.NewEncoder(file)
    if err := encoder.Encode(rollcall); err != nil {
        return fmt.Errorf("Error encoding struct to JSON: %v", err)
    }
    return nil
}





func getVoteResults(year int, roll int, appendToFile bool, fileName string) (*RollcallVote, error) {
    rollStr := strconv.Itoa(roll)
    rollStr = fmt.Sprintf("%03d", roll)
    url := fmt.Sprintf("https://clerk.house.gov/evs/%d/roll%s.xml", year, rollStr)

    rollcall, err := GetRollcallVote(url)
    if err != nil {
        fmt.Println(err)
        return nil, err
    }
    if appendToFile {
      err = AppendToJSONFile(fileName, rollcall)
      if err != nil {
        return nil, err
      }
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

func parseResultsForYear(year int, fileName string) error {
  highestRollCall, err := GetYearMaximum(year)
  if err != nil {
    return err
  }
  for roll := 1; roll <= highestRollCall; roll++ {
    _, err = getVoteResults(year, roll, true, fileName)
    if err != nil {
      return err
    }
  }
  return nil
}

/* getPossibleResultTotal takes a before and after year, gets all possible
results and returns an int or if an error; 0 int and error code
*/
func getPossibleResultTotal(beforeYear int, afterYear int) (int, error) {
  totalPossibleRequests := 0
  for selectedYear := afterYear; selectedYear >= beforeYear; selectedYear -= 1 {
    result, err := GetYearMaximum(selectedYear)
    if err != nil {
      return 0, err
    }
    
    fmt.Printf("[getPossibleResultTotal] %v\n", selectedYear)
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
  
  for selectedYear := afterYear; selectedYear >= beforeYear; selectedYear -= 1 {
    err := parseResultsForYear(selectedYear, fileName)
    if err != nil {
      return "", err
    }
    fmt.Printf("[getResultsBetweenYears] %v\n", selectedYear)
  }
  return fmt.Sprintf("%v.json", fileName), nil
}

func main() {
    startTime := time.Now()
    //fmt.Println(getPossibleResultTotal(2000,2025))
    endTime := time.Now()
    fmt.Println("Time Taken to calculate total: ", endTime.Sub(startTime))

    startTime = time.Now()
    fmt.Println(getResultsBetweenYears(2000,2025,"results.json"))
    endTime = time.Now()
    fmt.Println("Time taken to parse files: ", endTime.Sub(startTime))
}