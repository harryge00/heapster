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
	"github.com/golang/glog"

	v1listers "k8s.io/client-go/listers/core/v1"
	v2listers "k8s.io/client-go/listers/apps/v1beta2"
	apps "k8s.io/api/apps/v1beta2"

	"k8s.io/client-go/pkg/api/v1"
	"k8s.io/heapster/metrics/core"
)

type SsBasedEnricher struct {
	ssLister  v2listers.StatefulSetLister
	podLister v1listers.PodLister
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
			pod, err := this.podLister.Pods(namespace).Get(podName)
			if err != nil {
				glog.V(3).Infof("Failed to get pod %s from cache: %v", core.PodKey(namespace, podName), err)
				continue
			}
			sss, err := this.ssLister.GetPodStatefulSets(pod)
			if (err != nil) || (len(sss) == 0) {
				glog.V(3).Infof("Failed to get statefulset for %s pod from cache: %v", core.PodKey(namespace, podName), err)
				continue
			}
			for _, ss := range sss {
				addSsPodInfo(k, v, ss, pod, batch, newMs)
			}
		}
	}
	for k, v := range newMs {
		batch.MetricSets[k] = v
	}
	return batch, nil
}

func addSsPodInfo(key string, podMs *core.MetricSet,
	ss *apps.StatefulSet, pod *v1.Pod,
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
			},
		}
		newMs[ssKey] = ssMs
	}
}

func NewSsBasedEnricher(ssLister v2listers.StatefulSetLister, podLister v1listers.PodLister) (*SsBasedEnricher, error) {
	return &SsBasedEnricher{
		ssLister:  ssLister,
		podLister: podLister,
	}, nil
}
