//go:build ignore

package main

import (
	"bytes"
	"fmt"
	"os"

	"github.com/alecthomas/chroma/v2/formatters/html"
	"github.com/alecthomas/chroma/v2/styles"
)

func main() {
	lightStyle := styles.Get("github")
	lightFormatter := html.New(html.WithClasses(true))
	var lightBuf bytes.Buffer
	lightFormatter.WriteCSS(&lightBuf, lightStyle)
	os.WriteFile("internal/server/chroma_light.css", lightBuf.Bytes(), 0644)
	fmt.Printf("Generated chroma_light.css (%d bytes)\n", lightBuf.Len())

	darkStyle := styles.Get("monokai")
	darkFormatter := html.New(html.WithClasses(true))
	var darkBuf bytes.Buffer
	darkFormatter.WriteCSS(&darkBuf, darkStyle)
	os.WriteFile("internal/server/chroma_dark.css", darkBuf.Bytes(), 0644)
	fmt.Printf("Generated chroma_dark.css (%d bytes)\n", darkBuf.Len())
}
