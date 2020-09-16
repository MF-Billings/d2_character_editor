package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"math/bits"
)

func main() {
	data, err := ioutil.ReadFile("paladin.d2s")

	if err != nil {
		fmt.Printf("An error occurred: %s", err)
		return
	}

	// print character name
	fmt.Println(header()["character_name"].value(data))

	attributes(data)
}

func checksum(data []byte, offset, length int) uint32 {
	var cheksum uint32 = 0
	var x uint32
	var val byte

	checksum_starts := 0x0C // offset of first byte of checksum from beginning of file
	checksum_length_in_bytes := 4

	// calculate checksum
	for i := offset; i < offset+length; i++ {
		val = data[i]
		// each byte that's part of the stored checksum is considered equal to 0 for the purpose of calculation
		if i >= checksum_starts && i < checksum_starts+checksum_length_in_bytes {
			val = 0
		}
		x = 0
		if cheksum&0x80000000 != 0 {
			x = 1
		}
		cheksum = (cheksum << 1) + uint32(val) + x
	}
	// byte order is little endian so the byte order needs to be reversed to give the expected value
	return bits.ReverseBytes32(cheksum)
}

func littleEndianChecksum(data []byte, offset, length int) uint32 {
	var sum uint32 = 0
	for i := offset; i < offset+length; i++ {
		var x uint32 = 0
		if sum&0x80000000 != 0 {
			x = 1
		}
		sum = (sum << 1) + uint32(data[i]) + x
	}
	return sum
}

func header() map[string]Field {
	fields := make(map[string]Field, 5)

	fields["character_name"] = Field{
		offset: 20,
		length: 16,
	}
	return fields
}

// ATTRIBUTES ------------------------------------------------------------------------------------------------------

// starts after header at byte 766
func attributes(data []byte) {
	const first_byte_of_section = 767
	const end_of_attributes = 0x1FF

	r := NewBitReader(bytes.NewBuffer(data[first_byte_of_section:]))
	i := 0
	for true {
		id, err := r.ReadBits(9, true)
		if err != nil {
			// halt and catch fire
		}

		if id == end_of_attributes {
			return
		}

		id = bitReversed(id, 9)
		num_bits_to_read := bit_length_by_attribute_id[uint(id)]
		reversed_attribute_value, err := r.ReadBits(num_bits_to_read, true)
		if err != nil {
			// halt and catch fire
		}

		attribute_value := bitReversed(reversed_attribute_value, int(num_bits_to_read))
		switch id {
		case hp_current, hp_max, mana_current, mana_max, stamina_current, stamina_max:
			attribute_value = attribute_value / 256
		}

		fmt.Printf("ID = %d; will read %d bits\n", id, num_bits_to_read)
		fmt.Printf("Attribute value is %b = %d\n", reversed_attribute_value, attribute_value)

		i++
	}
	return
}

// IDs that show up and their meaning
const (
	strength        = 0
	energy          = 1
	dexterity       = 2
	vitality        = 3
	unused_stats    = 4
	unused_skills   = 5
	hp_current      = 6
	hp_max          = 7
	mana_current    = 8
	mana_max        = 9
	stamina_current = 10
	stamina_max     = 11
	level           = 12
	experience      = 13
	gold            = 14
	gold_stashed    = 15
)

var bit_length_by_attribute_id = map[uint]uint{
	strength:        10,
	energy:          10,
	dexterity:       10,
	vitality:        10,
	unused_stats:    10,
	unused_skills:   8,
	hp_current:      21,
	hp_max:          21,
	mana_current:    21,
	mana_max:        21,
	stamina_current: 21,
	stamina_max:     21,
	level:           7,
	experience:      32,
	gold:            25,
	gold_stashed:    25,
}

// Bit Operations ------------------------------------------------------------------------------------------------------

func bitReversed(x uint64, num_bits int) (b uint64) {
	a := x // >> (64 - num_bits)
	b = a  // will become the reversed value
	for i := 0; i < num_bits; i++ {
		b <<= 1
		b |= a & 1 // copy value of rightmost bit to b
		a >>= 1
	}
	b &= (1 << num_bits) - 1 // use mask to remove unwanted bits
	return
}

func bitsToInt(bits []bool) int {
	x := 0
	for i := 0; i < len(bits); i++ {
		bit := 0
		if bits[i] {
			bit = 1
		}
		x += bit
		x = x << 1
	}
	return x
}

//	------------------------------------------------------------------------------------------------------

type D2s struct {
	data   []byte
	header Header
}

type Header struct {
	data   [765]byte
	fields []Field
}

func (h Header) checksum() []byte {
	chksum := h.fields[0]
	checksum := h.data[chksum.offset:chksum.end()]
	return checksum
}

type Field struct {
	offset int
	length int
}

func (f Field) end() int {
	return f.offset + f.length
}

func (f Field) value(bytes []byte) string {
	return string(bytes[f.offset:f.end()])
}

func testChecksum(data []byte) {
	// checksum := littleEndianChecksum(data, 0, len(data))
	checksum := checksum(data, 0, len(data))
	// fmt.Println(checksum)
	fmt.Printf("%x\n", uint32(checksum))
}
