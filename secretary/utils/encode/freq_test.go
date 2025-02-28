package encode

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"unicode"
)

// Count frequencies with phonetic grouping
func countASCII(root string) map[byte]int {
	freq := make(map[byte]int)

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
			for _, b := range buf[:n] {
				if b < 128 { // ASCII only
					char := unicode.ToLower(rune(b))

					freq[ASCII32Index[char]]++
				}
			}
		}
		return nil
	})

	return freq
}

func sortedMap() map[string]int {
	root := "../../"
	frequencies := countASCII(root)

	// Convert map to slice for sorting
	type kv struct {
		Char  string
		Count int
	}
	var sortedFreq []kv

	totalChar := 0

	for k, v := range frequencies {
		sortedFreq = append(sortedFreq, kv{Char: fmt.Sprintf("%2d %16q", k, string(SEC32KeyMap[k])), Count: v})
		totalChar += v
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

		fmt.Printf("%4d %7v %9v %4d %4d\n", i, kv.Char, kv.Count, 100*charCount/totalChar, 100*kv.Count/totalChar)
	}
	return sortedMap
}

func frequencyHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(sortedMap())
}

// func TestMain(t *testing.T) {
// 	sortedMap()

// 	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
// 		w.Write([]byte(html))
// 	}) // API endpoint
// 	http.HandleFunc("/frequency", frequencyHandler) // API endpoint

// 	fmt.Println("Server running at http://localhost:8080")
// 	http.ListenAndServe(":8080", nil)
// }

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

