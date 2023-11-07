package render

import (
	"encoding/xml"
	"fmt"

	"github.com/hasty/matterfmt/matter"
	"github.com/hasty/matterfmt/parse"
)

func (r *renderer) amendEnum(d xmlDecoder, e xmlEncoder, el xml.StartElement, cluster *matter.Cluster, clusterIDs []string, enums map[*matter.Enum]struct{}) (err error) {
	name := getAttributeValue(el.Attr, "name")

	var matchingEnum *matter.Enum
	for en := range enums {
		if en.Name == name {
			matchingEnum = en
			delete(enums, en)
			break
		}
	}

	if matchingEnum == nil {
		return writeThrough(d, e, el)
	}

	Ignore(d, "enum")

	return r.writeEnum(e, el, matchingEnum, clusterIDs)
}

func (r *renderer) writeEnum(e xmlEncoder, el xml.StartElement, en *matter.Enum, clusterIDs []string) (err error) {
	xfb := el.Copy()
	xfb.Attr = setAttributeValue(xfb.Attr, "name", en.Name)
	xfb.Attr = setAttributeValue(xfb.Attr, "type", en.Type)
	err = e.EncodeToken(xfb)
	if err != nil {
		return
	}
	err = r.renderClusterCodes(e, clusterIDs)
	if err != nil {
		return
	}

	for _, v := range en.Values {
		if v.Conformance == "Zigbee" {
			continue
		}

		val := v.Value
		valNum, er := parse.HexOrDec(val)
		if er == nil {
			val = fmt.Sprintf("%#02x", valNum)
		}

		elName := xml.Name{Local: "item"}
		xfs := xml.StartElement{Name: elName}
		xfs.Attr = setAttributeValue(xfs.Attr, "name", v.Name)
		xfs.Attr = setAttributeValue(xfs.Attr, "value", val)
		err = e.EncodeToken(xfs)
		if err != nil {
			return
		}
		xfe := xml.EndElement{Name: elName}
		err = e.EncodeToken(xfe)
		if err != nil {
			return
		}

	}
	return e.EncodeToken(xml.EndElement{Name: xfb.Name})
}
