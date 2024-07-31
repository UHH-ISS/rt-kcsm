package structure

import (
	"fmt"
	"testing"
)

func TestGetPreConditions(t *testing.T) {
	testCases := []struct {
		input    []UKCStage
		expected []UKCStage
	}{
		{input: []UKCStage{D1, R}, expected: []UKCStage{R}},
		{input: []UKCStage{E, O}, expected: []UKCStage{D1, D2, C, L, E, O, H}},
		{input: []UKCStage{L, H}, expected: []UKCStage{D1, D2, C, L, E, O, S, H}},
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("GetPreConditions(%s)", tc.input), func(t *testing.T) {
			result := GetPreConditions(tc.input...)
			if !result.Equal(NewSet(tc.expected...)) {
				t.Errorf("got %s, expected %s", result.String(), tc.expected)
			}
		})
	}
}

func TestDetermineStage(t *testing.T) {
	tests := []struct {
		name       string
		attackerIP IPAddress
		victimIP   IPAddress
		expected   SimplifiedUKCStage
	}{
		{
			name:       "Test Reconnaissance Stage",
			attackerIP: ParseIPAddress("1.1.1.1"),
			victimIP:   ParseIPAddress("10.0.0.1"),
			expected:   Recon,
		},
		{
			name:       "Test Host Stage",
			attackerIP: ParseIPAddress("10.0.0.1"),
			victimIP:   ParseIPAddress("10.0.0.1"),
			expected:   Host,
		},
		{
			name:       "Test Lateral Movement Stage",
			attackerIP: ParseIPAddress("10.0.0.1"),
			victimIP:   ParseIPAddress("10.0.0.2"),
			expected:   Lateral,
		},
		{
			name:       "Test Pivot Stage",
			attackerIP: ParseIPAddress("10.0.0.1"),
			victimIP:   ParseIPAddress("10.2.0.2"),
			expected:   Pivot,
		},
		{
			name:       "Test Exfiltration Stage",
			attackerIP: ParseIPAddress("10.0.0.1"),
			victimIP:   ParseIPAddress("1.1.1.1"),
			expected:   Exfiltration,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := DetermineStage(tt.attackerIP, tt.victimIP)
			if !NewSet(result).Equal(NewSet(tt.expected)) {
				t.Errorf("Test Case %s: Expected %v but got %v", tt.name, tt.expected, result)
			}
		})
	}
}
