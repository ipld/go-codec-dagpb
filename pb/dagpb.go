package pb

import (
	"bytes"
	"sort"

	cid "github.com/ipfs/go-cid"
)

// PBLink ...
type PBLink struct {
	Hash  *cid.Cid
	Name  *string
	Tsize *uint64
}

// PBNode ...
type PBNode struct {
	Links []*PBLink
	Data  []byte
}

func NewPBNode() *PBNode {
	n := &PBNode{Links: make([]*PBLink, 0)}
	return n
}

func NewPBNodeFromData(data []byte) *PBNode {
	n := &PBNode{Data: data, Links: make([]*PBLink, 0)}
	return n
}

type linkBuilder struct {
	link *PBLink
}

func (lb *linkBuilder) SetHash(hash *cid.Cid) error {
	lb.link.Hash = hash
	return nil
}

func (lb *linkBuilder) SetName(name *string) error {
	lb.link.Name = name
	return nil
}

func (lb *linkBuilder) SetTsize(tsize uint64) error {
	lb.link.Tsize = &tsize
	return nil
}

func (lb *linkBuilder) Done() error { return nil }

type nodeBuilder struct {
	node *PBNode
}

func (nb *nodeBuilder) SetData(data []byte) error {
	nb.node.Data = data
	return nil
}

func (nb *nodeBuilder) AddLink() (PBLinkBuilder, error) {
	nb.mklinks()
	nb.node.Links = append(nb.node.Links, &PBLink{})
	return &linkBuilder{nb.node.Links[len(nb.node.Links)-1]}, nil
}

func (nb *nodeBuilder) mklinks() {
	if nb.node.Links == nil {
		nb.node.Links = make([]*PBLink, 0)
	}
}

func (nb *nodeBuilder) Done() error {
	nb.mklinks()
	return nil
}

func UnmarshalPBNode(byts []byte) (*PBNode, error) {
	nb := nodeBuilder{NewPBNode()}

	if err := Unmarshal(bytes.NewReader(byts), &nb); err != nil {
		return nil, err
	}
	nb.Done()
	return nb.node, nil
}

func NewPBLinkFromCid(c cid.Cid) *PBLink {
	l := &PBLink{Hash: &c}
	return l
}

func NewPBLink(name string, c cid.Cid, tsize uint64) *PBLink {
	l := &PBLink{Name: &name, Hash: &c, Tsize: &tsize}
	return l
}

func (node *PBNode) SortLinks() {
	sort.Stable(pbLinkSlice(node.Links))
}

type pbLinkSlice []*PBLink

func (ls pbLinkSlice) Len() int           { return len(ls) }
func (ls pbLinkSlice) Swap(a, b int)      { ls[a], ls[b] = ls[b], ls[a] }
func (ls pbLinkSlice) Less(a, b int) bool { return pbLinkLess(ls[a], ls[b]) }

func pbLinkLess(a *PBLink, b *PBLink) bool {
	return *a.Name < *b.Name
}
