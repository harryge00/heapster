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

// Added by luobingli
package processors

import (
	"fmt"

	"github.com/golang/glog"

	"k8s.io/heapster/metrics/core"
)

var RcLabelsToPopulate = []core.LabelDescriptor{
	core.LabelRcId,
	core.LabelRcName,
	core.LabelRcNamespace,
	core.LabelNamespaceName,
	core.LabelRcNamespaceUID,
	core.LabelHostname,
	core.LabelHostID,
}

// Provided by Kubelet/cadvisor.
var specialMetrics = []core.Metric{
	core.MetricUptime,
}

type RcAggregator struct {
	specialLabels map[string]struct{}
}

func (this *RcAggregator) Name() string {
	return "rc_aggregator"
}

func (this *RcAggregator) Process(batch *core.DataBatch) (*core.DataBatch, error) {
	newRcs := make(map[string]*core.MetricSet)

	for key, metricSet := range batch.MetricSets {
		if metricSetType, found := metricSet.Labels[core.LabelMetricSetType.Key]; found && metricSetType == core.MetricSetTypePod {
			// Aggregating pods
			rcName, found1 := metricSet.Labels[core.LabelRcName.Key]
			ns, found2 := metricSet.Labels[core.LabelRcNamespace.Key]
			if found1 && found2 {
				rcKey := core.RcKey(ns, rcName)
				rc, found := batch.MetricSets[rcKey]
				if !found {
					rc, found = newRcs[rcKey]
					if !found {
						glog.V(2).Infof("Rc not found adding %s", rcKey)
						rc = this.rcMetricSet(metricSet.Labels)
						newRcs[rcKey] = rc
					}
				}
				for metricName, metricValue := range metricSet.MetricValues {
					if _, found := this.specialLabels[metricName]; found {
						glog.V(4).Infof("Process: special metrics %s", metricName)
						rc.MetricValues[metricName] = metricValue
						continue
					}
					aggregatedValue, found := rc.MetricValues[metricName]
					if found {
						if aggregatedValue.ValueType != metricValue.ValueType {
							glog.Errorf("RcAggregator: inconsistent type in %s", metricName)
							continue
						}
						switch aggregatedValue.ValueType {
						case core.ValueInt64:
							aggregatedValue.IntValue += metricValue.IntValue
						case core.ValueFloat:
							aggregatedValue.FloatValue += metricValue.FloatValue
						default:
							return nil, fmt.Errorf("RcAggregator: type not supported in %s", metricName)
						}
					} else {
						aggregatedValue = metricValue
					}
					rc.MetricValues[metricName] = aggregatedValue
				}

				// 遍历LabeledMetrics，对filesystem metrics值做累加
				if len(rc.LabeledMetrics) == 0 {
					rc.LabeledMetrics = metricSet.LabeledMetrics
				} else {
					for _, batchLabeledMetric := range metricSet.LabeledMetrics {
						found := false
						for i, rcLabeledMetric := range rc.LabeledMetrics {
							if batchLabeledMetric.Name == rcLabeledMetric.Name {
								found = true
								if batchLabeledMetric.ValueType != rcLabeledMetric.ValueType {
									glog.Errorf("RcLabelAggregator: inconsistent type in %s", batchLabeledMetric.Name)
									continue
								}
								switch rcLabeledMetric.ValueType {
								case core.ValueInt64:
									rc.LabeledMetrics[i].IntValue += batchLabeledMetric.IntValue
								case core.ValueFloat:
									rc.LabeledMetrics[i].FloatValue += batchLabeledMetric.FloatValue
								default:
									return nil, fmt.Errorf("RcLabelAggregator: type not supported in %s", batchLabeledMetric.Name)
								}
								break
							}
						}
						if !found {
							rc.LabeledMetrics = append(rc.LabeledMetrics, batchLabeledMetric)
						}
					}
				}
				glog.V(6).Infof("metricSet.LabeledMetrics:%v", metricSet.LabeledMetrics)
				glog.V(6).Infof("rc.LabeledMetrics:%v", rc.LabeledMetrics)
			} else {
				glog.Errorf("No namespace and/or rc info in pod %s: %v", key, metricSet.Labels)
				continue
			}
		}
	}

	for key, val := range newRcs {
		batch.MetricSets[key] = val
	}
	return batch, nil
}

func (this *RcAggregator) rcMetricSet(labels map[string]string) *core.MetricSet {
	newLabels := map[string]string{
		core.LabelMetricSetType.Key: core.MetricSetTypeRc,
	}
	for _, l := range RcLabelsToPopulate {
		if val, ok := labels[l.Key]; ok {
			newLabels[l.Key] = val
		}
	}

	return &core.MetricSet{
		MetricValues: make(map[string]core.MetricValue),
		Labels:       newLabels,
	}
}

func NewRcAggregator() *RcAggregator {
	metrics := make(map[string]struct{})
	for _, metric := range specialMetrics {
		metrics[metric.MetricDescriptor.Name] = struct{}{}
	}
	return &RcAggregator{
		specialLabels: metrics,
	}
}
