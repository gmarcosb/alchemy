package dm

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"sync"

	"github.com/hasty/alchemy/internal/files"
	"github.com/hasty/alchemy/internal/pipeline"
	"github.com/hasty/alchemy/matter"
	"github.com/hasty/alchemy/matter/spec"
	"github.com/hasty/alchemy/matter/types"
	"github.com/iancoleman/orderedmap"
	"github.com/iancoleman/strcase"
)

type Renderer struct {
	sdkRoot string

	clusters     []*matter.Cluster
	clustersLock sync.Mutex
}

func NewRenderer(sdkRoot string) *Renderer {
	return &Renderer{sdkRoot: sdkRoot}
}

func (p *Renderer) Name() string {
	return "Saving data model"
}

func (p *Renderer) Type() pipeline.ProcessorType {
	return pipeline.ProcessorTypeIndividual
}

func (p *Renderer) Process(cxt context.Context, input *pipeline.Data[*spec.Doc], index int32, total int32) (outputs []*pipeline.Data[string], extra []*pipeline.Data[*spec.Doc], err error) {
	doc := input.Content
	entites, err := doc.Entities()
	if err != nil {
		slog.ErrorContext(cxt, "error converting doc to entities", "doc", doc.Path, "error", err)
		err = nil
		return
	}
	var appClusters []types.Entity
	var deviceTypes []*matter.DeviceType
	for _, e := range entites {
		switch e := e.(type) {
		case *matter.ClusterGroup, *matter.Cluster:
			appClusters = append(appClusters, e)
		case *matter.DeviceType:
			deviceTypes = append(deviceTypes, e)
		}
	}

	if len(appClusters) == 1 {
		var s string
		switch e := appClusters[0].(type) {
		case *matter.ClusterGroup:
			if len(e.Clusters) == 0 {
				err = fmt.Errorf("empty cluster group %s", doc.Path)
				return
			}
			s, err = p.renderAppCluster(doc, e.Clusters...)
		case *matter.Cluster:
			s, err = p.renderAppCluster(doc, e)
		}
		if err != nil {
			err = fmt.Errorf("failed rendering app clusters %s: %w", doc.Path, err)
			return
		}
		outputs = append(outputs, &pipeline.Data[string]{Path: getAppClusterPath(p.sdkRoot, doc.Path, ""), Content: s})
	} else if len(appClusters) > 1 {
		for _, e := range appClusters {
			var s string
			var clusterName string
			switch e := e.(type) {
			case *matter.ClusterGroup:
				s, err = p.renderAppCluster(doc, e.Clusters...)
				clusterName = e.Clusters[0].Name
			case *matter.Cluster:
				s, err = p.renderAppCluster(doc, e)
				clusterName = e.Name
			}
			if err != nil {
				err = fmt.Errorf("failed rendering app clusters %s: %w", doc.Path, err)
				return
			}
			clusterName = strcase.ToCamel(clusterName + " Cluster")
			outputs = append(outputs, &pipeline.Data[string]{Path: getAppClusterPath(p.sdkRoot, doc.Path, clusterName), Content: s})
		}
	}

	if len(deviceTypes) > 0 {
		var s string
		s, err = renderDeviceType(doc, deviceTypes)
		if err != nil {
			err = fmt.Errorf("failed rendering device types %s: %w", doc.Path, err)
			return
		}
		outputs = append(outputs, &pipeline.Data[string]{Path: getDeviceTypePath(p.sdkRoot, doc.Path), Content: s})
	}
	for _, o := range outputs {
		o.Content, err = patchLicense(o.Content, o.Path)
		if err != nil {
			err = fmt.Errorf("error patching license for %s: %w", o.Path, err)
			return
		}
	}
	return
}

func (p *Renderer) GenerateClusterIDsJson() (*pipeline.Data[string], error) {

	clusters := make(map[uint64]string)

	path := filepath.Join(p.sdkRoot, "/data_model/clusters/cluster_ids.json")

	exists, err := files.Exists(path)
	if err != nil {
		return nil, err
	}
	if exists {
		var clusterListBytes []byte
		clusterListBytes, err = os.ReadFile(path)
		if err != nil {
			return nil, err
		}
		var clusterList map[string]any
		err = json.Unmarshal(clusterListBytes, &clusterList)
		if err != nil {
			return nil, err
		}
		for id, name := range clusterList {
			mid := matter.ParseNumber(id)
			if mid.Valid() {
				clusters[mid.Value()] = name.(string)
			}
		}
	}

	p.clustersLock.Lock()
	defer p.clustersLock.Unlock()
	for _, c := range p.clusters {
		if c.ID.Valid() {
			clusters[c.ID.Value()] = c.Name
		}
	}

	var clusterIDs []uint64
	for id := range clusters {
		clusterIDs = append(clusterIDs, id)
	}

	slices.Sort(clusterIDs)
	o := orderedmap.New()
	for _, cid := range clusterIDs {
		name := clusters[cid]
		id := strconv.FormatUint(cid, 10)
		o.Set(id, name)

	}
	b, err := json.MarshalIndent(o, "", "    ")
	if err != nil {
		err = fmt.Errorf("error marshaling cluster ID json: %w", err)
		return nil, err
	}
	return pipeline.NewData(path, string(b)), nil
}
