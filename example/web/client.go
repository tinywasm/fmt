//go:build wasm

package main

import (
	"syscall/js"

	. "github.com/tinywasm/fmt"
	"github.com/tinywasm/fmt/lang"
)

func main() {
	// Your WebAssembly code here ok

	// Crear el elemento div
	dom := js.Global().Get("document").Call("createElement", "div")

	// Demonstrate conversion processes like in README line 48
	items := []string{"  ÁPPLE  ", "  banána  ", "  piñata  ", "  ÑANDÚ  "}

	buf := Convert().
		Write("<h1>fmt Conversion Processes</h1>").
		Write("<h2>Original Items:</h2>").
		Write("<ul>")

	// Show original items
	for _, item := range items {
		buf.Write("<li>").Write(item).Write("</li>")
	}

	buf.Write("</ul>").
		Write("<h2>After Processing:</h2>").
		Write("<ul>")

	// Process items like in README example
	builder := Convert()
	for i, item := range items {
		processed := Convert(item).
			TrimSpace().  // TrimSpace whitespace
			Tilde().      // Normalize accents
			ToLower().    // Convert to lowercase
			Capitalize(). // Capitalize first letter
			String()      // Finalize the string
		builder.Write(processed)
		if i < len(items)-1 {
			builder.Write(" - ")
		}
	}

	buf.Write("<li>").Write(builder.String()).Write("</li>").
		Write("</ul>").
		Write("<h2>Conversion Steps Applied:</h2>").
		Write("<ol>").
		Write("<li>TrimSpace() - Remove leading/trailing whitespace</li>").
		Write("<li>Tilde() - Normalize accents (á→a, ñ→n, etc.)</li>").
		Write("<li>ToLower() - Convert to lowercase</li>").
		Write("<li>Capitalize() - Capitalize first letter</li>").
		Write("</ol>")

	dom.Set("innerHTML", buf.String())

	// Obtener el body del documento y agregar el elemento
	body := js.Global().Get("document").Get("body")
	body.Call("appendChild", dom)

	logger := func(msg ...any) {
		js.Global().Get("console").Call("log", lang.Translate(msg...).String())
	}

	logger("hello tinystring:", 123, 45.67, true, []string{"a", "b", "c"})

	select {}
}
