package dagpb

import (
	"io"

	"github.com/ipfs/go-cid"
	ipld "github.com/ipld/go-ipld-prime"
	cidlink "github.com/ipld/go-ipld-prime/linking/cid"
	"github.com/polydawn/refmt/shared"
	"golang.org/x/xerrors"
)

// ErrIntOverflow is returned a varint overflows during decode, it indicates
// malformed data
var ErrIntOverflow = xerrors.Errorf("protobuf: varint overflow")

// Unmarshal provides an IPLD codec decode interface for DAG-PB data. Provide
// a compatible NodeAssembler and a byte source to unmarshal a DAG-PB IPLD
// Node. Use the NodeAssembler from the PBNode type for safest construction
// (Type.PBNode.NewBuilder()). A Map assembler will also work.
func Unmarshal(na ipld.NodeAssembler, in io.Reader) error {
	ma, err := na.BeginMap(2)
	if err != nil {
		return err
	}
	// always make "Links", even if we don't use it
	if err = ma.AssembleKey().AssignString("Links"); err != nil {
		return err
	}
	links, err := ma.AssembleValue().BeginList(0)
	if err != nil {
		return err
	}

	haveData := false
	reader := shared.NewReader(in)
	for {
		_, err := reader.Readn1()
		if err == io.EOF {
			break
		}
		reader.Unreadn1()

		fieldNum, wireType, err := decodeKey(reader)
		if err != nil {
			return err
		}
		if wireType != 2 {
			return xerrors.Errorf("protobuf: (PBNode) invalid wireType, expected 2, got %d", wireType)
		}

		if fieldNum == 1 {
			if haveData {
				return xerrors.Errorf("protobuf: (PBNode) duplicate Data section")
			}
			var chunk []byte
			if chunk, err = decodeBytes(reader); err != nil {
				return err
			}
			// Data must come after Links, so it's safe to close this here even if we
			// didn't use it
			if err := links.Finish(); err != nil {
				return err
			}
			links = nil
			if err := ma.AssembleKey().AssignString("Data"); err != nil {
				return err
			}
			if err := ma.AssembleValue().AssignBytes(chunk); err != nil {
				return err
			}
			haveData = true
		} else if fieldNum == 2 {
			if haveData {
				return xerrors.Errorf("protobuf: (PBNode) invalid order, found Data before Links content")
			}

			bytesLen, err := decodeVarint(reader)
			if err != nil {
				return err
			}
			curLink, err := links.AssembleValue().BeginMap(3)
			if err != nil {
				return err
			}
			if err = unmarshalLink(reader, int(bytesLen), curLink); err != nil {
				return err
			}
			if err := curLink.Finish(); err != nil {
				return err
			}
		} else {
			return xerrors.Errorf("protobuf: (PBNode) invalid fieldNumber, expected 1 or 2, got %d", fieldNum)
		}
	}

	if links != nil {
		if err := links.Finish(); err != nil {
			return err
		}
	}
	return ma.Finish()
}

func unmarshalLink(reader shared.SlickReader, length int, ma ipld.MapAssembler) error {
	haveHash := false
	haveName := false
	haveTsize := false
	startOffset := reader.NumRead()
	for {
		readBytes := reader.NumRead() - startOffset
		if readBytes == length {
			break
		} else if readBytes > length {
			return xerrors.Errorf("protobuf: (PBLink) bad length for link")
		}
		fieldNum, wireType, err := decodeKey(reader)
		if err != nil {
			return err
		}

		if fieldNum == 1 {
			if haveHash {
				return xerrors.Errorf("protobuf: (PBLink) duplicate Hash section")
			}
			if haveName {
				return xerrors.Errorf("protobuf: (PBLink) invalid order, found Name before Hash")
			}
			if haveTsize {
				return xerrors.Errorf("protobuf: (PBLink) invalid order, found Tsize before Hash")
			}
			if wireType != 2 {
				return xerrors.Errorf("protobuf: (PBLink) wrong wireType (%d) for Hash", wireType)
			}

			var chunk []byte
			if chunk, err = decodeBytes(reader); err != nil {
				return err
			}
			var c cid.Cid
			if _, c, err = cid.CidFromBytes(chunk); err != nil {
				return xerrors.Errorf("invalid Hash field found in link, expected CID (%v)", err)
			}
			if err := ma.AssembleKey().AssignString("Hash"); err != nil {
				return err
			}
			if err := ma.AssembleValue().AssignLink(cidlink.Link{Cid: c}); err != nil {
				return err
			}
			haveHash = true
		} else if fieldNum == 2 {
			if haveName {
				return xerrors.Errorf("protobuf: (PBLink) duplicate Name section")
			}
			if haveTsize {
				return xerrors.Errorf("protobuf: (PBLink) invalid order, found Tsize before Name")
			}
			if wireType != 2 {
				return xerrors.Errorf("protobuf: (PBLink) wrong wireType (%d) for Name", wireType)
			}

			var chunk []byte
			if chunk, err = decodeBytes(reader); err != nil {
				return err
			}
			if err := ma.AssembleKey().AssignString("Name"); err != nil {
				return err
			}
			if err := ma.AssembleValue().AssignString(string(chunk)); err != nil {
				return err
			}
			haveName = true
		} else if fieldNum == 3 {
			if haveTsize {
				return xerrors.Errorf("protobuf: (PBLink) duplicate Tsize section")
			}
			if wireType != 0 {
				return xerrors.Errorf("protobuf: (PBLink) wrong wireType (%d) for Tsize", wireType)
			}

			var v uint64
			if v, err = decodeVarint(reader); err != nil {
				return err
			}
			if err := ma.AssembleKey().AssignString("Tsize"); err != nil {
				return err
			}
			if err := ma.AssembleValue().AssignInt(int(v)); err != nil {
				return err
			}
			haveTsize = true
		} else {
			return xerrors.Errorf("protobuf: (PBLink) invalid fieldNumber, expected 1, 2 or 3, got %d", fieldNum)
		}
	}

	if !haveHash {
		return xerrors.Errorf("invalid Hash field found in link, expected CID")
	}

	return nil
}

// decode the lead for a PB chunk, fieldNum & wireType, that tells us which
// field in the schema we're looking at and what data type it is
func decodeKey(reader shared.SlickReader) (int, int, error) {
	var wire uint64
	var err error
	if wire, err = decodeVarint(reader); err != nil {
		return 0, 0, err
	}
	fieldNum := int(wire >> 3)
	wireType := int(wire & 0x7)
	return fieldNum, wireType, nil
}

// decode a byte string from PB
func decodeBytes(reader shared.SlickReader) ([]byte, error) {
	bytesLen, err := decodeVarint(reader)
	if err != nil {
		return nil, err
	}
	byts, err := reader.Readn(int(bytesLen))
	if err != nil {
		return nil, xerrors.Errorf("protobuf: unexpected read error: %w", err)
	}
	return byts, nil
}

// decode a varint from PB
func decodeVarint(reader shared.SlickReader) (uint64, error) {
	var v uint64
	for shift := uint(0); ; shift += 7 {
		if shift >= 64 {
			return 0, ErrIntOverflow
		}
		b, err := reader.Readn1()
		if err != nil {
			return 0, xerrors.Errorf("protobuf: unexpected read error: %w", err)
		}
		v |= uint64(b&0x7F) << shift
		if b < 0x80 {
			break
		}
	}
	return v, nil
}
