package matter

import (
	"fmt"
	"regexp"
	"slices"
	"strings"

	"github.com/hasty/alchemy/matter/conformance"
	"github.com/hasty/alchemy/matter/types"
	"github.com/hasty/alchemy/parse"
)

type Bitmap struct {
	Name        string          `json:"name,omitempty"`
	Description string          `json:"description,omitempty"`
	Type        *types.DataType `json:"type,omitempty"`
	Bits        BitSet          `json:"bits,omitempty"`
}

func (c *Bitmap) EntityType() types.EntityType {
	return types.EntityTypeBitmap
}

func (c *Bitmap) BaseDataType() types.BaseDataType {
	return c.Type.BaseType
}

func (c *Bitmap) NullValue() uint64 {
	return c.Type.NullValue()
}

func (c *Bitmap) Size() int {
	if c.Type == nil {
		return 8
	}
	switch c.Type.BaseType {
	case types.BaseDataTypeMap64:
		return 64
	case types.BaseDataTypeMap32:
		return 32
	case types.BaseDataTypeMap16:
		return 16
	default:
		return 8
	}
}

func (bm *Bitmap) Clone() *Bitmap {
	nbm := &Bitmap{Name: bm.Name, Description: bm.Description}
	if bm.Type != nil {
		nbm.Type = bm.Type.Clone()
	}
	for _, b := range bm.Bits {
		nbm.Bits = append(nbm.Bits, b.Clone())
	}
	return nbm
}

func (bm Bitmap) Identifier(id string) (types.Entity, bool) {
	if len(bm.Bits) == 0 {
		return nil, false
	}
	for _, b := range bm.Bits {
		if b.Code == id {
			return b, true
		}
	}
	return nil, false
}

func (bm *Bitmap) Inherit(parent *Bitmap) error {
	mergedBits := make(BitSet, 0, len(parent.Bits))
	for _, b := range parent.Bits {
		mergedBits = append(mergedBits, b.Clone())
	}
	for _, b := range bm.Bits {
		var matching *BitmapBit
		for _, mb := range mergedBits {
			if b.Bit == mb.Bit {
				matching = b
				break
			}
		}
		if matching == nil {
			mergedBits = append(mergedBits, b.Clone())
			continue
		}
		if len(b.Summary) > 0 {
			matching.Summary = b.Summary
		}
		if len(b.Conformance) > 0 {
			matching.Conformance = b.Conformance.CloneSet()
		}
	}
	if bm.Type == nil {
		bm.Type = parent.Type
	}
	if len(bm.Description) == 0 {
		bm.Description = parent.Description
	}
	slices.SortFunc(mergedBits, func(a, b *BitmapBit) int {
		return strings.Compare(a.Bit, b.Bit)
	})
	bm.Bits = mergedBits
	return nil
}

type BitmapSet []*Bitmap

func (bs BitmapSet) Identifier(name string) (types.Entity, bool) {
	for _, e := range bs {
		if e.Name == name {
			return e, true
		}
	}
	return nil, false
}

type BitmapBit struct {
	Bit         string          `json:"bit,omitempty"`
	Code        string          `json:"code,omitempty"`
	Name        string          `json:"name,omitempty"`
	Summary     string          `json:"summary,omitempty"`
	Conformance conformance.Set `json:"conformance,omitempty"`
}

func (c *BitmapBit) EntityType() types.EntityType {
	return types.EntityTypeBitmapValue
}

func (c *BitmapBit) Clone() *BitmapBit {
	nb := &BitmapBit{Bit: c.Bit, Code: c.Code, Name: c.Name, Summary: c.Summary}
	if len(c.Conformance) > 0 {
		nb.Conformance = c.Conformance.CloneSet()
	}
	return nb
}

var bitRangePattern = regexp.MustCompile(`^(?P<From>[0-9]+)(?:\.{2,}|\s*\-\s*)(?P<To>[0-9]+)$`)

func (bv *BitmapBit) Bits() (from uint64, to uint64, err error) {
	from, err = parse.HexOrDec(bv.Bit)
	if err == nil {
		to = from
		return
	}
	matches := bitRangePattern.FindStringSubmatch(bv.Bit)
	if len(matches) < 3 {
		err = fmt.Errorf("invalid bit mask range: \"%s\"", bv.Bit)
		return
	}
	from, err = parse.HexOrDec(matches[1])
	if err != nil {
		return
	}
	to, err = parse.HexOrDec(matches[2])
	if err != nil {
		return
	}
	return
}

func (bv *BitmapBit) Mask() (uint64, error) {
	from, to, err := bv.Bits()
	if err != nil {
		return 0, err
	}
	if from == to {
		return 1 << (from), nil
	}
	var val uint64
	for i := from; i <= to; i++ {
		val |= (1 << i)
	}
	return val, nil
}

func (bv *BitmapBit) GetConformance() conformance.Set {
	return bv.Conformance
}

type BitSet []*BitmapBit

func (bs BitSet) Identifier(name string) (types.Entity, bool) {
	for _, b := range bs {
		if b.Name == name {
			return b, true
		}
	}
	return nil, false
}
