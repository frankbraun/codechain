// Copyright (c) 2015 Mute Communications Ltd.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package fuzzer provides a sequential fuzzer for testing purposes.
package fuzzer

// SequentialFuzzer fuzzes data one bit at a time, sequentially
type SequentialFuzzer struct {
	Data       []byte             // Data to fuzz
	Start, End int                // Range to fuzz
	Errors     []error            // Errors returned
	ErrorCount int                // Total number of errors
	TestCount  int                // Total number of tests
	TestFunc   func([]byte) error // The test function
}

// Fuzz runs a fuzzing test on a SequentialFuzzer and returns false if less
// errors were returned than tests run.
func (sf *SequentialFuzzer) Fuzz() bool {
	if sf.Data == nil || len(sf.Data) == 0 || sf.TestFunc == nil {
		panic("Fuzz setup failed")
	}
	sf.ErrorCount, sf.TestCount = 0, 0
	l := len(sf.Data) * 8
	if sf.End > l || sf.End == 0 {
		sf.End = l
	}
	if sf.Start < 0 || sf.Start > l {
		sf.Start = 0
	}
	numTests := sf.End - sf.Start
	sf.Errors = make([]error, numTests)
	for i := sf.Start; i < sf.End; i++ {
		err := sf.TestFunc(switchBit(sf.Data, i))
		sf.Errors[i-sf.Start] = err
		sf.TestCount++
		if err != nil {
			sf.ErrorCount++
		}
	}
	if sf.TestCount != sf.ErrorCount {
		return false
	}
	return true
}

// switchBit negates a single bit in d
func switchBit(d []byte, pos int) []byte {
	mask := [8]byte{1, 2, 4, 8, 16, 32, 64, 128}
	x := make([]byte, len(d))
	copy(x, d) // Work on copies since []byte are pointers
	x[pos/8] ^= mask[pos%8]
	return x
}
