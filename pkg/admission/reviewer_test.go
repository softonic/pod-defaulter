package admission

import (
	"github.com/bmizerany/assert"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"testing"
)

func TestDefaultedPodSimpleMerge(t *testing.T) {
	r := &AdmissionReviewer{
		defaultTemplate: &v1.PodTemplateSpec{
			ObjectMeta: metav1.ObjectMeta{
				Annotations: map[string]string{
					"foo": "bar",
				},
			},
		},
	}
	pod := &v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Annotations: map[string]string{
				"bar": "foo",
			},
		},
	}
	expectedPod := &v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Annotations: map[string]string{
				"bar": "foo",
				"foo": "bar",
			},
		},
	}
	result := r.defaultPodValues(pod)
	assert.Equal(t, result, expectedPod)
}

func TestDefaultedPodWithValue(t *testing.T) {
	r := &AdmissionReviewer{
		defaultTemplate: &v1.PodTemplateSpec{
			ObjectMeta: metav1.ObjectMeta{
				Annotations: map[string]string{
					"foo": "bar",
				},
			},
		},
	}
	pod := &v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Annotations: map[string]string{
				"foo": "foo",
			},
		},
	}
	expectedPod := &v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Annotations: map[string]string{
				"foo": "foo",
			},
		},
	}
	result := r.defaultPodValues(pod)
	assert.Equal(t, result, expectedPod)
}

func TestDefaultedPodAddSpec(t *testing.T) {
	r := &AdmissionReviewer{
		defaultTemplate: &v1.PodTemplateSpec{
			Spec: v1.PodSpec{
				ImagePullSecrets: []v1.LocalObjectReference{{Name: "myregistry"}},
			},
		},
	}
	pod := &v1.Pod{
		ObjectMeta: metav1.ObjectMeta{},
	}
	expectedPod := &v1.Pod{
		Spec: v1.PodSpec{
			ImagePullSecrets: []v1.LocalObjectReference{{Name: "myregistry"}},
		},
	}
	result := r.defaultPodValues(pod)
	assert.Equal(t, result, expectedPod)
}

func TestDefaultedPodMergeSpec(t *testing.T) {
	r := &AdmissionReviewer{
		defaultTemplate: &v1.PodTemplateSpec{
			Spec: v1.PodSpec{
				ImagePullSecrets: []v1.LocalObjectReference{{Name: "myregistry"}},
			},
		},
	}
	pod := &v1.Pod{
		ObjectMeta: metav1.ObjectMeta{},
		Spec: v1.PodSpec{
			ImagePullSecrets: []v1.LocalObjectReference{{Name: "defaultregistry"}},
		},
	}
	expectedPod := &v1.Pod{
		Spec: v1.PodSpec{
			ImagePullSecrets: []v1.LocalObjectReference{{Name: "defaultregistry"}},
		},
	}
	result := r.defaultPodValues(pod)
	assert.Equal(t, result, expectedPod)
}
