package ascii

import (
	"fmt"
	"log/slog"
	"path/filepath"

	"github.com/hasty/alchemy/matter"
	"github.com/hasty/alchemy/matter/types"
)

type entityIndex map[string]map[types.Entity]*matter.Cluster

func (mi entityIndex) addEntity(name string, entity types.Entity, cluster *matter.Cluster) {
	m, ok := mi[name]
	if !ok {
		m = make(map[types.Entity]*matter.Cluster)
		mi[name] = m
	}
	m[entity] = cluster
}

func BuildSpec(docs []*Doc) (spec *matter.Spec, err error) {

	buildTree(docs)

	spec = &matter.Spec{

		ClustersByID:   make(map[uint64]*matter.Cluster),
		ClustersByName: make(map[string]*matter.Cluster),
		DeviceTypes:    make(map[uint64]*matter.DeviceType),
		ClusterRefs:    make(map[types.Entity]map[*matter.Cluster]struct{}),
		DocRefs:        make(map[types.Entity]string),

		Bitmaps: make(map[string]*matter.Bitmap),
		Enums:   make(map[string]*matter.Enum),
		Structs: make(map[string]*matter.Struct),
	}
	entityIndex := make(entityIndex)

	for _, d := range docs {
		slog.Debug("building spec", "path", d.Path)
		var entities []types.Entity
		entities, err = d.Entities()
		if err != nil {
			slog.Warn("error building entities", "doc", d.Path, "error", err)
			continue
		}
		fileName := filepath.Base(d.Path)
		switch fileName {
		case "BaseDeviceType.adoc":
			spec.BaseDeviceType, err = d.toBaseDeviceType()
			if err != nil {
				return
			}
			continue
		}
		for _, m := range entities {
			switch m := m.(type) {
			case *matter.Cluster:
				if m.ID.Valid() {
					spec.ClustersByID[m.ID.Value()] = m
				}
				spec.ClustersByName[m.Name] = m
				for _, en := range m.Bitmaps {
					_, ok := spec.Bitmaps[en.Name]
					if ok {
						slog.Debug("multiple bitmaps with same name", "name", en.Name)
					} else {
						spec.Bitmaps[en.Name] = en
					}
					spec.DocRefs[en] = d.Path
					entityIndex.addEntity(en.Name, en, m)
				}
				for _, en := range m.Enums {
					_, ok := spec.Enums[en.Name]
					if ok {
						slog.Debug("multiple enums with same name", "name", en.Name)
					} else {
						spec.Enums[en.Name] = en
					}
					spec.DocRefs[en] = d.Path
					entityIndex.addEntity(en.Name, en, m)
				}
				for _, en := range m.Structs {
					_, ok := spec.Structs[en.Name]
					if ok {
						slog.Debug("multiple structs with same name", "name", en.Name)
					} else {
						spec.Structs[en.Name] = en
					}
					spec.DocRefs[en] = d.Path
					entityIndex.addEntity(en.Name, en, m)
				}
			case *matter.DeviceType:
				spec.DeviceTypes[m.ID.Value()] = m
			case *matter.Bitmap:
				_, ok := spec.Bitmaps[m.Name]
				if ok {
					slog.Debug("multiple bitmaps with same name", "name", m.Name)
				} else {

					spec.Bitmaps[m.Name] = m
				}
				entityIndex.addEntity(m.Name, m, nil)
			case *matter.Enum:
				_, ok := spec.Enums[m.Name]
				if ok {
					slog.Debug("multiple enums with same name", "name", m.Name)
				} else {
					spec.Enums[m.Name] = m
				}
				entityIndex.addEntity(m.Name, m, nil)
			case *matter.Struct:
				_, ok := spec.Structs[m.Name]
				if ok {
					slog.Debug("multiple structs with same name", "name", m.Name)
				} else {
					spec.Structs[m.Name] = m
				}
				entityIndex.addEntity(m.Name, m, nil)
			default:
				slog.Warn("unknown entity type", "type", fmt.Sprintf("%T", m))
			}
			spec.DocRefs[m] = d.Path
		}
	}

	resolveHierarchy(spec)
	resolveDataTypeReferences(spec, entityIndex)

	for _, c := range spec.ClustersByID {
		for _, en := range c.Enums {
			switch en.Name {
			case "ModeTag":
				spec.ClusterRefs.Add(c, en)
			}
		}
	}

	for _, dt := range spec.DeviceTypes {
		for _, cr := range dt.ClusterRequirements {
			if c, ok := spec.ClustersByID[cr.ID.Value()]; ok {
				cr.Cluster = c
			}
		}
		for _, er := range dt.ElementRequirements {
			if c, ok := spec.ClustersByID[er.ID.Value()]; ok {
				er.Cluster = c
			}
		}
	}
	return
}

