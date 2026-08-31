package dagpb

import (
	"fmt"
	"io"
	"sort"

	"github.com/ipfs/go-cid"
	ipld "github.com/ipld/go-ipld-prime"
	cidlink "github.com/ipld/go-ipld-prime/linking/cid"
	"google.golang.org/protobuf/encoding/protowire"
)

type pbLink struct {
	hash     cid.Cid
	name     string
	hasName  bool
	tsize    uint64
	hasTsize bool
}

// FieldOrder selects the order of the top-level PBNode fields in the
// serialized form. Both orders decode to the same node, but produce
// different bytes and therefore different CIDs.
type FieldOrder int

const (
	// LinksFirst writes the repeated Links field (field number 2) before
	// the Data field (field number 1). This is the canonical DAG-PB order
	// and the default.
	LinksFirst FieldOrder = iota

	// DataFirst writes the Data field (field number 1) before the repeated
	// Links field (field number 2), so streaming readers can process Data
	// (for example UnixFS HAMT parameters) before reading links. Opt-in:
	// proposed for the unixfs-v1-2026 UnixFS profile by IPIP-550
	// (https://github.com/ipfs/specs/pull/550).
	DataFirst
)

// EncodeOptions customizes the encoded form of a PBNode. The zero value
// matches the behavior of the package-level Encode and AppendEncode.
type EncodeOptions struct {
	// FieldOrder selects the order of the top-level PBNode fields. The
	// default, LinksFirst, is the canonical DAG-PB form.
	FieldOrder FieldOrder
}

// Encode is like the package-level Encode, customized by o.
func (o EncodeOptions) Encode(node ipld.Node, w io.Writer) error {
	// 1KiB can be allocated on the stack, and covers most small nodes
	// without having to grow the buffer and cause allocations.
	enc := make([]byte, 0, 1024)

	enc, err := o.AppendEncode(enc, node)
	if err != nil {
		return err
	}
	_, err = w.Write(enc)
	return err
}

// AppendEncode is like the package-level AppendEncode, customized by o.
func (o EncodeOptions) AppendEncode(enc []byte, inNode ipld.Node) ([]byte, error) {
	// Wrap in a typed node for some basic schema form checking
	builder := Type.PBNode.NewBuilder()
	if err := builder.AssignNode(inNode); err != nil {
		return enc, err
	}
	node := builder.Build()

	var err error
	if o.FieldOrder == DataFirst {
		if enc, err = appendData(enc, node); err != nil {
			return enc, err
		}
		return appendLinks(enc, node)
	}
	if enc, err = appendLinks(enc, node); err != nil {
		return enc, err
	}
	return appendData(enc, node)
}

// Encode provides an IPLD codec encode interface for DAG-PB data. Provide a
// conforming Node and a destination for bytes to marshal a DAG-PB IPLD Node.
// The Node must strictly conform to the DAG-PB schema
// (https://github.com/ipld/specs/blob/master/block-layer/codecs/dag-pb.md).
// For safest use, build Nodes using the Type.PBNode type.
// This function is registered via the go-ipld-prime link loader for multicodec
// code 0x70 when this package is invoked via init.
func Encode(node ipld.Node, w io.Writer) error {
	return EncodeOptions{}.Encode(node, w)
}

// AppendEncode is like Encode, but it uses a destination buffer directly.
// This means less copying of bytes, and if the destination has enough capacity,
// fewer allocations.
func AppendEncode(enc []byte, inNode ipld.Node) ([]byte, error) {
	return EncodeOptions{}.AppendEncode(enc, inNode)
}

