package dagpb

// based on the original pb codegen

import (
	"fmt"
	math_bits "math/bits"
)

// Marshal TODO
func (node *PBNode) Marshal() (data []byte, err error) {
	if err := node.validate(); err != nil {
		return nil, err
	}
	size := node.size()
	data = make([]byte, size)

	i := len(data)
	if node.Data != nil {
		i -= len(node.Data)
		copy(data[i:], node.Data)
		i = encodeVarint(data, i, uint64(len(node.Data))) - 1
		data[i] = 0xa
	}
	if len(node.Links) > 0 {
		for iNdEx := len(node.Links) - 1; iNdEx >= 0; iNdEx-- {
			size, err := node.Links[iNdEx].marshal(data[:i])
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

func (link *PBLink) marshal(data []byte) (int, error) {
	i := len(data)
	if link.Tsize != nil {
		i = encodeVarint(data, i, uint64(*link.Tsize)) - 1
		data[i] = 0x18
	}
	if link.Name != nil {
		i -= len(*link.Name)
		copy(data[i:], *link.Name)
		i = encodeVarint(data, i, uint64(len(*link.Name))) - 1
		data[i] = 0x12
	}
	if link.Hash != nil {
		byts := link.Hash.Bytes()
		i -= len(byts)
		copy(data[i:], byts)
		i = encodeVarint(data, i, uint64(len(byts))) - 1
		data[i] = 0xa
	} else {
		return 0, fmt.Errorf("invalid DAG-PB form (link must have a Hash)")
	}
	return len(data) - i, nil
}

func (node *PBNode) validate() error {
	if node == nil {
		return fmt.Errorf("PBNode not defined")
	}

	if node.Links == nil {
		return fmt.Errorf("invalid DAG-PB form (Links must be an array)")
	}

	for i, link := range node.Links {
		if link.Hash == nil {
			return fmt.Errorf("invalid DAG-PB form (link must have a Hash)")
		}

		if i > 0 && pbLinkLess(link, node.Links[i-1]) {
			return fmt.Errorf("invalid DAG-PB form (links must be sorted by Name bytes)")
		}
	}

	return nil
}

func (link *PBLink) size() (n int) {
	if link == nil {
		return 0
	}
	var l int
	if link.Hash != nil {
		l = link.Hash.ByteLen()
		n += 1 + l + sizeOfVarint(uint64(l))
	}
	if link.Name != nil {
		l = len(*link.Name)
		n += 1 + l + sizeOfVarint(uint64(l))
	}
	if link.Tsize != nil {
		n += 1 + sizeOfVarint(uint64(*link.Tsize))
	}
	return n
}

func (node *PBNode) size() (n int) {
	if node == nil {
		return 0
	}
	var l int
	if node.Data != nil {
		l = len(node.Data)
		n += 1 + l + sizeOfVarint(uint64(l))
	}
	if len(node.Links) > 0 {
		for _, e := range node.Links {
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
