package protocol

import "testing"

func TestDefaultSignatureType(t *testing.T) {
	if DefaultSignatureType != 7 {
		t.Errorf("DefaultSignatureType = %v, want 7", DefaultSignatureType)
	}
}

func TestSAMVersions(t *testing.T) {
	if MinSAMVersion != "3.0" {
		t.Errorf("MinSAMVersion = %v, want 3.0", MinSAMVersion)
	}

	if MaxSAMVersion != "3.3" {
		t.Errorf("MaxSAMVersion = %v, want 3.3", MaxSAMVersion)
	}
}

func TestResultConstants(t *testing.T) {
	if ResultOK != "OK" {
		t.Errorf("ResultOK = %v, want OK", ResultOK)
	}
}
