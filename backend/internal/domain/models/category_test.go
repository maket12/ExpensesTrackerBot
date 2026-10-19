package models

import (
	"testing"
)

func TestNewCategory(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    Category
		wantErr bool
	}{
		{
			name:    "valid food",
			input:   "еда",
			want:    CategoryFood,
			wantErr: false,
		},
		{
			name:    "valid transport trimmed and lowercased",
			input:   "  ТРАНСПОРТ ",
			want:    CategoryTransport,
			wantErr: false,
		},
		{
			name:    "invalid category",
			input:   "крипта",
			want:    "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewCategory(tt.input)
			if (err != nil) != tt.wantErr {
				t.Fatalf("NewCategory() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr && got != tt.want {
				t.Errorf("NewCategory() = %v, want %v", got, tt.want)
			}
		})
	}
}
