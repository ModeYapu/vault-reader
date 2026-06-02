package server

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func BenchmarkWriteJSON(b *testing.B) {
	data := map[string]interface{}{
		"title":   "Test Note",
		"content": "This is a test note with some content",
		"tags":    []string{"test", "benchmark"},
	}

	b.Run("Pooled", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			w := httptest.NewRecorder()
			writeJSON(w, http.StatusOK, data)
		}
	})

	b.Run("Standard", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			w := httptest.NewRecorder()
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(data)
		}
	})
}

func BenchmarkBufferPool(b *testing.B) {
	b.Run("Pooled", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			buf := getBuffer()
			buf.WriteString("test data")
			putBuffer(buf)
		}
	})

	b.Run("NewAllocation", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			buf := bytes.NewBuffer(make([]byte, 0, 1024))
			buf.WriteString("test data")
		}
	})
}

func BenchmarkResponseBuffer(b *testing.B) {
	data := []byte("test data for response buffer")

	b.Run("Pooled", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			rb := getResponseBuffer()
			rb.Write(data)
			_ = rb.Bytes()
			putResponseBuffer(rb)
		}
	})

	b.Run("NewAllocation", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			var buf bytes.Buffer
			buf.Write(data)
		}
	})
}

func BenchmarkCopyBufferPool(b *testing.B) {
	src := bytes.NewBuffer(make([]byte, 1024*1024)) // 1MB

	b.Run("Pooled", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			buf := copyBufPool.Get().([]byte)
			dst := io.Discard
			src.Reset()
			src.Write(make([]byte, 1024*1024))
			io.CopyBuffer(dst, src, buf)
			copyBufPool.Put(buf)
		}
	})

	b.Run("NewBuffer", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			buf := make([]byte, 32*1024)
			dst := io.Discard
			src.Reset()
			src.Write(make([]byte, 1024*1024))
			io.CopyBuffer(dst, src, buf)
		}
	})
}

func BenchmarkJSONWriter(b *testing.B) {
	data := map[string]interface{}{
		"title":       "Benchmark Test",
		"description": "Testing JSON writer performance",
		"items":       []int{1, 2, 3, 4, 5},
	}

	b.Run("PooledWriter", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			var buf bytes.Buffer
			jw := newJSONWriter(&buf)
			_ = jw.Write(data)
			jw.Close()
		}
	})

	b.Run("DirectEncoding", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			var buf bytes.Buffer
			_ = json.NewEncoder(&buf).Encode(data)
		}
	})
}
