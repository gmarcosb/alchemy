package compare

import (
	"fmt"
	"strings"

	"github.com/hasty/matterfmt/matter"
)

func compareModels(specModels map[string][]any, zapModels map[string][]any) (diffs []any, err error) {
	for path, sm := range specModels {
		zm, ok := zapModels[path]
		if !ok {
			fmt.Printf("path %s missing from ZAP models\n", path)
			continue
		}
		fmt.Printf("path %s found in ZAP models\n", path)

		specClusters := make(map[string]*matter.Cluster)
		for _, m := range sm {
			switch v := m.(type) {
			case *matter.Cluster:
				fmt.Printf("adding spec cluster: %s\n", v.ID)
				specClusters[strings.ToLower(v.ID)] = v
			default:
				fmt.Printf("unexpected spec model: %T\n", m)
			}
		}
		zapClusters := make(map[string]*matter.Cluster)
		for _, m := range zm {
			switch v := m.(type) {
			case *matter.Cluster:
				fmt.Printf("adding ZAP cluster: %s\n", v.ID)
				zapClusters[strings.ToLower(v.ID)] = v
			default:
				fmt.Printf("unexpected ZAP model: %T\n", m)
			}

		}
		delete(zapModels, path)
		for cid, sc := range specClusters {
			if zc, ok := zapClusters[cid]; ok {
				var clusterDiffs *ClusterDifferences
				clusterDiffs, err = compareClusters(sc, zc)
				if err != nil {
					fmt.Printf("unable to compare clusters (%s): %v\n", cid, err)
					err = nil
				} else if clusterDiffs != nil {
					diffs = append(diffs, clusterDiffs)
				}
				delete(zapClusters, cid)
			} else {
				fmt.Printf("ZAP cluster %s missing from %s; ", cid, path)
				for zid := range zapClusters {
					fmt.Printf("have %s,", zid)
				}
				fmt.Println()
			}
		}
		for cid := range zapClusters {
			fmt.Printf("Spec cluster %s missing from %s\n", cid, path)
		}
	}

	for path := range zapModels {
		fmt.Printf("path %s missing from Spec models\n", path)
	}
	return
}
