package dagpb

import (
	"fmt"
	"io"
)

// Unmarshal TODO
func (node *PBNode) Unmarshal(data []byte) error {
	var err error
	var fieldNum, wireType int
	l := len(data)
	index := 0
	for index < l {
		if fieldNum, wireType, index, err = decodeKey(data, index); err != nil {
			return err
		}
		if wireType != 2 {
			return fmt.Errorf("protobuf: (PBNode) invalid wireType, expected 2, got %d", wireType)
		}

		if fieldNum == 1 {
			if node.Data != nil {
				return fmt.Errorf("protobuf: (PBNode) duplicate Data section")
			}
			if node.Data, index, err = decodeBytes(data, index); err != nil {
				return err
			}
		} else if fieldNum == 2 {
			if node.Data != nil {
				return fmt.Errorf("protobuf: (PBNode) invalid order, found Data before Links content")
			}

			var chunk []byte
			if chunk, index, err = decodeBytes(data, index); err != nil {
				return err
			}
			node.Links = append(node.Links, &PBLink{})
			if err := node.Links[len(node.Links)-1].unmarshal(chunk); err != nil {
				return err
			}
		} else {
			return fmt.Errorf("protobuf: (PBNode) invalid fieldNumber, expected 1 or 2, got %d", fieldNum)
		}
	}

	if index > l {
		return io.ErrUnexpectedEOF
	}
	return nil
}

func (link *PBLink) unmarshal(data []byte) error {
	var err error
	var fieldNum, wireType int
	l := len(data)
	index := 0
	for index < l {
		if fieldNum, wireType, index, err = decodeKey(data, index); err != nil {
			return err
		}

		if fieldNum == 1 {
			if link.Hash != nil {
				return fmt.Errorf("protobuf: (PBLink) duplicate Hash section")
			}
			if link.Name != nil {
				return fmt.Errorf("protobuf: (PBLink) invalid order, found Name before Hash")
			}
			if link.Tsize != nil {
				return fmt.Errorf("protobuf: (PBLink) invalid order, found Tsize before Hash")
			}
			if wireType != 2 {
				return fmt.Errorf("protobuf: (PBLink) wrong wireType (%d) for Hash", wireType)
			}

			if link.Hash, index, err = decodeBytes(data, index); err != nil {
				return err
			}
		} else if fieldNum == 2 {
			if link.Name != nil {
				return fmt.Errorf("protobuf: (PBLink) duplicate Name section")
			}
			if link.Tsize != nil {
				return fmt.Errorf("protobuf: (PBLink) invalid order, found Tsize before Name")
			}
			if wireType != 2 {
				return fmt.Errorf("protobuf: (PBLink) wrong wireType (%d) for Name", wireType)
			}

			var chunk []byte
			if chunk, index, err = decodeBytes(data, index); err != nil {
				return err
			}
			s := string(chunk)
			link.Name = &s
		} else if fieldNum == 3 {
			if link.Tsize != nil {
				return fmt.Errorf("protobuf: (PBLink) duplicate Tsize section")
			}
			if wireType != 0 {
				return fmt.Errorf("protobuf: (PBLink) wrong wireType (%d) for Tsize", wireType)
			}

			var v uint64
			if v, index, err = decodeVarint(data, index); err != nil {
				return err
			}
			link.Tsize = &v
		} else {
			return fmt.Errorf("protobuf: (PBLink) invalid fieldNumber, expected 1, 2 or 3, got %d", fieldNum)
		}
	}

	if index > l {
		return io.ErrUnexpectedEOF
	}
	return nil
}

func decodeKey(data []byte, offset int) (int, int, int, error) {
	var wire uint64
	var err error
	if wire, offset, err = decodeVarint(data, offset); err != nil {
		return 0, 0, 0, err
	}
	fieldNum := int(wire >> 3)
	wireType := int(wire & 0x7)
	return fieldNum, wireType, offset, nil

}

func decodeBytes(data []byte, offset int) ([]byte, int, error) {
	var bytesLen uint64
	var err error
	if bytesLen, offset, err = decodeVarint(data, offset); err != nil {
		return nil, 0, err
	}
	postOffset := offset + int(bytesLen)
	if postOffset > len(data) {
		return nil, 0, io.ErrUnexpectedEOF
	}
	return data[offset:postOffset], postOffset, nil
}

func decodeVarint(data []byte, offset int) (uint64, int, error) {
	var v uint64
	l := len(data)
	for shift := uint(0); ; shift += 7 {
		if shift >= 64 {
			return 0, 0, ErrIntOverflow
		}
		if offset >= l {
			return 0, 0, io.ErrUnexpectedEOF
		}
		b := data[offset]
		offset++
		v |= uint64(b&0x7F) << shift
		if b < 0x80 {
			break
		}
	}
	return v, offset, nil
}

// ErrIntOverflow TODO
var ErrIntOverflow = fmt.Errorf("protobuf: varint overflow")
