package eggwalker

import (
	"testing"
)

func TestGetContent(t *testing.T) {
	doc := createDoc()
	docGetcontent := getContent(doc)
	if docGetcontent != "" {
		t.Fatal("It should be empty, got", docGetcontent)
	}
}
