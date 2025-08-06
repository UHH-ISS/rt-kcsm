package structure

import (
	"fmt"
	"rtkcsm/component/structure/set"
	"testing"
)

func TestGetPrecedingStages(t *testing.T) {
	stageMapper := SimplifiedUKCStageMapper{}

	testCases := []struct {
		input    SimplifiedUKCStage
		expected []SimplifiedUKCStage
	}{
		{input: Incoming, expected: []SimplifiedUKCStage{}},
		{input: Outgoing, expected: []SimplifiedUKCStage{Incoming, SameZone, DifferentZone, Host, Outgoing}},
		{input: SameZone, expected: []SimplifiedUKCStage{Incoming, SameZone, DifferentZone, Host, Outgoing}},
		{input: DifferentZone, expected: []SimplifiedUKCStage{Incoming, DifferentZone, SameZone, Host, Outgoing}},
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("GetPrecedingStages(%v)", tc.input), func(t *testing.T) {
			result := stageMapper.GetPrecedingStages(tc.input)
			if !set.NewSet(result...).Equal(set.NewSet(tc.expected...)) {
				t.Errorf("got %v, expected %v", result, tc.expected)
			}
		})
	}
}

func TestDetermineStage(t *testing.T) {
	stageMapper := SimplifiedUKCStageMapper{}
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
			expected:   Incoming,
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
			expected:   SameZone,
		},
		{
			name:       "Test Pivot Stage",
			attackerIP: ParseIPAddress("10.0.0.1"),
			victimIP:   ParseIPAddress("10.2.0.2"),
			expected:   DifferentZone,
		},
		{
			name:       "Test Exfiltration Stage",
			attackerIP: ParseIPAddress("10.0.0.1"),
			victimIP:   ParseIPAddress("1.1.1.1"),
			expected:   Outgoing,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := stageMapper.DetermineStage(Alert{
				SourceIP:      tt.attackerIP,
				DestinationIP: tt.victimIP,
			})
			if err != nil && !set.NewSet(result).Equal(set.NewSet(tt.expected)) {
				t.Errorf("Test Case %s: Expected %v but got %v", tt.name, tt.expected, result)
			}
		})
	}
}
