package server

import (
	"bufio"
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"sync"
)

// bufferPool reuses byte buffers for JSON encoding and I/O operations.
var bufferPool = sync.Pool{
	New: func() interface{} {
		return new(bytes.Buffer)
	},
}

// getBuffer retrieves a buffer from the pool.
func getBuffer() *bytes.Buffer {
	return bufferPool.Get().(*bytes.Buffer)
}

// putBuffer returns a buffer to the pool after resetting it.
func putBuffer(b *bytes.Buffer) {
	b.Reset()
	bufferPool.Put(b)
}

// writeJSON writes JSON response using a pooled buffer for zero-allocation encoding.
func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	buf := getBuffer()
	defer putBuffer(buf)

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.WriteHeader(status)

	// Encode to buffer first for efficient zero-copy write
	if err := json.NewEncoder(buf).Encode(v); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Zero-copy write to response
	_, _ = buf.WriteTo(w)
}

// pooledReader wraps io.Reader for reuse.
type pooledReader struct {
	*bytes.Reader
}

// readerPool reuses bytes.Reader instances.
var readerPool = sync.Pool{
	New: func() interface{} {
		return &pooledReader{Reader: bytes.NewReader(nil)}
	},
}

// getPooledReader creates a reader from pooled bytes.Reader.
func getPooledReader(data []byte) io.Reader {
	pr := readerPool.Get().(*pooledReader)
	pr.Reset(data)
	return pr
}

// copyBuffer uses a pooled buffer for io.CopyBuffer operations.
var copyBufPool = sync.Pool{
	New: func() interface{} {
		return make([]byte, 32*1024) // 32KB buffer for efficient copying
	},
}

// zeroCopyFile serves a file using zero-copy techniques when possible.
func zeroCopyFile(w http.ResponseWriter, r *http.Request, path string, contentType string) {
	file, err := http.Dir(path).Open(path)
	if err != nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	defer file.Close()

	stat, err := file.Stat()
	if err != nil {
		http.Error(w, "error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", contentType)
	http.ServeContent(w, r, stat.Name(), stat.ModTime(), file)
}

// jsonWriter provides efficient JSON writing with pooling.
type jsonWriter struct {
	w    io.Writer
	enc  *json.Encoder
	buf  *bytes.Buffer
	once sync.Once
}

// newJSONWriter creates a new JSON writer with pooled resources.
func newJSONWriter(w io.Writer) *jsonWriter {
	buf := getBuffer()
	return &jsonWriter{
		w:   w,
		enc: json.NewEncoder(buf),
		buf: buf,
	}
}

// Write writes JSON data efficiently.
func (jw *jsonWriter) Write(v interface{}) error {
	if err := jw.enc.Encode(v); err != nil {
		return err
	}
	_, err := jw.buf.WriteTo(jw.w)
	return err
}

// Close releases resources back to the pool.
func (jw *jsonWriter) Close() {
	jw.once.Do(func() {
		putBuffer(jw.buf)
	})
}

// flushWriter wraps http.ResponseWriter for buffered writes.
type flushWriter struct {
	w   http.ResponseWriter
	buf *bufio.Writer
}

// newFlushWriter creates a buffered writer with automatic flushing.
func newFlushWriter(w http.ResponseWriter) *flushWriter {
	return &flushWriter{
		w:   w,
		buf: bufio.NewWriterSize(w, 4096),
	}
}

// Write writes data to the buffer.
func (fw *flushWriter) Write(p []byte) (int, error) {
	return fw.buf.Write(p)
}

// Flush flushes any buffered data.
func (fw *flushWriter) Flush() error {
	return fw.buf.Flush()
}

// responseBufferPool pools response buffers for high-throughput scenarios.
type responseBuffer struct {
	buf []byte
	sz  int
}

var responseBufferPool = sync.Pool{
	New: func() interface{} {
		return &responseBuffer{
			buf: make([]byte, 4096),
			sz:  0,
		}
	},
}

// getResponseBuffer retrieves a response buffer.
func getResponseBuffer() *responseBuffer {
	return responseBufferPool.Get().(*responseBuffer)
}

// putResponseBuffer returns a buffer to the pool.
func putResponseBuffer(rb *responseBuffer) {
	rb.sz = 0
	responseBufferPool.Put(rb)
}

// Write appends data to the buffer.
func (rb *responseBuffer) Write(p []byte) (int, error) {
	if len(rb.buf) < rb.sz+len(p) {
		newBuf := make([]byte, cap(rb.buf)*2)
		copy(newBuf, rb.buf[:rb.sz])
		rb.buf = newBuf
	}
	n := copy(rb.buf[rb.sz:], p)
	rb.sz += n
	return n, nil
}

// Bytes returns the buffer contents.
func (rb *responseBuffer) Bytes() []byte {
	return rb.buf[:rb.sz]
}

// Reset clears the buffer.
func (rb *responseBuffer) Reset() {
	rb.sz = 0
}
