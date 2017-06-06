package decor_test

import (
	"testing"

	"github.com/vbauerster/mpb/decor"
)

func TestFormatNoUnits(t *testing.T) {
	actual := decor.Format(1234567).String()
	expected := "1234567"
	if actual != expected {
		t.Errorf("Expected %q but found %q", expected, actual)
	}
}

func TestFormatWidth(t *testing.T) {
	actual := decor.Format(1234567).Width(10).String()
	expected := "   1234567"
	if actual != expected {
		t.Errorf("Expected %q but found %q", expected, actual)
	}
}

func TestFormatToBytes(t *testing.T) {
	inputs := []struct {
		v int64
		e string
	}{
		{v: 1000, e: "1000b"},
		{v: 1024, e: "1.0KiB"},
		{v: 3*decor.MiB + 140*decor.KiB, e: "3.1MiB"},
		{v: 2 * decor.GiB, e: "2.0GiB"},
		{v: 4 * decor.TiB, e: "4.0TiB"},
	}

	for _, input := range inputs {
		actual := decor.Format(input.v).To(decor.Unit_KiB).String()
		if actual != input.e {
			t.Errorf("Expected %q but found %q", input.e, actual)
		}
	}
}
