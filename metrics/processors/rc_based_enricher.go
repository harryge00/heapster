// Copyright 2015 Google Inc. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package processors

import (
	"fmt"

	"github.com/golang/glog"

	"k8s.io/heapster/metrics/util"

	"k8s.io/heapster/metrics/core"
	kube_api "k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/client/cache"
)

type RcBasedEnricher struct {
	rcLister  *cache.StoreToReplicationControllerLister
	podLister *cache.StoreToPodLister
}

func (this *RcBasedEnricher) Name() string {
	return "rc_base_enricher"
}

func (this *RcBasedEnricher) Process(batch *core.DataBatch) (*core.DataBatch, error) {
	newMs := make(map[string]*core.MetricSet, len(batch.MetricSets))
	for k, v := range batch.MetricSets {
		if v.Labels[core.LabelMetricSetType.Key] == core.MetricSetTypePod {
			namespace := v.Labels[core.LabelNamespaceName.Key]
			podName := v.Labels[core.LabelPodName.Key]
			pod, err := this.getPod(namespace, podName)
			if err != nil {
				glog.V(3).Infof("Failed to get pod %s from cache: %v", core.PodKey(namespace, podName), err)
				continue
			}
			rcs, err := this.getPodControllers(pod)
			if (err != nil) || (len(rcs) == 0) {
				glog.V(3).Infof("Failed to get replicationControllers for %s pod from cache: %v", core.PodKey(namespace, podName), err)
				continue
			}
			for _, rc := range rcs {
				addRcPodInfo(k, v, &rc, pod, batch, newMs)
			}
		}
	}
	for k, v := range newMs {
		batch.MetricSets[k] = v
	}
	return batch, nil
}

func (this *RcBasedEnricher) getPodControllers(pod *kube_api.Pod) ([]kube_api.ReplicationController, error) {
	var rcs []kube_api.ReplicationController
	rcs, err := this.rcLister.GetPodControllers(pod)
	if err != nil {
		return nil, err
	}
	if len(rcs) == 0 {
		return nil, fmt.Errorf("cannot find rc")
	}

	return rcs, nil
}

func (this *RcBasedEnricher) getPod(namespace, name string) (*kube_api.Pod, error) {
	o, exists, err := this.podLister.Get(
		&kube_api.Pod{
			ObjectMeta: kube_api.ObjectMeta{
				Namespace: namespace,
				Name:      name,
			},
		},
	)
	if err != nil {
		return nil, err
	}
	if !exists || o == nil {
		return nil, fmt.Errorf("cannot find pod definition")
	}
	pod, ok := o.(*kube_api.Pod)
	if !ok {
		return nil, fmt.Errorf("cache contains wrong type")
	}
	return pod, nil
}

func addRcPodInfo(key string, podMs *core.MetricSet,
	rc *kube_api.ReplicationController, pod *kube_api.Pod,
	batch *core.DataBatch, newMs map[string]*core.MetricSet) {
	if key == core.PodKey(pod.Namespace, pod.Name) {
		if _, ok := podMs.Labels[core.LabelRcName.Key]; !ok {
			podMs.Labels[core.LabelRcId.Key] = string(rc.UID)
			podMs.Labels[core.LabelRcName.Key] = rc.Name
			podMs.Labels[core.LabelRcNamespace.Key] = rc.Namespace
		}
	}

	namespace := rc.Namespace
	rcName := rc.Name

	rcKey := core.RcKey(namespace, rcName)
	_, found := batch.MetricSets[rcKey]
	if !found {
		glog.V(2).Infof("Rc %s not found, creating a stub", rcKey)
		rcMs := &core.MetricSet{
			MetricValues: make(map[string]core.MetricValue),
			Labels: map[string]string{
				core.LabelMetricSetType.Key: core.MetricSetTypeRc,
				core.LabelNamespaceName.Key: namespace,
				core.LabelRcNamespace.Key:   namespace,
				core.LabelRcName.Key:        rcName,
				core.LabelRcId.Key:          string(rc.UID),
				core.LabelLabels.Key:        util.LabelsToString(rc.Labels, ","),
			},
		}
		newMs[rcKey] = rcMs
	}
}

func NewRcBasedEnricher(rcLister *cache.StoreToReplicationControllerLister, podLister *cache.StoreToPodLister) (*RcBasedEnricher, error) {
	return &RcBasedEnricher{
		rcLister:  rcLister,
		podLister: podLister,
	}, nil
}
