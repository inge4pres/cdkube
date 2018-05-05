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
		CheckInterval: 500 * time.Millisecond,
	}
}

type Handler struct {
	CheckInterval time.Duration
}

func (h *Handler) Handle(ctx types.Context, event types.Event) error {
	switch o := event.Object.(type) {
	case *v1alpha1.Pipeline:
		pipe := o
		builder := newDockerBuilderPod(o)
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
		done := make(chan bool)
		timer := time.NewTicker(h.CheckInterval)
		go finished(builder, done, h.CheckInterval)
	Building:
		for {
			select {
			case <-done:
				pipe.Status.Success = true
				break Building

			case <-timer.C:
				logrus.Infof("build still running, status of builder container %v", builder.Status.ContainerStatuses)
			}
		}

	}
	return nil
}

func newDockerBuilderPod(pipe *v1alpha1.Pipeline) *v1.Pod {

	pod := namespacedPodForPipeline(pipe, "default")
	pod.Spec.Containers = []v1.Container{
		{
			Name:    "builder",
			Image:   pipe.Spec.BuildImage,
			Command: pipe.Spec.BuildCmds,
		},
	}
	return pod
}

func namespacedPodForPipeline(pipe *v1alpha1.Pipeline, namespace string) *v1.Pod {
	labels := map[string]string{
		"app":  pipe.Spec.TargetName,
		"repo": pipe.Spec.Repo,
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
	}
}

func finished(b *v1.Pod, done chan bool, checkInterval time.Duration) {
	for {
		if b.Status.Phase == v1.PodSucceeded {
			done <- true
			return
		}
		time.Sleep(checkInterval)
	}
}
