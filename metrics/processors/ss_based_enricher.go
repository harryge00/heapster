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

type SsBasedEnricher struct {
	ssLister  *cache.StoreToReplicationControllerLister
	podLister *cache.StoreToPodLister
}

func (this *SsBasedEnricher) Name() string {
	return "ss_base_enricher"
}

func (this *SsBasedEnricher) Process(batch *core.DataBatch) (*core.DataBatch, error) {
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
			sss, err := this.getPodControllers(pod)
			if (err != nil) || (len(sss) == 0) {
				glog.V(3).Infof("Failed to get replicationControllers for %s pod from cache: %v", core.PodKey(namespace, podName), err)
				continue
			}
			for _, ss := range sss {
				addSsPodInfo(k, v, &ss, pod, batch, newMs)
			}
		}
	}
	for k, v := range newMs {
		batch.MetricSets[k] = v
	}
	return batch, nil
}

func (this *SsBasedEnricher) getPodControllers(pod *kube_api.Pod) ([]kube_api.ReplicationController, error) {
	var sss []kube_api.ReplicationController
	sss, err := this.ssLister.GetPodControllers(pod)
	if err != nil {
		return nil, err
	}
	if len(sss) == 0 {
		return nil, fmt.Errorf("cannot find ss")
	}

	return sss, nil
}

func (this *SsBasedEnricher) getPod(namespace, name string) (*kube_api.Pod, error) {
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

func addSsPodInfo(key string, podMs *core.MetricSet,
	ss *kube_api.ReplicationController, pod *kube_api.Pod,
	batch *core.DataBatch, newMs map[string]*core.MetricSet) {
	if key == core.PodKey(pod.Namespace, pod.Name) {
		if _, ok := podMs.Labels[core.LabelSsName.Key]; !ok {
			podMs.Labels[core.LabelSsId.Key] = string(ss.UID)
			podMs.Labels[core.LabelSsName.Key] = ss.Name
			podMs.Labels[core.LabelSsNamespace.Key] = ss.Namespace
		}
	}

	namespace := ss.Namespace
	ssName := ss.Name

	ssKey := core.SsKey(namespace, ssName)
	_, found := batch.MetricSets[ssKey]
	if !found {
		glog.V(2).Infof("Ss %s not found, creating a stub", ssKey)
		ssMs := &core.MetricSet{
			MetricValues: make(map[string]core.MetricValue),
			Labels: map[string]string{
				core.LabelMetricSetType.Key: core.MetricSetTypeSs,
				core.LabelNamespaceName.Key: namespace,
				core.LabelSsNamespace.Key:   namespace,
				core.LabelSsName.Key:        ssName,
				core.LabelSsId.Key:          string(ss.UID),
				core.LabelLabels.Key:        util.LabelsToString(ss.Labels, ","),
			},
		}
		newMs[ssKey] = ssMs
	}
}

func NewSsBasedEnricher(ssLister *cache.StoreToReplicationControllerLister, podLister *cache.StoreToPodLister) (*SsBasedEnricher, error) {
	return &SsBasedEnricher{
		ssLister:  ssLister,
		podLister: podLister,
	}, nil
}
