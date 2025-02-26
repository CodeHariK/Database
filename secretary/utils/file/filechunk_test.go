package file

import (
	"encoding/base64"
	"fmt"
	"hash/crc32"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/codeharik/secretary/utils/sec64"
)

// Helper function to create a temp file with random data
func createTempFile(filePath string) (string, error) {
	file, err := os.Create(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	fmt.Println(file.Name())

	text := `Once upon a time, in a distant galaxy far beyond the reach of human telescopes, there existed a vibrant planet named Zyloria. This planet was home to a unique species of aliens known as the Zylorians. They were small, luminescent beings with elongated limbs and large, expressive eyes that shimmered like stars. The Zylorians communicated through a series of melodic sounds and colors that changed with their emotions.
One day, a young Zylorian named Luma was exploring the lush, bioluminescent forests of her home. Luma was curious and adventurous, often wandering farther than her friends dared to go. As she ventured deeper into the woods, she stumbled upon a hidden glade filled with strange, glowing crystals. Intrigued, she reached out to touch one, and as her fingers brushed against its surface, a brilliant light enveloped her.
When the light faded, Luma found herself in an unfamiliar place. She was no longer on Zyloria but aboard a massive spaceship. The walls were sleek and metallic, and strange machines hummed softly around her. Confused but excited, Luma began to explore her new surroundings.
As she wandered through the ship, she encountered a group of beings unlike any she had ever seen. They were tall and slender, with skin that shimmered like the night sky. They introduced themselves as the Celestians, intergalactic travelers who had come to explore the universe. They were fascinated by Luma and her vibrant colors, which changed rapidly as she expressed her emotions.
The Celestians explained that they had been drawn to the energy of the crystals and had accidentally transported Luma aboard their ship. Rather than being frightened, Luma felt a sense of wonder. She shared stories of Zyloria, its beautiful landscapes, and the harmonious way of life of her people. In return, the Celestians told her about their adventures across the cosmos, visiting planets filled with strange creatures and breathtaking sights.
As days turned into weeks, Luma formed a deep bond with the Celestians. They taught her about their technology and the wonders of the universe, while she shared the beauty of her home planet. Together, they explored the ship, discovering new worlds and experiencing the thrill of space travel.
However, Luma began to miss her home. She longed to share her newfound knowledge and experiences with her friends and family. Sensing her feelings, the Celestians decided to help her return to Zyloria. They used their advanced technology to create a portal that would take Luma back to her glade.
With a heavy heart, Luma said goodbye to her new friends, promising to carry their stories with her. As she stepped through the portal, she felt a rush of emotions—excitement, sadness, and gratitude. When she emerged on the other side, she found herself back in the glade, the glowing crystals still shimmering around her.
Luma rushed home, eager to share her incredible adventure with her fellow Zylorians. She spoke of the Celestians, the wonders of space, and the importance of friendship and exploration. Inspired by her tales, the Zylorians began to dream of their own adventures beyond the stars.
From that day on, Luma became a storyteller, sharing her experiences and encouraging her people to embrace curiosity and the unknown. And though she never saw the Celestians again, she knew that the universe was vast and full of possibilities, waiting for those brave enough to explore it.`

	n, err := file.Write([]byte(text))
	if err != nil {
		return "", err
	}
	file.Write([]byte(fmt.Sprintf("\n--->Text %d\n", n)))

	n, err = file.Write([]byte(sec64.AsciiToSec64Expand(text)))
	if err != nil {
		return "", err
	}

	file.Write([]byte(fmt.Sprintf("\n--->Sec64Expand %d\n", n)))

	n, err = file.Write([]byte(sec64.StringToSec64(text)))
	if err != nil {
		return "", err
	}
	file.Write([]byte(fmt.Sprintf("\n--->Sec64 %d\n", n)))

	n, _ = base64.NewEncoder(base64.StdEncoding, file).Write([]byte(text))
	file.Write([]byte(fmt.Sprintf("\n--->Base64 %d\n", n)))

	return file.Name(), nil
}

// Helper function to compute CRC32 hash of a file
func getFileCRC32(filename string) (uint32, error) {
	file, err := os.Open(filename)
	if err != nil {
		return 0, err
	}
	defer file.Close()

	hash := crc32.NewIEEE()
	_, err = io.Copy(hash, file)
	if err != nil {
		return 0, err
	}

	return hash.Sum32(), nil
}

// Test function for split and merge
func TestSplitAndMerge(t *testing.T) {
	dir := "../../SECRETARY/Chunks"

	os.RemoveAll(dir)

	err := EnsureDir(dir)
	if err != nil {
		t.Fatal(err)
	}

	// Create a temp file with random data
	originalFile, err := createTempFile(filepath.Join(dir, "tempfile.bin"))
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	// defer os.Remove(originalFile)

	// Create a temporary directory for chunks
	metaFile := filepath.Join(dir, "metadata")
	// defer os.RemoveAll(metaFile) // Cleanup after test

	fmt.Println(metaFile)

	// Split file into chunks
	if err := splitFile(originalFile, metaFile); err != nil {
		t.Fatalf("File splitting failed: %v", err)
	}

	// Create output file path
	reconstructedFile := filepath.Join(dir, "reconstructed.bin")
	// defer os.Remove(reconstructedFile)

	// Merge chunks back
	if err := mergeChunks(metaFile, reconstructedFile); err != nil {
		t.Fatalf("File merging failed: %v", err)
	}

	// Compare original and reconstructed file hashes
	originalHash, _ := getFileCRC32(originalFile)
	reconstructedHash, _ := getFileCRC32(reconstructedFile)

	if originalHash != reconstructedHash {
		t.Fatalf("Hash mismatch! Expected %08x, got %08x", originalHash, reconstructedHash)
	}

	t.Logf("Test passed! Original and reconstructed files match. Hash: %08x", originalHash)
}
