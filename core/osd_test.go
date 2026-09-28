package core

import (
	"encoding/json"
	"testing"
)

func TestHasConfiguredChannelName(t *testing.T) {
	tests := []struct {
		name    string
		table   []any
		channel string
		want    bool
	}{
		{
			name:    "all channels match",
			table:   []any{map[string]any{"Name": "p2pwn"}, map[string]any{"Name": "p2pwn"}},
			channel: "p2pwn",
			want:    true,
		},
		{
			name:    "one channel differs",
			table:   []any{map[string]any{"Name": "p2pwn"}, map[string]any{"Name": "Camera 2"}},
			channel: "p2pwn",
		},
		{
			name:    "nested channel differs",
			table:   []any{[]any{map[string]any{"Name": "Camera 1"}}},
			channel: "p2pwn",
		},
		{
			name:    "missing channel names",
			table:   []any{map[string]any{"Index": 0}},
			channel: "p2pwn",
		},
		{
			name:    "empty configured channel",
			table:   []any{map[string]any{"Name": "p2pwn"}},
			channel: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := hasConfiguredChannelName(tt.table, tt.channel); got != tt.want {
				t.Errorf("hasConfiguredChannelName() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestResetBrightnessTable(t *testing.T) {
	table := []any{
		map[string]any{"Brightness": 20, "Contrast": 61},
		map[string]any{"Brightness": 80},
		"unexpected row",
	}

	changed, err := resetBrightnessTable(table)
	if err != nil {
		t.Fatalf("resetBrightnessTable() error = %v", err)
	}
	if !changed {
		t.Fatal("resetBrightnessTable() changed = false, want true")
	}
	if got := table[0].(map[string]any)["Brightness"]; got != 50 {
		t.Errorf("first brightness = %v, want 50", got)
	}
	if got := table[0].(map[string]any)["Contrast"]; got != 61 {
		t.Errorf("contrast = %v, want 61", got)
	}
	if got := table[1].(map[string]any)["Brightness"]; got != 50 {
		t.Errorf("second brightness = %v, want 50", got)
	}
}

func TestResetBrightnessTableWithoutBrightness(t *testing.T) {
	changed, err := resetBrightnessTable([]any{map[string]any{"Contrast": 50}})
	if err == nil {
		t.Fatal("resetBrightnessTable() error = nil, want missing-field error")
	}
	if changed {
		t.Fatal("resetBrightnessTable() changed = true, want false")
	}
}

func TestResetBrightnessTableNestedVideoColor(t *testing.T) {
	table := []any{[]any{
		map[string]any{"Brightness": 20, "Gamma": 100},
		map[string]any{"Brightness": 80, "Saturation": 51},
	}}

	changed, err := resetBrightnessTable(table)
	if err != nil {
		t.Fatalf("resetBrightnessTable() error = %v", err)
	}
	if !changed {
		t.Fatal("resetBrightnessTable() changed = false, want true")
	}
	rows := table[0].([]any)
	if got := rows[0].(map[string]any)["Brightness"]; got != 50 {
		t.Errorf("first brightness = %v, want 50", got)
	}
	if got := rows[1].(map[string]any)["Brightness"]; got != 50 {
		t.Errorf("second brightness = %v, want 50", got)
	}
	if got := rows[0].(map[string]any)["Gamma"]; got != 100 {
		t.Errorf("gamma = %v, want 100", got)
	}
	if err := verifyBrightnessTable(table); err != nil {
		t.Errorf("verifyBrightnessTable() error = %v", err)
	}
}

func TestResetVideoControlFieldsNested(t *testing.T) {
	colorTable := []any{[]any{
		map[string]any{"Contrast": 32, "Saturation": 61, "Gamma": 100, "Brightness": 20, "Style": "Standard"},
		map[string]any{"Contrast": 70, "Saturation": 40, "Gamma": 25},
	}}
	fields := []string{"Contrast", "Saturation", "Gamma"}
	if err := resetConfigFields(colorTable, fields...); err != nil {
		t.Fatalf("resetConfigFields(VideoColor) error = %v", err)
	}
	if err := verifyConfigFields(colorTable, fields...); err != nil {
		t.Fatalf("verifyConfigFields(VideoColor) error = %v", err)
	}
	colorRows := colorTable[0].([]any)
	for _, raw := range colorRows {
		row := raw.(map[string]any)
		for _, field := range fields {
			if got := row[field]; got != 50 {
				t.Errorf("%s = %v, want 50", field, got)
			}
		}
	}
	if got := colorRows[0].(map[string]any)["Brightness"]; got != 20 {
		t.Errorf("brightness = %v, want unchanged value 20", got)
	}
	if got := colorRows[0].(map[string]any)["Style"]; got != "Standard" {
		t.Errorf("style = %v, want unchanged value Standard", got)
	}

	sharpnessTable := []any{[]any{map[string]any{"Sharpness": 12, "Level": 50, "Mode": 1}}}
	if err := resetConfigFields(sharpnessTable, "Sharpness"); err != nil {
		t.Fatalf("resetConfigFields(VideoInSharpness) error = %v", err)
	}
	if err := verifyConfigFields(sharpnessTable, "Sharpness"); err != nil {
		t.Fatalf("verifyConfigFields(VideoInSharpness) error = %v", err)
	}
	sharpnessRow := sharpnessTable[0].([]any)[0].(map[string]any)
	if got := sharpnessRow["Sharpness"]; got != 50 {
		t.Errorf("sharpness = %v, want 50", got)
	}
}

func TestVerifyBrightnessTable(t *testing.T) {
	tests := []struct {
		name    string
		value   any
		wantErr bool
	}{
		{name: "integer", value: 50},
		{name: "JSON number", value: float64(50)},
		{name: "string number", value: json.Number("50")},
		{name: "ignored write", value: 49, wantErr: true},
		{name: "malformed value", value: map[string]any{"value": 50}, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := verifyBrightnessTable([]any{map[string]any{"Brightness": tt.value}})
			if (err != nil) != tt.wantErr {
				t.Fatalf("verifyBrightnessTable() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
