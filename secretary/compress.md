🚀 Zstd (Zstandard) in Golang

Zstandard (Zstd) is a fast compression algorithm that provides better compression ratios than gzip while being faster. It’s great for high-performance databases, logs, network transmission, and storage.

📌 Why Use Zstd in Golang?
	•	🔥 High compression ratio (better than gzip)
	•	⚡ Fast decompression (ideal for real-time applications)
	•	📉 Adjustable compression levels (trade-off between speed & compression)
	•	🔄 Streaming support (works well for large files)

📦 Install Zstd for Golang

Use the optimized Zstd package by Klaus Post:

go get github.com/klauspost/compress/zstd



⸻

✍️ Example: Compress & Decompress Using Zstd

package main

import (
	"bytes"
	"fmt"
	"io"
	"log"

	"github.com/klauspost/compress/zstd"
)

// Compresses input data using Zstd
func compress(data []byte) []byte {
	var buf bytes.Buffer
	writer, _ := zstd.NewWriter(&buf)
	writer.Write(data)
	writer.Close()
	return buf.Bytes()
}

// Decompresses Zstd compressed data
func decompress(compressed []byte) []byte {
	reader, _ := zstd.NewReader(bytes.NewReader(compressed))
	decompressed, _ := io.ReadAll(reader)
	return decompressed
}

func main() {
	data := []byte("Hello, this is a repeated text. Hello, this is a repeated text. Hello, this is a repeated text.")

	// Compress data
	compressed := compress(data)
	fmt.Println("Original Size:", len(data))
	fmt.Println("Compressed Size:", len(compressed))

	// Decompress data
	decompressed := decompress(compressed)
	fmt.Println("Decompressed:", string(decompressed))
}



⸻

📊 How Much Can Zstd Compress?

Zstd compression depends on the data type. Here’s a rough estimate:

Data Type	Compression Ratio
Text (logs, JSON)	70-85% smaller
Binary (executables, data)	15-40% smaller
Already Compressed Data (JPEG, MP3, MP4)	~0%

	•	Level 1 (fastest) → Lower compression, high speed.
	•	Level 22 (slowest) → Best compression, slow speed.
	•	Default level is 3, which is a good balance.

🚀 Streaming Compression (for Large Data)

If you’re dealing with large files, use streaming compression to avoid excessive memory usage:

func streamCompress(reader io.Reader, writer io.Writer) error {
	zstdWriter, _ := zstd.NewWriter(writer)
	_, err := io.Copy(zstdWriter, reader)
	zstdWriter.Close()
	return err
}

This is useful for compressing database logs, network data, and large files without loading everything into memory.

⸻

✅ Where Can Zstd Be Useful in Your Projects?
	•	Bitcask: Compress log segments to reduce disk usage.
	•	B+ Tree: Compress index nodes/pages for better storage efficiency.
	•	Kademlia (DHT): Compress large values to save bandwidth.
	•	Gemini Server: Compress responses for faster transmission.

Would you like a benchmark for your specific data in Go? 🚀

