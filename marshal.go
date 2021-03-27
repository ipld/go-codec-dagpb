package dagpb

import (
	"fmt"
	"io"
	math_bits "math/bits"
	"sort"

	"github.com/ipfs/go-cid"
	ipld "github.com/ipld/go-ipld-prime"
	cidlink "github.com/ipld/go-ipld-prime/linking/cid"
)

type pbLink struct {
	hash     cid.Cid
	name     string
	hasName  bool
	tsize    uint64
	hasTsize bool
}

// Marshal provides an IPLD codec encode interface for DAG-PB data. Provide a
// conforming Node and a destination for bytes to marshal a DAG-PB IPLD Node.
// The Node must strictly conform to the DAG-PB schema
// (https://github.com/ipld/specs/blob/master/block-layer/codecs/dag-pb.md).
// For safest use, build Nodes using the Type.PBNode type.
func Marshal(inNode ipld.Node, out io.Writer) error {
	// Wrap in a typed node for some basic schema form checking
	builder := Type.PBNode.NewBuilder()
	if err := builder.AssignNode(inNode); err != nil {
		return err
	}
	node := builder.Build()

	links, err := node.LookupByString("Links")
	if err != nil {
		return err
	}
	if links.Length() > 0 {
		// collect links into a slice so we can properly sort for encoding
		pbLinks := make([]pbLink, links.Length())

		linksIter := links.ListIterator()
		for !linksIter.Done() {
			ii, link, err := linksIter.Next()
			if err != nil {
				return err
			}

			{ // Hash (required)
				d, err := link.LookupByString("Hash")
				if err != nil {
					return err
				}
				l, err := d.AsLink()
				if err != nil {
					return err
				}
				if err != nil {
					return err
				}
				cl, ok := l.(cidlink.Link)
				if !ok {
					// this _should_ be taken care of by the Typed conversion above with
					// "missing required fields: Hash"
					return fmt.Errorf("invalid DAG-PB form (link must have a Hash)")
				}
				pbLinks[ii].hash = cl.Cid
			}

			{ // Name (optional)
				nameNode, err := link.LookupByString("Name")
				if err != nil {
					return err
				}
				if !nameNode.IsAbsent() {
					name, err := nameNode.AsString()
					if err != nil {
						return err
					}
					pbLinks[ii].name = name
					pbLinks[ii].hasName = true
				}
			}

			{ // Tsize (optional)
				tsizeNode, err := link.LookupByString("Tsize")
				if err != nil {
					return err
				}
				if !tsizeNode.IsAbsent() {
					tsize, err := tsizeNode.AsInt()
					if err != nil {
						return err
					}
					if tsize < 0 {
						return fmt.Errorf("Link has negative Tsize value [%v]", tsize)
					}
					utsize := uint64(tsize)
					pbLinks[ii].tsize = utsize
					pbLinks[ii].hasTsize = true
				}
			}
		} // for

		// links must be strictly sorted by Name before encoding, leaving stable
		// ordering where the names are the same (or absent)
		sortLinks(pbLinks)
		for _, link := range pbLinks {
			size := link.encodedSize()
			chunk := make([]byte, size+sizeOfVarint(uint64(size))+1)
			chunk[0] = 0x12 // field & wire type for Links
			offset := encodeVarint(chunk, 1, uint64(size))
			wrote, err := link.marshal(chunk, offset)
			if err != nil {
				return err
			}
			if wrote != size {
				return fmt.Errorf("bad PBLink marshal, wrote wrong number of bytes")
			}
			out.Write(chunk)
		}
	} // if links

	// Data (optional)
	data, err := node.LookupByString("Data")
	if err != nil {
		return err
	}
	if !data.IsAbsent() {
		byts, err := data.AsBytes()
		if err != nil {
			return err
		}
		size := uint64(len(byts))
		lead := make([]byte, sizeOfVarint(size)+1)
		lead[0] = 0xa // field and wireType for Data
		encodeVarint(lead, 1, size)
		out.Write(lead)
		out.Write(byts)
	}

	return nil
}

// predict the byte size of the encoded Link
func (link pbLink) encodedSize() (n int) {
	l := link.hash.ByteLen()
	n += 1 + l + sizeOfVarint(uint64(l))
	if link.hasName {
		l = len(link.name)
		n += 1 + l + sizeOfVarint(uint64(l))
	}
	if link.hasTsize {
		n += 1 + sizeOfVarint(uint64(link.tsize))
	}
	return n
}

// encode a Link to PB
func (link pbLink) marshal(data []byte, offset int) (int, error) {
	base := offset
	data[offset] = 0xa // field and wireType for Hash
	byts := link.hash.Bytes()
	offset = encodeVarint(data, offset+1, uint64(len(byts)))
	copy(data[offset:], byts)
	offset += len(byts)
	if link.hasName {
		data[offset] = 0x12 // field and wireType for Name
		offset = encodeVarint(data, offset+1, uint64(len(link.name)))
		copy(data[offset:], link.name)
		offset += len(link.name)
	}
	if link.hasTsize {
		data[offset] = 0x18 // field and wireType for Tsize
		offset = encodeVarint(data, offset+1, uint64(link.tsize))
	}
	return offset - base, nil
}

// predict the size of a varint for PB before creating it
func sizeOfVarint(x uint64) (n int) {
	return (math_bits.Len64(x|1) + 6) / 7
}

// encode a varint to a PB chunk
func encodeVarint(data []byte, offset int, v uint64) int {
	for v >= 1<<7 {
		data[offset] = uint8(v&0x7f | 0x80)
		v >>= 7
		offset++
	}
	data[offset] = uint8(v)
	return offset + 1
}

// stable sorting of Links using the strict sorting rules
func sortLinks(links []pbLink) {
	sort.Stable(pbLinkSlice(links))
}

type pbLinkSlice []pbLink

func (ls pbLinkSlice) Len() int           { return len(ls) }
func (ls pbLinkSlice) Swap(a, b int)      { ls[a], ls[b] = ls[b], ls[a] }
func (ls pbLinkSlice) Less(a, b int) bool { return pbLinkLess(ls[a], ls[b]) }

func pbLinkLess(a pbLink, b pbLink) bool {
	return a.name < b.name
}
