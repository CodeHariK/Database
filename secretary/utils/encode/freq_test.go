package encode

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"testing"

	"github.com/codeharik/secretary/utils"
)

var (
	SECKEYMAP  []SECKEY
	ASCIIINDEX [256]byte
)

type FreqType int

const (
	SNo FreqType = 0
	S16          = 16
	S32          = 32
	S64          = 64
)

const (
	freq   FreqType = S32
	SERVER          = false
)

func TestFreq(t *testing.T) {
	if freq == S16 {
		SECKEYMAP = SEC16KeyMap[:]
		ASCIIINDEX = ASCII16Index
	} else if freq == S32 {
		SECKEYMAP = SEC32KeyMap[:]
		ASCIIINDEX = ASCII32Index
	} else if freq == S64 {
		SECKEYMAP = SEC64KeyMap[:]
		ASCIIINDEX = ASCII64Index
	}

	sortedMap()

	if SERVER {
		http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte(html))
		})

		http.HandleFunc("/frequency", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(sortedMap())
		})

		fmt.Println("Server running at http://localhost:8080")
		http.ListenAndServe(":8080", nil)
	}
}

// Count frequencies with phonetic grouping
func countASCII(root string) ([]int, []int) {
	groupfreq := make([]int, freq)
	charfreq := make([]int, 256)

	_ = filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			fmt.Println("Error:", err)
			return nil
		}
		if info.IsDir() {
			return nil
		}

		file, err := os.Open(path)
		if err != nil {
			fmt.Println("Error opening file:", err)
			return nil
		}
		defer file.Close()

		buf := make([]byte, 4096)
		for {
			n, err := file.Read(buf)
			if err != nil && err != io.EOF {
				break
			}
			if n == 0 {
				break
			}
			for _, char := range buf[:n] {
				if char < 128 { // ASCII only

					group := ASCIIINDEX[char]
					groupfreq[group]++

					charfreq[char]++
				}
			}
		}
		return nil
	})

	return groupfreq, charfreq
}

func sortedMap() map[string]int {
	root := "../../"
	groupfreq, charfreq := countASCII(root)

	// Convert map to slice for sorting
	type kv struct {
		Char  string
		Count int
	}
	var sortedFreq []kv

	totalChar := 0

	for _, groupcount := range groupfreq {
		totalChar += groupcount
	}

	for groupkey, groupcount := range groupfreq {

		mmm := utils.Map(SECKEYMAP[groupkey].keys, func(a byte) string {
			return fmt.Sprintf("%q,%d,%d", a, 100*charfreq[a]/groupcount, 100*charfreq[a]/totalChar)
		})

		sortedFreq = append(sortedFreq,
			kv{
				Char:  fmt.Sprintf("%8d %4d %8v", groupcount, groupkey, mmm),
				Count: groupcount,
			})
	}

	// Sort by frequency (descending)
	sort.Slice(sortedFreq, func(i, j int) bool {
		return sortedFreq[i].Count > sortedFreq[j].Count
	})

	// Convert back to a sorted map for JSON response
	sortedMap := make(map[string]int)

	charCount := 0
	for i, kv := range sortedFreq {
		sortedMap[kv.Char] = kv.Count

		charCount += kv.Count

		fmt.Printf("%4d %4d %4d %10v \n", i, 100*charCount/totalChar, 100*kv.Count/totalChar, kv.Char)
	}
	return sortedMap
}

var html = `<!DOCTYPE html>
<html lang="en">

<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Character Frequency</title>
    <script src="https://cdn.jsdelivr.net/npm/chart.js"></script>
</head>

<body>
    <canvas id="charChart"></canvas>
    <script>
        async function fetchData() {
            const response = await fetch('/frequency');
            const data = await response.json();
            const labels = Object.keys(data).map(ch => ch === " " ? "Space" : ch);
            const values = Object.values(data);

            const ctx = document.getElementById('charChart').getContext('2d');
            new Chart(ctx, {
                type: 'bar',
                data: {
                    labels: labels,
                    datasets: [{
                        label: 'Character Frequency',
                        data: values,
                        backgroundColor: 'rgba(228, 5, 20, 0.6)',
                        borderColor: 'rgb(255, 255, 255)',
                        borderWidth: 1
                    }]
                },
                options: {
                    scales: {
                        y: { beginAtZero: true }
                    }
                }
            });
        }
        fetchData();
    </script>
</body>

</html>`
