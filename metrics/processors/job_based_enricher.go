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

	batchlisters "k8s.io/client-go/listers/batch/v1"
	v1listers "k8s.io/client-go/listers/core/v1"
	"k8s.io/client-go/pkg/api/v1"
	v1job "k8s.io/client-go/pkg/apis/batch/v1"
	"k8s.io/heapster/metrics/core"
)

type JobBasedEnricher struct {
	jobLister batchlisters.JobLister
	podLister v1listers.PodLister
}

func (this *JobBasedEnricher) Name() string {
	return "job_base_enricher"
}

func (this *JobBasedEnricher) Process(batch *core.DataBatch) (*core.DataBatch, error) {
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
			if len(pod.OwnerReferences) == 0 || pod.OwnerReferences[0].Kind != "Job" {
				continue
			}
			job, err := this.jobLister.Jobs(namespace).Get(pod.OwnerReferences[0].Name)
			if err != nil {
				glog.V(3).Infof("Failed to get job for %s pod from cache: %v", core.PodKey(namespace, podName), err)
				continue
			}
			addJobPodInfo(k, v, job, pod, batch, newMs)
		}
	}
	for k, v := range newMs {
		batch.MetricSets[k] = v
	}
	return batch, nil
}

func addJobPodInfo(key string, podMs *core.MetricSet,
	job *v1job.Job, pod *v1.Pod,
	batch *core.DataBatch, newMs map[string]*core.MetricSet) {
	if key == core.PodKey(pod.Namespace, pod.Name) {
		if _, ok := podMs.Labels[core.LabelJobName.Key]; !ok {
			podMs.Labels[core.LabelJobId.Key] = string(job.UID)
			podMs.Labels[core.LabelJobName.Key] = job.Name
			podMs.Labels[core.LabelJobNamespace.Key] = job.Namespace
		}
	}

	namespace := job.Namespace
	jobName := job.Name

	jobKey := core.JobKey(namespace, jobName)
	_, found := batch.MetricSets[jobKey]
	if !found {
		glog.V(2).Infof("Job %s not found, creating a stub", jobKey)
		jobMs := &core.MetricSet{
			MetricValues: make(map[string]core.MetricValue),
			Labels: map[string]string{
				core.LabelMetricSetType.Key: core.MetricSetTypeJob,
				core.LabelNamespaceName.Key: namespace,
				core.LabelJobNamespace.Key:  namespace,
				core.LabelJobName.Key:       jobName,
				core.LabelJobId.Key:         string(job.UID),
			},
		}
		newMs[jobKey] = jobMs
	}
}

func NewJobBasedEnricher(jobLister batchlisters.JobLister, podLister v1listers.PodLister) (*JobBasedEnricher, error) {
	return &JobBasedEnricher{
		jobLister: jobLister,
		podLister: podLister,
	}, nil
}
