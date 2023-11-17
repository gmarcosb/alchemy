package render

import (
	"github.com/beevik/etree"
	"github.com/hasty/alchemy/matter"
	"github.com/hasty/alchemy/zap"
)

func renderDataTypes(cluster *matter.Cluster, clusters []*matter.Cluster, cx *etree.Element, errata *Errata) {
	var clusterIDs []string
	for _, cluster := range clusters {
		clusterIDs = append(clusterIDs, cluster.ID.HexString())
	}
	for _, s := range errata.dataTypeOrder {
		switch s {
		case matter.SectionDataTypeBitmap:
			renderBitmaps(cluster.Bitmaps, clusterIDs, cx)
		case matter.SectionDataTypeEnum:
			renderEnums(cluster.Enums, clusterIDs, cx)
		case matter.SectionDataTypeStruct:
			renderStructs(cluster.Structs, clusterIDs, cx)
		}
	}
}

func writeAttributeDataType(x *etree.Element, dt *matter.DataType) {
	if dt == nil {
		return
	}
	dts := zap.ConvertDataTypeNameToZap(dt.Name)
	if dt.IsArray {
		x.CreateAttr("type", "ARRAY")
		x.CreateAttr("entryType", dts)
	} else {
		x.CreateAttr("type", dts)
	}
}

func writeDataType(x *etree.Element, dt *matter.DataType) {
	if dt == nil {
		return
	}
	dts := zap.ConvertDataTypeNameToZap(dt.Name)
	if dt.IsArray {
		x.CreateAttr("array", "true")
	}
	x.CreateAttr("type", dts)
}
