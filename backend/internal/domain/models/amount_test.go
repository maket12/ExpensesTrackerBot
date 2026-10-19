package models

import (
	"testing"
)

func TestNewAmount(t *testing.T) {
	tests := []struct {
		name    string
		val     float64
		want    float64
		wantErr bool
	}{
		{
			name:    "valid amount",
			val:     150.55,
			want:    150.55,
			wantErr: false,
		},
		{
			name:    "rounds to 2 decimal places",
			val:     150.556,
			want:    150.56,
			wantErr: false,
		},
		{
			name:    "zero amount is invalid",
			val:     0,
			want:    0,
			wantErr: true,
		},
		{
			name:    "negative amount is invalid",
			val:     -10.5,
			want:    0,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewAmount(tt.val)
			if (err != nil) != tt.wantErr {
				t.Fatalf("NewAmount() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr && got.Value() != tt.want {
				t.Errorf("NewAmount() = %v, want %v", got.Value(), tt.want)
			}
		})
	}
}
