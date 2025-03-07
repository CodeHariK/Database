package k7

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// HTML content as a string
const httpHtmlContent = `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Benchmark Results</title>
    <script src="https://cdn.jsdelivr.net/npm/chart.js"></script>
    <style>
        body { font-family: Arial, sans-serif; text-align: center; }
        .grid-container { display: grid; grid-template-columns: 1fr 1fr 1fr; gap: 20px; padding: 20px; }
        .chart-container { width: 100%; height: 300px; }
        canvas { background: #f9f9f9; border-radius: 8px; }
    </style>
</head>
<body>
    <h2>Benchmark Results</h2>
    <div class="grid-container">
        <div class="chart-container"><canvas id="requestsChart"></canvas></div>
        <div class="chart-container"><canvas id="latencyChart"></canvas></div>
        <div class="chart-container"><canvas id="cpuChart"></canvas></div>
        <div class="chart-container"><canvas id="memoryChart"></canvas></div>
        <div class="chart-container"><canvas id="netSentChart"></canvas></div>
        <div class="chart-container"><canvas id="netRecvChart"></canvas></div>
    </div>

    <script>
        async function loadBenchmarkData() {
            const response = await fetch("/results");
            const data = await response.json();

            const labels = data.map((_, index) => "Sec " + index);
            const success = data.map(item => item.Success);
            const errors = data.map(item => item.Errors);
            const requests = data.map(item => item.Requests);
            const avgLatency = data.map(item => item.Requests > 0 ? item.TotalLatency / (item.Requests*1000) : 0);
            const cpuUsage = data.map(item => item.CPUUsage);
            const memUsage = data.map(item => item.MemUsage);
            const netSent = data.map(item => item.NetSent / 1024); // Convert to KB
            const netRecv = data.map(item => item.NetRecv / 1024);

            function createChart(canvasId, datasets) {
                new Chart(document.getElementById(canvasId), {
                    type: "line",
                    data: {
                        labels: labels,
                        datasets: datasets
                    },
                    options: { 
                        responsive: true, 
                        scales: { y: { beginAtZero: true } } 
                    }
                });
            }

            // Combine Requests, Success, and Errors into one chart
            createChart("requestsChart", [
                { label: "Requests/sec", data: requests, borderColor: "blue", fill: false },
                { label: "Success/sec", data: success, borderColor: "green", fill: false },
                { label: "Errors/sec", data: errors, borderColor: "red", fill: false }
            ]);

            createChart("latencyChart", [
                { label: "Avg Latency (ms)", data: avgLatency, borderColor: "red", fill: false }
            ]);
            createChart("cpuChart", [
                { label: "CPU Usage (%)", data: cpuUsage, borderColor: "green", fill: false }
            ]);
            createChart("memoryChart", [
                { label: "Memory Usage (%)", data: memUsage, borderColor: "purple", fill: false }
            ]);
            createChart("netSentChart", [
                { label: "Network Sent (KB)", data: netSent, borderColor: "orange", fill: false }
            ]);
            createChart("netRecvChart", [
                { label: "Network Received (KB)", data: netRecv, borderColor: "brown", fill: false }
            ]);
        }

        loadBenchmarkData();
    </script>
</body>
</html>`

const wsHtmlContent = `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Benchmark Results</title>
    <script src="https://cdn.jsdelivr.net/npm/chart.js"></script>
    <style>
        body { font-family: Arial, sans-serif; text-align: center; }
        .grid-container { display: grid; grid-template-columns: 1fr 1fr 1fr; gap: 20px; padding: 20px; }
        .chart-container { width: 100%; height: 300px; }
        canvas { background: #f9f9f9; border-radius: 8px; }
    </style>
</head>
<body>
    <h2>Benchmark Results (Live)</h2>
    <div class="grid-container">
        <div class="chart-container"><canvas id="requestsChart"></canvas></div>
        <div class="chart-container"><canvas id="latencyChart"></canvas></div>
        <div class="chart-container"><canvas id="cpuChart"></canvas></div>
        <div class="chart-container"><canvas id="memoryChart"></canvas></div>
        <div class="chart-container"><canvas id="netSentChart"></canvas></div>
        <div class="chart-container"><canvas id="netRecvChart"></canvas></div>
    </div>

    <script>
        const socket = new WebSocket("ws://localhost:8888/ws");

        const maxPoints = 100; // Store last 60,000ms (100ms intervals)
        let labels = Array.from({ length: maxPoints }, (_, i) => (i * 100)+"ms");

        function createChart(canvasId, datasets) {
            return new Chart(document.getElementById(canvasId), {
                type: "line",
                data: {
                    labels: labels,
                    datasets: datasets
                },
                options: {
                    responsive: true,
                    scales: {
                        x: { 
                            type: "linear", 
                            position: "bottom",
                            ticks: { autoSkip: true, maxTicksLimit: 20 } 
                        },
                        y: { beginAtZero: true }
                    },
                    animation: false, // Real-time performance optimization
                }
            });
        }

        // Create charts
        const requestsChart = createChart("requestsChart", [
            { label: "Requests/sec", data: [], borderColor: "blue", fill: false },
            { label: "Success/sec", data: [], borderColor: "green", fill: false },
            { label: "Errors/sec", data: [], borderColor: "red", fill: false }
        ]);
        const latencyChart = createChart("latencyChart", [
            { label: "Avg Latency (ms)", data: [], borderColor: "red", fill: false }
        ]);
        const cpuChart = createChart("cpuChart", [
            { label: "CPU Usage (%)", data: [], borderColor: "green", fill: false }
        ]);
        const memoryChart = createChart("memoryChart", [
            { label: "Memory Usage (%)", data: [], borderColor: "purple", fill: false }
        ]);
        const netSentChart = createChart("netSentChart", [
            { label: "Network Sent (KB)", data: [], borderColor: "orange", fill: false }
        ]);
        const netRecvChart = createChart("netRecvChart", [
            { label: "Network Received (KB)", data: [], borderColor: "brown", fill: false }
        ]);

        function updateChart(chart, newData, datasetIndex = 0) {
            const dataset = chart.data.datasets[datasetIndex];
            dataset.data.push({ x: Date.now() % 60000, y: newData });

            if (dataset.data.length > maxPoints) {
                dataset.data.shift();
            }

            chart.update();
        }

        socket.onmessage = function(event) {
            const data = JSON.parse(event.data);

            updateChart(requestsChart, data.Requests, 0);
            updateChart(requestsChart, data.Success, 1);
            updateChart(requestsChart, data.Errors, 2);
            updateChart(latencyChart, data.Requests > 0 ? data.TotalLatency / (data.Requests * 1000) : 0);
            updateChart(cpuChart, data.CPUUsage);
            updateChart(memoryChart, data.MemUsage);
            updateChart(netSentChart, data.NetSent / 1024);
            updateChart(netRecvChart, data.NetRecv / 1024);
        };

        socket.onclose = function() {
            console.log("WebSocket closed.");
        };
    </script>
</body>
</html>`

// Serve the HTML page
func httpServeHTML(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	fmt.Fprint(w, httpHtmlContent)
}

// Serve the HTML page
func wsServeHTML(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	fmt.Fprint(w, wsHtmlContent)
}

// Serve benchmark results as JSON
func (config BenchmarkConfig) serveResults(w http.ResponseWriter, r *http.Request) {
	buckets := RunBenchmark(config)

	data, err := json.MarshalIndent(buckets, "", " ")
	if err != nil {
		fmt.Println("Error encoding JSON:", err)
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(data)
}
