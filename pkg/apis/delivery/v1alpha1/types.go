package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type PipelineList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`
	Items           []Pipeline `json:"items"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type Pipeline struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata"`
	Spec              PipelineSpec   `json:"spec"`
	Status            PipelineStatus `json:"status,omitempty"`
}

// PipelineSpec is the specification of the object
type PipelineSpec struct {
	Repo          string   `json:"repo"`
	BuildImage    string   `json:"buildImage"`
	BuildCmds     []string `json:"buildCommands"`
	TargetVersion string   `json:"targetVersion"`
	TargetName    string   `json:"targetName"`
}

// PipelineStatus represents the status in k8s
type PipelineStatus struct {
	ID      string `json:"id"`
	Success bool   `json:"success"`
}
