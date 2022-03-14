package helpers

import "testing"

/**
ParseRangeHeader should return the range values for a given header
*/
func TestParseRangeHeader_ok(t *testing.T) {
	sampleContent := "Bytes=12-38/96"

	result, err := ParseRangeHeader(sampleContent)
	if err != nil {
		t.Error("Got unexpected error from ParseRangeHeader: ", err)
	} else {
		if result == nil {
			t.Error("ParseRangeHeader returned nil, nil when we expected data")
		} else {
			if result.Start != 12 {
				t.Error("Expected start to be 12 got ", result.Start)
			}
			if result.End != 38 {
				t.Error("Expected end to be 38 got ", result.End)
			}
			if result.TotalSize != 96 {
				t.Error("Expected totalSize to be 96 got ", result.TotalSize)
			}
		}
	}
}

func TestParseRangeHeader_malformed(t *testing.T) {
	sampleContent := "sdjksdfjksfsf fsdfsdfsf"

	_, err := ParseRangeHeader(sampleContent)
	if err == nil {
		t.Error("Expected error for malformed header but got none")
	}
}

func TestParseRangeHeader_invalid_values(t *testing.T) {
	sampleContent := "Bytes 12-4a/8888"

	_, err := ParseRangeHeader(sampleContent)
	if err == nil {
		t.Error("Expected error for malformed header but got none")
	}
}

func TestParseRangeHeader_notbytes(t *testing.T) {
	sampleContent := "Records 12-38/96"

	_, err := ParseRangeHeader(sampleContent)
	if err == nil {
		t.Error("Expected error for non-bytes field but got none")
	}
}
