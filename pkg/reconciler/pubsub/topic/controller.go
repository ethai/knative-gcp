/*
Copyright 2019 Google LLC

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

package topic

import (
	"context"

	"github.com/kelseyhightower/envconfig"
	"go.uber.org/zap"
	"k8s.io/client-go/tools/cache"
	"knative.dev/pkg/configmap"
	"knative.dev/pkg/controller"
	"knative.dev/pkg/logging"
	tracingconfig "knative.dev/pkg/tracing/config"

	"github.com/google/knative-gcp/pkg/apis/pubsub/v1alpha1"
	gpubsub "github.com/google/knative-gcp/pkg/gclient/pubsub"
	"github.com/google/knative-gcp/pkg/reconciler"
	"github.com/google/knative-gcp/pkg/reconciler/identity"
	"github.com/google/knative-gcp/pkg/reconciler/pubsub"

	topicinformer "github.com/google/knative-gcp/pkg/client/injection/informers/pubsub/v1alpha1/topic"
	topicreconciler "github.com/google/knative-gcp/pkg/client/injection/reconciler/pubsub/v1alpha1/topic"
	serviceaccountinformers "knative.dev/pkg/client/injection/kube/informers/core/v1/serviceaccount"
	serviceinformer "knative.dev/serving/pkg/client/injection/informers/serving/v1/service"
)

const (
	// reconcilerName is the name of the reconciler
	reconcilerName = "Topics"

	// controllerAgentName is the string used by this controller to identify
	// itself when creating events.
	controllerAgentName = "cloud-run-events-pubsub-topic-controller"
)

type envConfig struct {
	// Publisher is the image used to publish to Pub/Sub. Required.
	Publisher string `envconfig:"PUBSUB_PUBLISHER_IMAGE" required:"true"`
}

// NewController initializes the controller and is called by the generated code
// Registers event handlers to enqueue events
func NewController(
	ctx context.Context,
	cmw configmap.Watcher,
) *controller.Impl {
	topicInformer := topicinformer.Get(ctx)
	serviceInformer := serviceinformer.Get(ctx)
	serviceAccountInformer := serviceaccountinformers.Get(ctx)

	logger := logging.FromContext(ctx).Named(controllerAgentName).Desugar()

	var env envConfig
	if err := envconfig.Process("", &env); err != nil {
		logger.Fatal("Failed to process env var", zap.Error(err))
	}

	pubsubBase := &pubsub.PubSubBase{
		Base: reconciler.NewBase(ctx, controllerAgentName, cmw),
	}

	r := &Reconciler{
		PubSubBase:     pubsubBase,
		Identity:       identity.NewIdentity(ctx),
		topicLister:    topicInformer.Lister(),
		serviceLister:  serviceInformer.Lister(),
		publisherImage: env.Publisher,
		createClientFn: gpubsub.NewClient,
	}

	impl := topicreconciler.NewImpl(ctx, r)

	pubsubBase.Logger.Info("Setting up event handlers")
	topicInformer.Informer().AddEventHandlerWithResyncPeriod(controller.HandleAll(impl.Enqueue), reconciler.DefaultResyncPeriod)

	serviceInformer.Informer().AddEventHandler(cache.FilteringResourceEventHandler{
		FilterFunc: controller.Filter(v1alpha1.SchemeGroupVersion.WithKind("Topic")),
		Handler:    controller.HandleAll(impl.EnqueueControllerOf),
	})

	serviceAccountInformer.Informer().AddEventHandler(cache.FilteringResourceEventHandler{
		FilterFunc: controller.FilterGroupVersionKind(v1alpha1.SchemeGroupVersion.WithKind("Topic")),
		Handler:    controller.HandleAll(impl.EnqueueControllerOf),
	})

	cmw.Watch(tracingconfig.ConfigName, r.UpdateFromTracingConfigMap)

	return impl
}
