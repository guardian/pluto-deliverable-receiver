package helpers

import (
	"errors"
	"net/http"
	"regexp"
	"strconv"
	"strings"
)

type RangeHeader struct {
	Start     int64
	End       int64
	TotalSize int64
}

/**
returns true if the given range has a start of zero
*/
func (h *RangeHeader) IsFirst() bool {
	return h.Start == 0
}

/**
returns true if the given range ends at the total size
*/
func (h *RangeHeader) IsLast() bool {
	return h.End >= h.TotalSize
}

/**
returns true if this is the one and only chunk
*/
func (h *RangeHeader) IsComplete() bool {
	return h.IsFirst() && h.IsLast()
}
func ParseRangeHeader(headerContent string) (*RangeHeader, error) {
	xtractor := regexp.MustCompile("^(\\w+)=(\\d+)-(\\d+)/(\\d+)$")

	matches := xtractor.FindAllStringSubmatch(headerContent, -1)
	if matches == nil {
		return nil, errors.New("range header malformed")
	}

	if strings.ToLower(matches[0][1]) != "bytes" {
		return nil, errors.New("need ranges in bytes")
	}

	startNum, startParseErr := strconv.ParseInt(matches[0][2], 10, 64)
	endNum, endParseErr := strconv.ParseInt(matches[0][3], 10, 64)
	totalNum, totalParseErr := strconv.ParseInt(matches[0][4], 10, 64)

	if startParseErr != nil || endParseErr != nil || totalParseErr != nil {
		return nil, errors.New("need all range parameters as numbers")
	}

	return &RangeHeader{
		Start:     startNum,
		End:       endNum,
		TotalSize: totalNum,
	}, nil
}

/**
extracts any Range header, if present.
if no range header is present, (nil, nil) is returned
if there was an error interpreting it, an error is returned only
otherwise a pointer to a new RangeHeader object is returned
*/
func ExtractRange(h *http.Request) (*RangeHeader, error) {
	rangeHeaderRaw := h.Header.Get("Range")
	if rangeHeaderRaw == "" {
		return nil, nil
	} else {
		return ParseRangeHeader(rangeHeaderRaw)
	}
}
