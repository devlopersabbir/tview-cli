package chart

import "testing"

func TestCompactNumber(t *testing.T) {
	tests := []struct {
		name  string
		value float64
		want  string
	}{
		{name: "plain", value: 538.54, want: "538.54"},
		{name: "thousands", value: 2370, want: "2.37K"},
		{name: "millions", value: 12_500_000, want: "12.50M"},
		{name: "negative", value: -1420, want: "-1.42K"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := compactNumber(tt.value); got != tt.want {
				t.Fatalf("compactNumber(%v) = %q, want %q", tt.value, got, tt.want)
			}
		})
	}
}
