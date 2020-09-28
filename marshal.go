package dagpb

// based on the original pb codegen

import (
	math_bits "math/bits"
)

// Marshal TODO
func (m *PBNode) Marshal() (data []byte, err error) {
	size := m.size()
	data = make([]byte, size)

	i := len(data)
	if m.Data != nil {
		i -= len(m.Data)
		copy(data[i:], m.Data)
		i = encodeVarint(data, i, uint64(len(m.Data))) - 1
		data[i] = 0xa
	}
	if len(m.Links) > 0 {
		for iNdEx := len(m.Links) - 1; iNdEx >= 0; iNdEx-- {
			size, err := m.Links[iNdEx].marshal(data[:i])
			if err != nil {
				return nil, err
			}
			i -= size
			i = encodeVarint(data, i, uint64(size)) - 1
			data[i] = 0x12
		}
	}
	return data[:size], nil
}

func (m *PBLink) marshal(data []byte) (int, error) {
	i := len(data)
	if m.Tsize != nil {
		i = encodeVarint(data, i, uint64(*m.Tsize)) - 1
		data[i] = 0x18
	}
	if m.Name != nil {
		i -= len(*m.Name)
		copy(data[i:], *m.Name)
		i = encodeVarint(data, i, uint64(len(*m.Name))) - 1
		data[i] = 0x12
	}
	if m.Hash != nil {
		i -= len(m.Hash)
		copy(data[i:], m.Hash)
		i = encodeVarint(data, i, uint64(len(m.Hash))) - 1
		data[i] = 0xa
	}
	return len(data) - i, nil
}

func (m *PBLink) size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	if m.Hash != nil {
		l = len(m.Hash)
		n += 1 + l + sizeOfVarint(uint64(l))
	}
	if m.Name != nil {
		l = len(*m.Name)
		n += 1 + l + sizeOfVarint(uint64(l))
	}
	if m.Tsize != nil {
		n += 1 + sizeOfVarint(uint64(*m.Tsize))
	}
	return n
}

func (m *PBNode) size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	if m.Data != nil {
		l = len(m.Data)
		n += 1 + l + sizeOfVarint(uint64(l))
	}
	if len(m.Links) > 0 {
		for _, e := range m.Links {
			l = e.size()
			n += 1 + l + sizeOfVarint(uint64(l))
		}
	}
	return n
}

func encodeVarint(data []byte, offset int, v uint64) int {
	offset -= sizeOfVarint(v)
	base := offset
	for v >= 1<<7 {
		data[offset] = uint8(v&0x7f | 0x80)
		v >>= 7
		offset++
	}
	data[offset] = uint8(v)
	return base
}

func sizeOfVarint(x uint64) (n int) {
	return (math_bits.Len64(x|1) + 6) / 7
}