/*

0    ' '    32   224434      9
1    ','    44   115891     15
2    'e'   101    90954     19
3    '0'    48    88983     23
4   '\n'    10    79910     26
5    't'   116    77762     30
6    '2'    50    74192     33
7    'i'   105    68819     36
8   '\t'     9    65162     39
9    '3'    51    64568     42
10    's'   115    64344     45
11    'n'   110    61169     47
12    '1'    49    60069     50
13    'r'   114    56912     53
14    'a'    97    56274     55
15    'l'   108    48418     57
16    '5'    53    46883     59
17    'o'   111    46422     61
18    'c'    99    45621     64
19    '4'    52    44040     65
20    'd'   100    41524     67
21    '6'    54    40700     69
22    'f'   102    40551     71
23    '_'    95    37853     73
24    '8'    56    36760     74
25    '7'    55    35389     76
26    '9'    57    34599     77
27    'u'   117    30030     79
28    ')'    41    29131     80
29    '('    40    29079     81
30    'p'   112    28625     83
31    '='    61    26891     84
32    'x'   120    25945     85
33    ';'    59    25008     86
34    '*'    42    24708     87
35    'm'   109    23479     88
36    'h'   104    21306     89
37    '.'    46    17628     90
38    'b'    98    17628     91
39    'g'   103    16661     91
40    '-'    45    16113     92
41 '\x00'     0    14173     93
42    'y'   121    11987     93
43    '/'    47    11130     94
44    '>'    62     9455     94
45    'v'   118     9052     95
46    'w'   119     9034     95
47    '+'    43     8586     95
48    '#'    35     7752     96
49    'k'   107     7298     96
50    '{'   123     7285     96
51    '}'   125     7281     97
52    '"'    34     5979     97
53    '<'    60     5758     97
54    'z'   122     5574     98
55   '\\'    92     5341     98
56    '['    91     5131     98
57    ']'    93     5129     98
58    '&'    38     4880     98
59    ':'    58     3563     99
60   '\''    39     3369     99
61    '|'   124     3135     99
62    '%'    37     3087     99
63    'q'   113     2212     99
64    '!'    33     2083     99
65    'j'   106     1850     99
66    '$'    36     1618     99
67    '?'    63      601     99
68    '^'    94      573     99
69    '@'    64      519     99
70    '~'   126      260     99
71 '\x01'     1      138     99
72   '\b'     8      133     99
73    '`'    96      123     99
74 '\x04'     4       71     99
75 '\x06'     6       58     99
76 '\x05'     5       40     99
77 '\x02'     2       34     99
78 '\x03'     3       32     99
79   '\a'     7       31     99
80 '\x10'    16       25     99
81   '\v'    11        9     99
82 '\x14'    20        4     99
83   '\f'    12        4     99
84   '\r'    13        3     99
85 '\x1d'    29        2     99
86 '\x1b'    27        2     99
87 '\x1f'    31        2     99
88 '\x1a'    26        2     99
89 '\x0f'    15        1     99
90 '\x15'    21        1     99
91 '\x18'    24        1     99
92 '\x11'    17        1     99
93 '\x16'    22        1     99
94 '\x19'    25        1    100




0    ' '    32 15827074     17
1    'e'   101  8521608     26
2    't'   116  5995059     33
3    'a'    97  5634185     39
4    'o'   111  4959336     44
5    'n'   110  4620222     49
6    'h'   104  4470892     54
7    'i'   105  4354779     59
8    's'   115  4164566     63
9    'r'   114  4047746     68
10    'd'   100  3386714     72
11    'l'   108  2931460     75
12    'u'   117  1862109     77
13    'w'   119  1662748     79
14    'm'   109  1623024     80
15    'g'   103  1581262     82
16    'c'    99  1525031     84
17    'f'   102  1410620     85
18    'y'   121  1278989     87
19 '\x00'     0  1242402     88
20    ','    44  1172181     89
21    '.'    46  1122071     91
22    'b'    98  1053480     92
23    'p'   112  1048503     93
24   '\n'    10   901311     94
25   '\r'    13   805623     95
26    'k'   107   687854     95
27    'v'   118   596802     96
28    '"'    34   589817     97
29   '\''    39   266935     97
30    '-'    45   146654     97
31    '_'    95   137890     97
32    'x'   120   125445     97
33    '1'    49   108131     98
34    'z'   122   103677     98
35    '0'    48    99303     98
36    '?'    63    93497     98
37 '\x01'     1    90709     98
38    '2'    50    86841     98
39    'j'   106    86222     98
40   '\t'     9    78048     98
41    '3'    51    70327     98
42    ':'    58    67016     98
43    ';'    59    61595     98
44    '!'    33    61315     99
45    'q'   113    55944     99
46    '5'    53    54857     99
47    '4'    52    51946     99
48    '6'    54    48144     99
49    '8'    56    44632     99
50    '7'    55    44525     99
51    '9'    57    43274     99
52    '('    40    37249     99
53    '='    61    36726     99
54    ')'    41    35620     99
55    '*'    42    33342     99
56    '>'    62    29395     99
57    '<'    60    25531     99
58 '\x02'     2    23263     99
59    '$'    36    21742     99
60 '\x05'     5    16554     99
61 '\x03'     3    16129     99
62    '/'    47    14509     99
63 '\x04'     4    14109     99
64   '\b'     8    12232     99
65 '\x13'    19    11834     99
66    '#'    35    10697     99
67    '+'    43    10663     99
68 '\x10'    16    10592     99
69    '@'    64     9604     99
70    '}'   125     9503     99
71    '{'   123     9050     99
72   '\\'    92     8428     99
73    ']'    93     8132     99
74 '\x12'    18     7844     99
75 '\x06'     6     7762     99
76    '['    91     7626     99
77    '&'    38     7145     99
78 '\x0e'    14     7035     99
79    '|'   124     6258     99
80   '\v'    11     6181     99
81 '\x0f'    15     5676     99
82    '%'    37     5664     99
83    '`'    96     5166     99
84   '\a'     7     4880     99
85    '^'    94     4647     99
86 '\x14'    20     4531     99
87 '\x18'    24     4347     99
88 '\x1c'    28     4162     99
89 '\x1a'    26     3624     99
90 '\x11'    17     3385     99
91 '\x19'    25     3336     99

*/
