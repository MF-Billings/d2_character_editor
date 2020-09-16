package main

import (
	"io"
	"math/bits"
)

var c cursor

type BitReader struct {
	reader io.ByteReader
	byte   byte
	offset byte
}

type cursor struct {
	num_bits_consumed uint
	buffer            uint64
}

func NewBitReader(r io.ByteReader) *BitReader {
	return &BitReader{r, 0, 0}
}

func (r *BitReader) ReadBits(num_bits_desired uint, reverse bool) (uint64, error) {
	var err error

	for c.num_bits_consumed < num_bits_desired {
		byt, _ := r.reader.ReadByte()
		b := uint8(byt)
		if reverse {
			b = bits.Reverse8(uint8(byt))
		}
		// fmt.Printf("%08b = b\n", b)
		feed(b)
		// fmt.Printf("%08b = buffer\n", c.buffer)
		c.num_bits_consumed += 8
	}

	num_excess_bits := c.num_bits_consumed - num_bits_desired
	n := c.buffer >> num_excess_bits

	n = desiredBitsExtracted(n, num_bits_desired)
	// fmt.Printf("%d = n\n", n)
	c.num_bits_consumed -= num_bits_desired
	return n, err
}

/**
 * Make room in buffer for new byte by discarding the oldest byte read
 */
func feed(b uint8) {
	c.buffer <<= 8
	c.buffer |= uint64(b)
}

func desiredBitsExtracted(n uint64, num_bits_desired uint) uint64 {
	least_significant_bits_mask := (1 << num_bits_desired) - 1
	n &= uint64(least_significant_bits_mask)
	return n
}
