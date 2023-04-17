package admission

import (
	"testing"

	"github.com/bmizerany/assert"
	"k8s.io/api/admission/v1beta1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

func TestDefaultPodValues(t *testing.T) {
	reviewer := &AdmissionReviewer{
		defaultTemplates: map[string]*v1.PodTemplateSpec{
			"default": {
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						"cluster-autoscaler.kubernetes.io/safe-to-evict": "false",
					},
				},
			},
			"some-label-value": {
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						"example.com/some-annotation": "true",
					},
				},
			},
		},
	}

	pod := &v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test",
			Namespace: "test",
			Labels: map[string]string{
				"annotation-type": "some-label-value",
			},
		},
	}

	expectedPod := &v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test",
			Namespace: "test",
			Labels: map[string]string{
				"annotation-type": "some-label-value",
			},
			Annotations: map[string]string{
				"example.com/some-annotation": "true",
			},
		},
	}

	result := reviewer.defaultPodValues(pod)
	assert.Equal(t, result, expectedPod)
}

// do me a test for function getPod
func TestGetPod(t *testing.T) {
	r := &AdmissionReviewer{
		defaultTemplates: map[string]*v1.PodTemplateSpec{
			"default": {
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						"cluster-autoscaler.kubernetes.io/safe-to-evict": "false",
					},
				},
			},
			"some-label-value": {
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						"example.com/some-annotation": "true",
					},
				},
			},
		},
	}
	a := &v1beta1.AdmissionReview{
		Request: &v1beta1.AdmissionRequest{
			UID: "test",
			Object: runtime.RawExtension{
				Raw: []byte(`{
					"apiVersion": "v1",
					"kind": "Pod",
					"metadata": {
						"name": "test",
						"namespace": "test"
					}
				}`),
			},
		},
	}
	expectedPod := &v1.Pod{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "Pod",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test",
			Namespace: "test",
		},
	}
	result, _ := r.getPod(a)
	assert.Equal(t, result, expectedPod)
}
func TestDefaultedPodSimpleMerge(t *testing.T) {
	r := &AdmissionReviewer{
		defaultTemplates: map[string]*v1.PodTemplateSpec{
			"default": {
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						"foo": "bar",
					},
				},
			},
			"some-label-value": {
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						"example.com/some-annotation": "true",
					},
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
		defaultTemplates: map[string]*v1.PodTemplateSpec{
			"default": {
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						"foo": "foo",
					},
				},
			},
			"some-label-value": {
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						"example.com/some-annotation": "true",
					},
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
		defaultTemplates: map[string]*v1.PodTemplateSpec{
			"default": {
				Spec: v1.PodSpec{
					ImagePullSecrets: []v1.LocalObjectReference{{Name: "myregistry"}},
				},
			},
			"some-label-value": {
				Spec: v1.PodSpec{
					ImagePullSecrets: []v1.LocalObjectReference{{Name: "myregistry"}},
				},
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
		defaultTemplates: map[string]*v1.PodTemplateSpec{
			"default": {
				Spec: v1.PodSpec{
					ImagePullSecrets: []v1.LocalObjectReference{{Name: "myregistry"}},
				},
			},
			"some-label-value": {
				Spec: v1.PodSpec{
					ImagePullSecrets: []v1.LocalObjectReference{{Name: "myregistry"}},
				},
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
