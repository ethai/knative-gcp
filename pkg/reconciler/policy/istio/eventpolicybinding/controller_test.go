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

package eventpolicybinding

import (
	"testing"

	"knative.dev/pkg/configmap"
	logtesting "knative.dev/pkg/logging/testing"
	. "knative.dev/pkg/reconciler/testing"

	// Fake injection informers
	_ "github.com/google/knative-gcp/pkg/client/injection/client/fake"
	_ "github.com/google/knative-gcp/pkg/client/injection/ducks/duck/v1alpha1/resource/fake"
	_ "github.com/google/knative-gcp/pkg/client/injection/informers/policy/v1alpha1/eventpolicy/fake"
	_ "github.com/google/knative-gcp/pkg/client/injection/informers/policy/v1alpha1/eventpolicybinding/fake"
	_ "github.com/google/knative-gcp/pkg/client/injection/informers/policy/v1alpha1/httppolicy/fake"
	_ "github.com/google/knative-gcp/pkg/client/injection/informers/policy/v1alpha1/httppolicybinding/fake"
	_ "github.com/google/knative-gcp/pkg/reconciler/testing"
)

func TestNew(t *testing.T) {
	defer logtesting.ClearAll()
	ctx, _ := SetupFakeContext(t)

	c := NewController(ctx, configmap.NewStaticWatcher())

	if c == nil {
		t.Fatal("Expected NewController to return a non-nil value")
	}
}
