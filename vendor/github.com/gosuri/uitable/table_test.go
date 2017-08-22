package uitable

import (
	"sync"
	"testing"
)

func TestCell(t *testing.T) {
	c := &Cell{
		Data:  "foo bar",
		Width: 5,
	}

	got := c.String()
	if got != "fo..." {
		t.Fatal("need", "fo...", "got", got)
	}
	if c.LineWidth() != 5 {
		t.Fatal("need", 5, "got", c.LineWidth())
	}

	c.Wrap = true
	got = c.String()
	if got != "foo\nbar" {
		t.Fatal("need", "foo\nbar", "got", got)
	}
	if c.LineWidth() != 3 {
		t.Fatal("need", 3, "got", c.LineWidth())
	}
}

func TestRow(t *testing.T) {
	row := &Row{
		Separator: "\t",
		Cells: []*Cell{
			{Data: "foo", Width: 3, Wrap: true},
			{Data: "bar baz", Width: 3, Wrap: true},
		},
	}
	got := row.String()
	need := "foo\tbar\n   \tbaz"

	if got != need {
		t.Fatalf("need: %q | got: %q ", need, got)
	}
}

func TestAlign(t *testing.T) {
	table := New()
	table.AddRow("foo", "bar baz")
	table.Rows = []*Row{{
		Separator: "\t",
		Cells: []*Cell{
			{Data: "foo", Width: 5, Wrap: true},
			{Data: "bar baz", Width: 10, Wrap: true},
		},
	}}
	table.RightAlign(1)
	got := table.String()
	need := "foo  \t   bar baz"

	if got != need {
		t.Fatalf("need: %q | got: %q ", need, got)
	}
}

func TestAddRow(t *testing.T) {
	var wg sync.WaitGroup
	table := New()
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			table.AddRow("foo")
		}()
	}
	wg.Wait()
	if len(table.Rows) != 100 {
		t.Fatal("want", 100, "got", len(table.Rows))
	}
}
