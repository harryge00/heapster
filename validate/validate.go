// Copyright 2014 Google Inc. All Rights Reserved.
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

package validate

import (
	"fmt"
	"net/http"

	"k8s.io/heapster/metrics/sinks/metric"
	"k8s.io/heapster/version"
)

const (
	ValidatePage = "/validate/"
)

func HandleRequest(w http.ResponseWriter, metricSink *metricsink.MetricSink) error {
	out := fmt.Sprintf("Heapster Version: %v\n\n", version.HeapsterVersion)
	/*
		    for _, source := range sources {
				out += source.DebugInfo()
			}
	*/

	out += fmt.Sprintf("Source type: Kube Node Metrics\nKubernetes Nodes plugin: \n\tHealthy Nodes:\n")
	nodes := metricSink.GetNodes()
	out += fmt.Sprintf("\n\t\t%v\n", nodes)

	out += fmt.Sprintf("External Sinks\n\tExported metrics:\n")
	//keys := metricSink.GetMetricSetKeys()
	//out += fmt.Sprintf("\t\t\t%v", keys)

	_, err := w.Write([]byte(out))
	return err
}