// appendLinks appends the wire encoding of the node's repeated Links field.
func appendLinks(enc []byte, node ipld.Node) ([]byte, error) {
	links, err := node.LookupByString("Links")
	if err != nil {
		return enc, err
	}

	if links.Length() > 0 {
		// collect links into a slice so we can properly sort for encoding
		pbLinks := make([]pbLink, links.Length())

		linksIter := links.ListIterator()
		for !linksIter.Done() {
			ii, link, err := linksIter.Next()
			if err != nil {
				return enc, err
			}

			{ // Hash (required)
				d, err := link.LookupByString("Hash")
				if err != nil {
					return enc, err
				}
				l, err := d.AsLink()
				if err != nil {
					return enc, err
				}
				cl, ok := l.(cidlink.Link)
				if !ok {
					// this _should_ be taken care of by the Typed conversion above with
					// "missing required fields: Hash"
					return enc, fmt.Errorf("invalid DAG-PB form (link must have a Hash)")
				}
				pbLinks[ii].hash = cl.Cid
			}

			{ // Name (optional)
				nameNode, err := link.LookupByString("Name")
				if err != nil {
					return enc, err
				}
				if !nameNode.IsAbsent() {
					name, err := nameNode.AsString()
					if err != nil {
						return enc, err
					}
					pbLinks[ii].name = name
					pbLinks[ii].hasName = true
				}
			}

			{ // Tsize (optional)
				tsizeNode, err := link.LookupByString("Tsize")
				if err != nil {
					return enc, err
				}
				if !tsizeNode.IsAbsent() {
					tsize, err := tsizeNode.AsInt()
					if err != nil {
						return enc, err
					}
					if tsize < 0 {
						return enc, fmt.Errorf("Link has negative Tsize value [%v]", tsize)
					}
					utsize := uint64(tsize)
					pbLinks[ii].tsize = utsize
					pbLinks[ii].hasTsize = true
				}
			}
		} // for

		// links must be strictly sorted by Name before encoding, leaving stable
		// ordering where the names are the same (or absent)
		sort.Stable(pbLinkSlice(pbLinks))
		for _, link := range pbLinks {
			hash := link.hash.Bytes()

			size := 0
			size += protowire.SizeTag(2)
			size += protowire.SizeBytes(len(hash))
			if link.hasName {
				size += protowire.SizeTag(2)
				size += protowire.SizeBytes(len(link.name))
			}
			if link.hasTsize {
				size += protowire.SizeTag(3)
				size += protowire.SizeVarint(uint64(link.tsize))
			}

			enc = protowire.AppendTag(enc, 2, 2) // field & wire type for Links
			enc = protowire.AppendVarint(enc, uint64(size))

			enc = protowire.AppendTag(enc, 1, 2) // field & wire type for Hash
			enc = protowire.AppendBytes(enc, hash)
			if link.hasName {
				enc = protowire.AppendTag(enc, 2, 2) // field & wire type for Name
				enc = protowire.AppendString(enc, link.name)
			}
			if link.hasTsize {
				enc = protowire.AppendTag(enc, 3, 0) // field & wire type for Tsize
				enc = protowire.AppendVarint(enc, uint64(link.tsize))
			}
		}
	} // if links

	return enc, nil
}

// appendData appends the wire encoding of the node's Data field, if present.
func appendData(enc []byte, node ipld.Node) ([]byte, error) {
	data, err := node.LookupByString("Data")
	if err != nil {
		return enc, err
	}
	if !data.IsAbsent() {
		byts, err := data.AsBytes()
		if err != nil {
			return enc, err
		}
		enc = protowire.AppendTag(enc, 1, 2) // field & wire type for Data
		enc = protowire.AppendBytes(enc, byts)
	}

	return enc, nil
}

type pbLinkSlice []pbLink

func (ls pbLinkSlice) Len() int           { return len(ls) }
func (ls pbLinkSlice) Swap(a, b int)      { ls[a], ls[b] = ls[b], ls[a] }
func (ls pbLinkSlice) Less(a, b int) bool { return pbLinkLess(ls[a], ls[b]) }

func pbLinkLess(a pbLink, b pbLink) bool {
	return a.name < b.name
}
