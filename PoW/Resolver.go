package PoW

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"math"
	"strconv"
	"strings"

	"golang.org/x/crypto/sha3"
)

type PowChallenge struct {
	Pow struct {
		Timestamp  int    `json:"timestamp"`
		Complexity int    `json:"complexity"`
		Type       string `json:"type"`
		Algorithm  struct {
			Version   int    `json:"version"`
			Resourse  string `json:"resourse"`
			Name      string `json:"name"`
			Extension string `json:"extension"`
		} `json:"algorithm"`
		Random_string string `json:"random_string"`
	} `json:"pow"`
}

func (challenge *PowChallenge) ResolveChallenge() int {
	powSchema := fmt.Sprintf("%d:%d:%d:%s:%s:%s:", challenge.Pow.Algorithm.Version, challenge.Pow.Complexity, challenge.Pow.Timestamp, challenge.Pow.Algorithm.Resourse, challenge.Pow.Algorithm.Extension, challenge.Pow.Random_string)
	h := sha3.NewLegacyKeccak512()
	var buffer bytes.Buffer

	for j := 0; j < math.MaxInt64; j++ {
		buffer.WriteString(powSchema)
		buffer.WriteString(strconv.Itoa(j))
		h.Write(buffer.Bytes())
		if strings.HasPrefix(hex.EncodeToString(h.Sum(nil)), "000") {
			return j
		}
		h.Reset()
		buffer.Reset()
	}

	return -1
}
