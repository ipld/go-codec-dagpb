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

			{ // Hash
				d, err := link.LookupByString("Hash")
				if err != nil {
					return err
				}
				// TODO:
				// 		return 0, fmt.Errorf("invalid DAG-PB form (link must have a Hash)")
				l, err := d.AsLink()
				if err != nil {
					return err
				}
				if err != nil {
					return err
				}
				cl, ok := l.(cidlink.Link)
				if !ok {
					return fmt.Errorf("unexpected Link type [%v]", l)
				}
				pbLinks[ii].hash = cl.Cid
			}

			{ // Name
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

			{ // Tsize
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

		sortLinks(pbLinks)
		for _, link := range pbLinks {
			size := link.encodedSize()
			chunk := make([]byte, size+sizeOfVarint(uint64(size))+1)
			chunk[0] = 0x12
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
		lead[0] = 0xa
		encodeVarint(lead, 1, size)
		out.Write(lead)
		out.Write(byts)
	}

	return nil
}

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

func (link pbLink) marshal(data []byte, offset int) (int, error) {
	base := offset
	data[offset] = 0xa
	byts := link.hash.Bytes()
	offset = encodeVarint(data, offset+1, uint64(len(byts)))
	copy(data[offset:], byts)
	offset += len(byts)
	if link.hasName {
		data[offset] = 0x12
		offset = encodeVarint(data, offset+1, uint64(len(link.name)))
		copy(data[offset:], link.name)
		offset += len(link.name)
	}
	if link.hasTsize {
		data[offset] = 0x18
		offset = encodeVarint(data, offset+1, uint64(link.tsize))
	}
	return offset - base, nil
}

func encodeVarint(data []byte, offset int, v uint64) int {
	for v >= 1<<7 {
		data[offset] = uint8(v&0x7f | 0x80)
		v >>= 7
		offset++
	}
	data[offset] = uint8(v)
	return offset + 1
}

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

func sizeOfVarint(x uint64) (n int) {
	return (math_bits.Len64(x|1) + 6) / 7
}
