package main

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/romshark/datastar-bench/template"
)

func TestFatMorphDataJSON(t *testing.T) {
	{
		var data template.DataFatMorph
		fc, err := os.ReadFile("./chat-resp-1.json")
		if err != nil {
			t.Fatal("reading file:", err)
		}
		err = json.Unmarshal(fc, &data)
		if err != nil {
			t.Fatal("unexpected JSON unmarshaling error:", err)
		}
	}

	{
		var data template.DataFatMorph
		fc, err := os.ReadFile("./chat-resp-2.json")
		if err != nil {
			t.Fatal("reading file:", err)
		}
		err = json.Unmarshal(fc, &data)
		if err != nil {
			t.Fatal("unexpected JSON unmarshaling error:", err)
		}
	}
}
