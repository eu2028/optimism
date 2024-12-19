package tmpl

import (
	"bytes"
	"strings"
	"testing"
)

func TestNewTemplateContext(t *testing.T) {
	t.Run("creates empty context", func(t *testing.T) {
		ctx := NewTemplateContext()
		if ctx.Data != nil {
			t.Error("expected nil Data in new context")
		}
		if len(ctx.Functions) != 0 {
			t.Error("expected empty Functions map in new context")
		}
	})

	t.Run("adds data with WithData option", func(t *testing.T) {
		data := map[string]string{"key": "value"}
		ctx := NewTemplateContext(WithData(data))
		if ctx.Data == nil {
			t.Error("expected non-nil Data in context")
		}
		if d, ok := ctx.Data.(map[string]string); !ok || d["key"] != "value" {
			t.Error("data not set correctly")
		}
	})

	t.Run("adds function with WithFunction option", func(t *testing.T) {
		fn := func(s string) (string, error) { return s + "test", nil }
		ctx := NewTemplateContext(WithFunction("testfn", fn))
		if len(ctx.Functions) != 1 {
			t.Error("expected one function in context")
		}
		if _, ok := ctx.Functions["testfn"]; !ok {
			t.Error("function not added with correct name")
		}
	})
}

func TestInstantiateTemplate(t *testing.T) {
	t.Run("simple template substitution", func(t *testing.T) {
		data := map[string]string{"name": "world"}
		ctx := NewTemplateContext(WithData(data))

		input := strings.NewReader("Hello {{.name}}!")
		var output bytes.Buffer

		err := ctx.InstantiateTemplate(input, &output)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		expected := "Hello world!"
		if got := output.String(); got != expected {
			t.Errorf("expected %q, got %q", expected, got)
		}
	})

	t.Run("template with custom function", func(t *testing.T) {
		upper := func(s string) (string, error) { return strings.ToUpper(s), nil }
		ctx := NewTemplateContext(
			WithData(map[string]string{"name": "world"}),
			WithFunction("upper", upper),
		)

		input := strings.NewReader("Hello {{upper .name}}!")
		var output bytes.Buffer

		err := ctx.InstantiateTemplate(input, &output)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		expected := "Hello WORLD!"
		if got := output.String(); got != expected {
			t.Errorf("expected %q, got %q", expected, got)
		}
	})

	t.Run("invalid template syntax", func(t *testing.T) {
		ctx := NewTemplateContext()
		input := strings.NewReader("Hello {{.name")
		var output bytes.Buffer

		err := ctx.InstantiateTemplate(input, &output)
		if err == nil {
			t.Error("expected error for invalid template syntax")
		}
	})

	t.Run("missing data field", func(t *testing.T) {
		ctx := NewTemplateContext()
		input := strings.NewReader("Hello {{.name}}!")
		var output bytes.Buffer

		err := ctx.InstantiateTemplate(input, &output)
		if err == nil {
			t.Error("expected error for missing data field")
		}
	})

	t.Run("multiple functions and data fields", func(t *testing.T) {
		upper := func(s string) (string, error) { return strings.ToUpper(s), nil }
		lower := func(s string) (string, error) { return strings.ToLower(s), nil }

		data := map[string]string{
			"greeting": "Hello",
			"name":     "World",
		}

		ctx := NewTemplateContext(
			WithData(data),
			WithFunction("upper", upper),
			WithFunction("lower", lower),
		)

		input := strings.NewReader("{{upper .greeting}} {{lower .name}}!")
		var output bytes.Buffer

		err := ctx.InstantiateTemplate(input, &output)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		expected := "HELLO world!"
		if got := output.String(); got != expected {
			t.Errorf("expected %q, got %q", expected, got)
		}
	})
}
