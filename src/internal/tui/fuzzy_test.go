package tui

import "testing"

func TestFuzzyMatch_Exact(t *testing.T) {
	result := FuzzyMatch("HTTP server", "HTTP server")
	if !result.Match {
		t.Error("expected exact match")
	}
}

func TestFuzzyMatch_Partial(t *testing.T) {
	result := FuzzyMatch("HTTP server with middleware", "http serv")
	if !result.Match {
		t.Error("expected partial match")
	}
	if result.Score <= 0 {
		t.Error("expected positive score")
	}
}

func TestFuzzyMatch_NoMatch(t *testing.T) {
	result := FuzzyMatch("HTTP server", "zzz")
	if result.Match {
		t.Error("expected no match")
	}
}

func TestFuzzyMatch_ConsecutiveBonus(t *testing.T) {
	r1 := FuzzyMatch("abcdef", "abc")
	r2 := FuzzyMatch("aXbXcX", "abc")
	if r1.Score <= r2.Score {
		t.Errorf("consecutive match should score higher: %d vs %d", r1.Score, r2.Score)
	}
}

func TestFuzzyMatch_WordBoundaryBonus(t *testing.T) {
	r1 := FuzzyMatch("http_server", "hs")
	r2 := FuzzyMatch("ahahs", "hs")
	if r1.Score <= r2.Score {
		t.Errorf("word boundary match should score higher: %d vs %d", r1.Score, r2.Score)
	}
}

func TestFuzzyMatch_Empty(t *testing.T) {
	result := FuzzyMatch("anything", "")
	if !result.Match {
		t.Error("empty query should always match")
	}
	if result.Score != 0 {
		t.Error("empty query should have 0 score")
	}
}