func resolveDataTypeReferences(spec *matter.Spec, mi entityIndex) {
	for _, s := range spec.Structs {
		for _, f := range s.Fields {
			resolveDataType(spec, mi, nil, f, f.Type)
		}
	}
	for _, cluster := range spec.ClustersByID {
		for _, a := range cluster.Attributes {
			if a.Type == nil {
				continue
			}
			resolveDataType(spec, mi, cluster, a, a.Type)
		}
		for _, s := range cluster.Structs {
			for _, f := range s.Fields {
				resolveDataType(spec, mi, cluster, f, f.Type)
			}
		}
		for _, s := range cluster.Events {
			for _, f := range s.Fields {
				resolveDataType(spec, mi, cluster, f, f.Type)
			}
		}
		for _, s := range cluster.Commands {
			for _, f := range s.Fields {
				resolveDataType(spec, mi, cluster, f, f.Type)
			}
		}
	}
}

func resolveDataType(spec *matter.Spec, mi entityIndex, cluster *matter.Cluster, field *matter.Field, dataType *types.DataType) {
	if dataType == nil {
		slog.Warn("missing type on field", "name", field.Name)
		return
	}
	switch dataType.BaseType {
	case types.BaseDataTypeList:
		resolveDataType(spec, mi, cluster, field, dataType.EntryType)
	case types.BaseDataTypeCustom:
		if dataType.Entity == nil {
			entities := mi[dataType.Name]
			if len(entities) == 0 {
				slog.Warn("unknown custom data type", "cluster", clusterName(cluster), "field", field.Name, "type", dataType.Name)
			} else if len(entities) == 1 {
				for m := range entities {
					dataType.Entity = m
					break
				}
			} else {
				for m, c := range entities {
					if c == cluster { // If there are multiple entities with the same name, prefer the one on the current cluster
						dataType.Entity = m
						break
					}
				}
				if dataType.Entity == nil {
					// OK, if the data type is defined on the direct parent of this cluster, take that one
					if cluster.Hierarchy != "Base" {
						for m, c := range entities {
							if c != nil && c.Name == cluster.Hierarchy {
								dataType.Entity = m
								break
							}
						}
					}
					if dataType.Entity == nil {
						// Can't disambiguate out this data model
						slog.Warn("ambiguous data type", "cluster", clusterName(cluster), "field", field.Name, "type", dataType.Name)
						for m, c := range entities {
							var clusterName string
							if c != nil {
								clusterName = c.Name
							} else {
								clusterName = "naked"
							}
							slog.Warn("ambiguous data type", "model", m, "cluster", clusterName)
						}
					}
				}
			}
		}
		if cluster == nil || dataType.Entity == nil {
			return
		}
		spec.ClusterRefs.Add(cluster, dataType.Entity)
		slog.Debug("setting cluster", "name", cluster.Name, "type", dataType.Name)
		s, ok := dataType.Entity.(*matter.Struct)
		if !ok {
			return
		}
		for _, f := range s.Fields {
			resolveDataType(spec, mi, cluster, f, f.Type)
		}
	}
}

func clusterName(cluster *matter.Cluster) string {
	if cluster != nil {
		return cluster.Name
	}
	return "none"
}

func resolveHierarchy(spec *matter.Spec) {
	for _, c := range spec.ClustersByID {
		if c.Hierarchy == "Base" {
			continue
		}
		base, ok := spec.ClustersByName[c.Hierarchy]
		if !ok {
			slog.Warn("Failed to find base cluster", "cluster", c.Name, "baseCluster", c.Hierarchy)
			continue
		}
		base.Base = true
		err := c.Inherit(base)
		if err != nil {
			slog.Warn("Failed to inherit from base cluster", "cluster", c.Name, "baseCluster", c.Hierarchy, "error", err)
		}

	}
}
