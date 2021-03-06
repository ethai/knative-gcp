/*
Copyright 2020 Google LLC

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package lib

import (
	"sync"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	eventingtestlib "knative.dev/eventing/test/lib"
)

const (
	pubSubSecretName      = "google-cloud-key"
	pubSubSecretNamespace = "default"
)

var setTracingConfigOnce = sync.Once{}

// DuplicatePubSubSecret duplicates the PubSub secret to the test namespace.
func DuplicatePubSubSecret(client *eventingtestlib.Client) {
	secret, err := client.Kube.Kube.CoreV1().Secrets(pubSubSecretNamespace).Get(pubSubSecretName, metav1.GetOptions{})
	if err != nil {
		client.T.Fatalf("could not get secret: %v", err)
	}

	if _, err = client.Kube.Kube.CoreV1().Secrets(client.Namespace).Create(&corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:        secret.Name,
			Labels:      secret.Labels,
			Annotations: secret.Annotations,
		},
		Type:       secret.Type,
		Data:       secret.Data,
		StringData: secret.StringData,
	}); err != nil {
		client.T.Fatalf("could not create secret: %v", err)
	}
}

func SetTracingToZipkin(client *eventingtestlib.Client) {
	setTracingConfigOnce.Do(func() {
		err := client.Kube.UpdateConfigMap("cloud-run-events", "config-tracing", map[string]string{
			"backend":         "zipkin",
			"zipkin-endpoint": "http://zipkin.istio-system.svc.cluster.local:9411/api/v2/spans",
		})
		if err != nil {
			client.T.Fatalf("Unable to set the ConfigMap: %v", err)
		}
		// Wait for 5 seconds to let the ConfigMap be synced up.
		time.Sleep(5 * time.Second)
	})
}
