package main

import (
	"bytes"
	"encoding/json"
	"os"
	"testing"
	"time"

	jsonv2 "github.com/go-json-experiment/json"

	"github.com/bytedance/sonic"
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

func BenchmarkChatHTML(b *testing.B) {
	var buffer bytes.Buffer
	buffer.Grow(8 * 1024)
	start := time.Now()
	ctx := b.Context()
	for b.Loop() {
		buffer.Reset()
		if err := template.FragSSEPatchFatMorphContent(
			1, start, start, dataFatMorph2,
		).Render(ctx, &buffer); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkChatJSONStd(b *testing.B) {
	var buffer bytes.Buffer
	buffer.Grow(8 * 1024)
	for b.Loop() {
		buffer.Reset()
		err := json.NewEncoder(&buffer).Encode(dataFatMorph2)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkChatJSONStdV2(b *testing.B) {
	var buffer bytes.Buffer
	buffer.Grow(8 * 1024)
	for b.Loop() {
		buffer.Reset()
		err := jsonv2.MarshalWrite(&buffer, dataFatMorph2)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkChatJSONSonic(b *testing.B) {
	var buffer bytes.Buffer
	buffer.Grow(8 * 1024)
	for b.Loop() {
		buffer.Reset()
		err := sonic.ConfigFastest.NewEncoder(&buffer).Encode(dataFatMorph2)
		if err != nil {
			b.Fatal(err)
		}
	}
}
