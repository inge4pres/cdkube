package stub

import (
	"time"

	"github.com/google/uuid"
	"github.com/inge4pres/cdkube/pkg/apis/delivery/v1alpha1"

	"github.com/operator-framework/operator-sdk/pkg/sdk/action"
	"github.com/operator-framework/operator-sdk/pkg/sdk/handler"
	"github.com/operator-framework/operator-sdk/pkg/sdk/query"
	"github.com/operator-framework/operator-sdk/pkg/sdk/types"
	"github.com/sirupsen/logrus"
	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

func NewHandler() handler.Handler {
	return &Handler{
		CheckInterval: 1000 * time.Millisecond,
	}
}

type Handler struct {
	CheckInterval time.Duration
}

func (h *Handler) Handle(ctx types.Context, event types.Event) error {
	switch o := event.Object.(type) {
	case *v1alpha1.Pipeline:
		pipe := o
		builder := newBuilderPod(o, "default")
		err := action.Create(builder)
		if err != nil && !errors.IsAlreadyExists(err) {
			logrus.Errorf("could not create pod for pipeline: %v", err)
			return err
		}
		// update builder with an ID
		uid, err := uuid.NewUUID()
		if err != nil {
			logrus.Errorf("error generating build ID: %v", err)
			return err
		}

		pipe.Status.ID = uid.String()
		if err := query.Get(builder); err != nil {
			logrus.Errorf("could not read the state of builder pod: %v", err)
			return err
		}

		logrus.Infof("current pipeline request: %v", pipe.Spec)
		logrus.Infof("current builder pod info: %v", builder.Spec.Containers[0])

		done := make(chan bool)
		timer := time.NewTicker(h.CheckInterval)
	Building:
		for {
			select {
			case <-done:
				break Building

			case <-timer.C:
				logrus.Infof("build still running, status of builder container: %s", builder.Status.Phase)
				podSucceded(builder, done)
			}
		}
		pipe.Status.Success = true
		logrus.Info("pipeline completed successfully")

	}
	return nil
}

func newBuilderPod(pipe *v1alpha1.Pipeline, namespace string) *v1.Pod {
	// spec := builderSpec(pipe.Spec.BuildImage, pipe.Spec.BuildCmds, pipe.Spec.BuildArgs)
	// return namespacedPodForPipeline(pipe, namespace, *spec)
	labels := map[string]string{
		"app": pipe.Spec.TargetName,
	}
	return &v1.Pod{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Pod",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      pipe.Spec.TargetName,
			Namespace: namespace,
			OwnerReferences: []metav1.OwnerReference{
				*metav1.NewControllerRef(pipe, schema.GroupVersionKind{
					Group:   v1alpha1.SchemeGroupVersion.Group,
					Version: v1alpha1.SchemeGroupVersion.Version,
					Kind:    "Pipeline",
				}),
			},
			Labels: labels,
		},
		Spec: v1.PodSpec{
			Containers: []v1.Container{
				{
					Name:    "builder",
					Image:   pipe.Spec.BuildImage,
					Command: pipe.Spec.BuildCmds,
					Args:    pipe.Spec.BuildArgs,
				},
			},
		},
	}
}

func builderSpec(image string, commands, args []string) *v1.PodSpec {
	return &v1.PodSpec{
		Containers: []v1.Container{
			{
				Name:    "builder",
				Image:   image,
				Command: commands,
				Args:    args,
			},
		},
	}
}

func namespacedPodForPipeline(pipe *v1alpha1.Pipeline, namespace string, containersSpec v1.PodSpec) *v1.Pod {
	labels := map[string]string{
		"app": pipe.Spec.TargetName,
	}
	return &v1.Pod{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Pod",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      pipe.Spec.TargetName,
			Namespace: namespace,
			OwnerReferences: []metav1.OwnerReference{
				*metav1.NewControllerRef(pipe, schema.GroupVersionKind{
					Group:   v1alpha1.SchemeGroupVersion.Group,
					Version: v1alpha1.SchemeGroupVersion.Version,
					Kind:    "Pipeline",
				}),
			},
			Labels: labels,
		},
		Spec: containersSpec,
	}
}

func podSucceded(pod *v1.Pod, done chan bool) {
	if pod.Status.Phase == v1.PodSucceeded {
		done <- true
	}
}
