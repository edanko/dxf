package query

import (
	"testing"
)

func TestQueryParser(t *testing.T) {
	t.Run("Lexer tokens", func(t *testing.T) {
		lexer := NewLexer(`layer=="construction"`)
		tokens, err := lexer.Lex()
		if err != nil {
			t.Fatalf("Failed to lex: %v", err)
		}

		for _, tok := range tokens {
			t.Logf("Token: %s", tok)
		}
	})

	t.Run("Parse simple comparison", func(t *testing.T) {
		parser, err := NewParser(`layer=="construction"`)
		if err != nil {
			t.Fatalf("Failed to create parser: %v", err)
		}

		expr, err := parser.Parse()
		if err != nil {
			t.Fatalf("Failed to parse: %v", err)
		}

		t.Logf("Parsed expression: %s", expr.String())
	})

	t.Run("Parse AND expression", func(t *testing.T) {
		parser, err := NewParser(`layer=="A" AND color==1`)
		if err != nil {
			t.Fatalf("Failed to create parser: %v", err)
		}

		expr, err := parser.Parse()
		if err != nil {
			t.Fatalf("Failed to parse: %v", err)
		}

		t.Logf("Parsed expression: %s", expr.String())
	})

	t.Run("Parse OR expression", func(t *testing.T) {
		parser, err := NewParser(`layer=="A" OR layer=="B"`)
		if err != nil {
			t.Fatalf("Failed to create parser: %v", err)
		}

		expr, err := parser.Parse()
		if err != nil {
			t.Fatalf("Failed to parse: %v", err)
		}

		t.Logf("Parsed expression: %s", expr.String())
	})

	t.Run("Parse NOT expression", func(t *testing.T) {
		parser, err := NewParser(`NOT layer=="A"`)
		if err != nil {
			t.Fatalf("Failed to create parser: %v", err)
		}

		expr, err := parser.Parse()
		if err != nil {
			t.Fatalf("Failed to parse: %v", err)
		}

		t.Logf("Parsed expression: %s", expr.String())
	})

	t.Run("Parse numeric comparison", func(t *testing.T) {
		parser, err := NewParser(`radius>5`)
		if err != nil {
			t.Fatalf("Failed to create parser: %v", err)
		}

		expr, err := parser.Parse()
		if err != nil {
			t.Fatalf("Failed to parse: %v", err)
		}

		t.Logf("Parsed expression: %s", expr.String())
	})

	t.Run("Parse CONTAINS operator", func(t *testing.T) {
		parser, err := NewParser(`text CONTAINS "note"`)
		if err != nil {
			t.Fatalf("Failed to create parser: %v", err)
		}

		expr, err := parser.Parse()
		if err != nil {
			t.Fatalf("Failed to parse: %v", err)
		}

		t.Logf("Parsed expression: %s", expr.String())
	})

	t.Run("Parse complex expression", func(t *testing.T) {
		parser, err := NewParser(`(layer=="A" OR layer=="B") AND color>0`)
		if err != nil {
			t.Fatalf("Failed to create parser: %v", err)
		}

		expr, err := parser.Parse()
		if err != nil {
			t.Fatalf("Failed to parse: %v", err)
		}

		t.Logf("Parsed expression: %s", expr.String())
	})

	t.Run("Parse list value", func(t *testing.T) {
		parser, err := NewParser(`color IN [1, 2, 3]`)
		if err != nil {
			t.Fatalf("Failed to create parser: %v", err)
		}

		expr, err := parser.Parse()
		if err != nil {
			t.Fatalf("Failed to parse: %v", err)
		}

		t.Logf("Parsed expression: %s", expr.String())
	})
}

func TestQueryParserIntegration(t *testing.T) {
	t.Run("Compare values", func(t *testing.T) {
		result := compareValues(10, 5, OpGreaterThan)
		if !result {
			t.Error("Expected 10 > 5 to be true")
		}

		result = compareValues("hello", "lo", OpContains)
		if !result {
			t.Error("Expected 'hello' to contain 'lo'")
		}

		result = compareValues("hello", "world", OpContains)
		if result {
			t.Error("Expected 'hello' NOT to contain 'world'")
		}

		result = compareValues(1, 1, OpEquals)
		if !result {
			t.Error("Expected 1 == 1 to be true")
		}
	})
}
