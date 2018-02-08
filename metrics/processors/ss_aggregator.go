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

var SatefulsetLabelsToPopulate = []core.LabelDescriptor{
	core.LabelSsId,
	core.LabelSsName,
	core.LabelSsNamespace,
	core.LabelNamespaceName,
	core.LabelSsNamespaceUID,
	core.LabelHostname,
	core.LabelHostID,
}

type SsAggregator struct {
	specialLabels map[string]struct{}
}

func (this *SsAggregator) Name() string {
	return "Ss_aggregator"
}

func (this *SsAggregator) Process(batch *core.DataBatch) (*core.DataBatch, error) {
	newSss := make(map[string]*core.MetricSet)

	for key, metricSet := range batch.MetricSets {
		if metricSetType, found := metricSet.Labels[core.LabelMetricSetType.Key]; found && metricSetType == core.MetricSetTypePod {
			// Aggregating pods
			ssName, found1 := metricSet.Labels[core.LabelSsName.Key]
			ns, found2 := metricSet.Labels[core.LabelSsNamespace.Key]
			if found1 && found2 {
				ssKey := core.SsKey(ns, ssName)
				ss, found := batch.MetricSets[ssKey]
				if !found {
					ss, found = newSss[ssKey]
					if !found {
						glog.V(2).Infof("Ss not found adding %s", ssKey)
						ss = this.ssMetricSet(metricSet.Labels)
						newSss[ssKey] = ss
					}
				}
				for metricName, metricValue := range metricSet.MetricValues {
					if _, found := this.specialLabels[metricName]; found {
						glog.V(4).Infof("Process: special metrics %s", metricName)
						ss.MetricValues[metricName] = metricValue
						continue
					}
					aggregatedValue, found := ss.MetricValues[metricName]
					if found {
						if aggregatedValue.ValueType != metricValue.ValueType {
							glog.Errorf("SsAggregator: inconsistent type in %s", metricName)
							continue
						}
						switch aggregatedValue.ValueType {
						case core.ValueInt64:
							aggregatedValue.IntValue += metricValue.IntValue
						case core.ValueFloat:
							aggregatedValue.FloatValue += metricValue.FloatValue
						default:
							return nil, fmt.Errorf("SsAggregator: type not supported in %s", metricName)
						}
					} else {
						aggregatedValue = metricValue
					}
					ss.MetricValues[metricName] = aggregatedValue
				}

				// 遍历LabeledMetrics，对filesystem metrics值做累加
				if len(ss.LabeledMetrics) == 0 {
					ss.LabeledMetrics = metricSet.LabeledMetrics
				} else {
					for _, batchLabeledMetric := range metricSet.LabeledMetrics {
						found := false
						for i, ssLabeledMetric := range ss.LabeledMetrics {
							if batchLabeledMetric.Name == ssLabeledMetric.Name {
								found = true
								if batchLabeledMetric.ValueType != ssLabeledMetric.ValueType {
									glog.Errorf("SsLabelAggregator: inconsistent type in %s", batchLabeledMetric.Name)
									continue
								}
								switch ssLabeledMetric.ValueType {
								case core.ValueInt64:
									ss.LabeledMetrics[i].IntValue += batchLabeledMetric.IntValue
								case core.ValueFloat:
									ss.LabeledMetrics[i].FloatValue += batchLabeledMetric.FloatValue
								default:
									return nil, fmt.Errorf("SsLabelAggregator: type not supported in %s", batchLabeledMetric.Name)
								}
								break
							}
						}
						if !found {
							ss.LabeledMetrics = append(ss.LabeledMetrics, batchLabeledMetric)
						}
					}
				}
				glog.V(6).Infof("metricSet.LabeledMetrics:%v", metricSet.LabeledMetrics)
				glog.V(6).Infof("ss.LabeledMetrics:%v", ss.LabeledMetrics)
			} else {
				glog.Errorf("No namespace and/or ss info in pod %s: %v", key, metricSet.Labels)
				continue
			}
		}
	}

	for key, val := range newSss {
		batch.MetricSets[key] = val
	}
	return batch, nil
}

func (this *SsAggregator) ssMetricSet(labels map[string]string) *core.MetricSet {
	newLabels := map[string]string{
		core.LabelMetricSetType.Key: core.MetricSetTypeSs,
	}
	for _, l := range SatefulsetLabelsToPopulate {
		if val, ok := labels[l.Key]; ok {
			newLabels[l.Key] = val
		}
	}

	return &core.MetricSet{
		MetricValues: make(map[string]core.MetricValue),
		Labels:       newLabels,
	}
}

func NewSsAggregator() *SsAggregator {
	metrics := make(map[string]struct{})
	for _, metric := range specialMetrics {
		metrics[metric.MetricDescriptor.Name] = struct{}{}
	}
	return &SsAggregator{
		specialLabels: metrics,
	}
}
