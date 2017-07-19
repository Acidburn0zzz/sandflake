/*
Copyright 2017 Salim Alami

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.

*/

package sandflake

import (
	"bytes"
	"encoding/base32"
	"fmt"
	"time"
)

const (
	alphabet   = "0123456789ABCDEFGHJKMNPQRSTVWXYZ"
	padding    = 6
	encodedLen = 26
	timeUnit   = int64(time.Millisecond)

	timestampLength = 6
	workerIDLength  = 4
	sequenceLength  = 3
	randomLength    = 3

	workerIDOffset = timestampLength
	sequenceOffset = timestampLength + workerIDLength
	randomOffset   = sequenceOffset + sequenceLength

	byteLength = timestampLength + workerIDLength + sequenceLength + randomLength
)

var (
	Nil   ID
	MaxID = ID{255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255}

	maxSequence = MaxID.Sequence()
)

var (
	// errors
	ErrInvalidLength      = fmt.Errorf("expected length of id to be %d", encodedLen)
	ErrInvalidBytesLength = fmt.Errorf("expected length of bytes to be %d", byteLength)

	// encoding
	encoding     = base32.NewEncoding(alphabet)
	paddingBytes = bytes.Repeat([]byte{'='}, padding)
)

// 48bit: timestamp
// 32bit: worker id (defaults to MAC Address)
// 24bit: sequence number
// 24bit: random number
type ID [byteLength]byte

// NewID returns a ID with the provided arguments.
func NewID(t time.Time, workerID WorkerID, seq uint32, randomBytes []byte) ID {
	var d ID
	ts := uint64(t.UnixNano() / timeUnit)
	d.setTimestamp(ts)
	copy(d[timestampLength:timestampLength+workerIDLength], workerID[:])
	d.setSequence(seq)
	copy(d[randomOffset:], randomBytes)

	return d
}

func Parse(str string) (ID, error) {
	var id ID
	err := id.UnmarshalText([]byte(str))
	if err != nil {
		return id, err
	}

	return id, nil
}

func FromBytes(b []byte) (ID, error) {
	if len(b) != byteLength {
		return Nil, ErrInvalidBytesLength
	}

	var id ID
	copy(id[:], b)
	return id, nil
}

func (id ID) String() string {
	b, _ := id.MarshalText()
	return string(b)
}

func (d ID) MarshalText() ([]byte, error) {
	dst := make([]byte, encodedLen+padding)
	encoding.Encode(dst, d[:])
	return dst[:encodedLen], nil
}

func (d *ID) UnmarshalText(b []byte) error {
	if len(b) != encodedLen {
		return ErrInvalidLength
	}

	dst := make([]byte, encodedLen)
	if _, err := encoding.Decode(dst, append(b, paddingBytes...)); err != nil {
		return err
	}

	copy((*d)[:], dst)

	return nil
}

// Time returns the underlying time
func (d ID) Time() time.Time {
	var ts uint64
	for i := 0; i < timestampLength; i++ {
		ts |= uint64(uint64(d[i]) << uint64((timestampLength-i-1)*8))
	}
	return time.Unix(0, int64(ts)*timeUnit).UTC()
}

func (d *ID) setTimestamp(ts uint64) {
	for i := 0; i < timestampLength; i++ {
		(*d)[i] = byte(ts >> uint64((timestampLength-i-1)*8))
	}
}
func (d *ID) setSequence(ts uint32) {
	n := sequenceOffset + sequenceLength
	for i := sequenceOffset; i < n; i++ {
		(*d)[i] = byte(ts >> uint32((n-i-1)*8))
	}
}

// WorkerID returns the underlying worker id
func (d ID) WorkerID() (wid WorkerID) {
	copy(wid[:], d[workerIDOffset:workerIDOffset+workerIDLength])
	return wid
}

// Sequence returns the underlying sequence number
func (d ID) Sequence() uint32 {
	var ts uint32
	n := sequenceOffset + sequenceLength
	for i := sequenceOffset; i < n; i++ {
		ts |= uint32(uint32(d[i]) << uint32((n-i-1)*8))
	}
	return ts
}

// RandomBytes returns the random stored in the ID
func (d ID) RandomBytes() []byte {
	return d[randomOffset:]
}

func (d ID) Before(other ID) bool {
	return Compare(d, other) < 0
}

func (d ID) After(other ID) bool {
	return Compare(d, other) > 0
}

func Compare(id1, id2 ID) int {
	return bytes.Compare(id1[:], id2[:])
}
