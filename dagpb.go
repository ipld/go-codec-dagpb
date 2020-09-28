package dagpb

// PBLink ...
type PBLink struct {
	Hash  []byte
	Name  *string
	Tsize *uint64
}

// PBNode ...
type PBNode struct {
	Links []*PBLink
	Data  []byte
}
