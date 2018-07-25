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

// Added by zhuzhen
package processors

import (
	"fmt"
	"github.com/golang/glog"
	"k8s.io/heapster/metrics/core"
)

var JobLabelsToPopulate = []core.LabelDescriptor{
	core.LabelJobId,
	core.LabelJobName,
	core.LabelJobNamespace,
	core.LabelNamespaceName,
	core.LabelJobNamespaceUID,
	core.LabelHostname,
	core.LabelHostID,
}

type JobAggregator struct {
	specialLabels map[string]struct{}
}

func (this *JobAggregator) Name() string {
	return "job_aggregator"
}

func (this *JobAggregator) Process(batch *core.DataBatch) (*core.DataBatch, error) {
	newJobs := make(map[string]*core.MetricSet)

	for key, metricSet := range batch.MetricSets {
		if metricSetType, found := metricSet.Labels[core.LabelMetricSetType.Key]; found && metricSetType == core.MetricSetTypePod {
			// Aggregating pods
			jobName, found1 := metricSet.Labels[core.LabelJobName.Key]
			ns, found2 := metricSet.Labels[core.LabelJobNamespace.Key]
			if found1 && found2 {
				jobKey := core.JobKey(ns, jobName)
				job, found := batch.MetricSets[jobKey]
				if !found {
					job, found = newJobs[jobKey]
					if !found {
						glog.V(2).Infof("Job not found adding %s", jobKey)
						job = this.jobMetricSet(metricSet.Labels)
						newJobs[jobKey] = job
					}
				}
				for metricName, metricValue := range metricSet.MetricValues {
					if _, found := this.specialLabels[metricName]; found {
						glog.V(4).Infof("Process: special metrics %s", metricName)
						job.MetricValues[metricName] = metricValue
						continue
					}
					aggregatedValue, found := job.MetricValues[metricName]
					if found {
						if aggregatedValue.ValueType != metricValue.ValueType {
							glog.Errorf("JobAggregator: inconsistent type in %s", metricName)
							continue
						}
						switch aggregatedValue.ValueType {
						case core.ValueInt64:
							aggregatedValue.IntValue += metricValue.IntValue
						case core.ValueFloat:
							aggregatedValue.FloatValue += metricValue.FloatValue
						default:
							return nil, fmt.Errorf("JobAggregator: type not supported in %s", metricName)
						}
					} else {
						aggregatedValue = metricValue
					}
					job.MetricValues[metricName] = aggregatedValue
				}

				// 遍历LabeledMetrics，对filesystem metrics值做累加
				if len(job.LabeledMetrics) == 0 {
					job.LabeledMetrics = metricSet.LabeledMetrics
				} else {
					for _, batchLabeledMetric := range metricSet.LabeledMetrics {
						found := false
						for i, jobLabeledMetric := range job.LabeledMetrics {
							if batchLabeledMetric.Name == jobLabeledMetric.Name {
								found = true
								if batchLabeledMetric.ValueType != jobLabeledMetric.ValueType {
									glog.Errorf("JobLabelAggregator: inconsistent type in %s", batchLabeledMetric.Name)
									continue
								}
								switch jobLabeledMetric.ValueType {
								case core.ValueInt64:
									job.LabeledMetrics[i].IntValue += batchLabeledMetric.IntValue
								case core.ValueFloat:
									job.LabeledMetrics[i].FloatValue += batchLabeledMetric.FloatValue
								default:
									return nil, fmt.Errorf("JobLabelAggregator: type not supported in %s", batchLabeledMetric.Name)
								}
								break
							}
						}
						if !found {
							job.LabeledMetrics = append(job.LabeledMetrics, batchLabeledMetric)
						}
					}
				}
				glog.V(6).Infof("metricSet.LabeledMetrics:%v", metricSet.LabeledMetrics)
				glog.V(6).Infof("job.LabeledMetrics:%v", job.LabeledMetrics)
			} else {
				glog.Errorf("No namespace and/or job info in pod %s: %v", key, metricSet.Labels)
				continue
			}
		}
	}

	for key, val := range newJobs {
		batch.MetricSets[key] = val
	}
	return batch, nil
}

func (this *JobAggregator) jobMetricSet(labels map[string]string) *core.MetricSet {
	newLabels := map[string]string{
		core.LabelMetricSetType.Key: core.MetricSetTypeJob,
	}
	for _, l := range JobLabelsToPopulate {
		if val, ok := labels[l.Key]; ok {
			newLabels[l.Key] = val
		}
	}

	return &core.MetricSet{
		MetricValues: make(map[string]core.MetricValue),
		Labels:       newLabels,
	}
}

func NewJobAggregator() *JobAggregator {
	metrics := make(map[string]struct{})
	for _, metric := range specialMetrics {
		metrics[metric.MetricDescriptor.Name] = struct{}{}
	}
	return &JobAggregator{
		specialLabels: metrics,
	}
}
