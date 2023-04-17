package admission

import (
	"encoding/json"
	"errors"

	"github.com/imdario/mergo"
	"github.com/softonic/pod-defaulter/pkg/log"
	"k8s.io/api/admission/v1beta1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/equality"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog"
	"knative.dev/pkg/apis/duck"
)

type AdmissionReviewer struct {
	defaultTemplates map[string]*v1.PodTemplateSpec
}

func NewPodDefaultValuesAdmissionReviewer(templates map[string]*v1.PodTemplateSpec) *AdmissionReviewer {
	return &AdmissionReviewer{
		defaultTemplates: templates,
	}
}

// PerformAdmissionReview : It generates the Adminission Review Response
func (r *AdmissionReviewer) PerformAdmissionReview(admissionReview *v1beta1.AdmissionReview) {
	pod, err := r.getPod(admissionReview)
	if err != nil {
		admissionReview.Response = r.newAdmissionError(pod, err)
		return
	}

	defaultedPod := r.defaultPodValues(pod)

	// if equals, don't patch
	if equality.Semantic.DeepEqual(pod, defaultedPod) {
		admissionReview.Response = r.admissionAllowedResponse(pod)
		return
	}

	// If we encountered changes, then synthesize and apply
	// a patch.
	patchBytes, err := duck.CreateBytePatch(pod, defaultedPod)

	if err != nil {
		admissionReview.Response = r.newAdmissionError(pod, err)
		return
	}

	klog.V(log.INFO).Infof("Patching pod %s/%s", pod.Namespace, pod.Name)
	patchType := v1beta1.PatchTypeJSONPatch

	admissionReview.Response = &v1beta1.AdmissionResponse{
		Result: &metav1.Status{
			Status: "Success",
		},
		Patch:     patchBytes,
		PatchType: &patchType,
		Allowed:   true,
		UID:       admissionReview.Request.UID,
	}
}

func (r *AdmissionReviewer) newAdmissionError(pod *v1.Pod, err error) *v1beta1.AdmissionResponse {
	if pod != nil {
		klog.Errorf("Pod %s/%s failed admission review: %v", pod.Namespace, pod.Name, err)
	} else {
		klog.Errorf("Failed admission review: %v", err)
	}
	return &v1beta1.AdmissionResponse{
		Result: &metav1.Status{
			Message: err.Error(),
			Status:  "Fail",
		},
	}
}

func (r *AdmissionReviewer) admissionAllowedResponse(pod *v1.Pod) *v1beta1.AdmissionResponse {
	klog.Errorf("Skipping admission review for pod %s/%s", pod.Namespace, pod.Name)
	return &v1beta1.AdmissionResponse{
		Allowed: true,
	}
}

// receives the admissionReview and returns the pod inside the admissionReview
func (r *AdmissionReviewer) getPod(admissionReview *v1beta1.AdmissionReview) (*v1.Pod, error) {
	var pod v1.Pod
	if admissionReview.Request == nil {
		return nil, errors.New("Request is nil")
	}
	if admissionReview.Request.Object.Raw == nil {
		return nil, errors.New("Request object raw is nil")
	}
	err := json.Unmarshal(admissionReview.Request.Object.Raw, &pod)
	if err != nil {
		return nil, err
	}
	return &pod, nil
}

func (r *AdmissionReviewer) defaultPodValues(pod *v1.Pod) *v1.Pod {
	result := &v1.Pod{}
	mergo.Merge(result, pod)

	annotationType, ok := pod.Labels["annotation-type"]
	if !ok {
		annotationType = "default"
	}

	template, ok := r.defaultTemplates[annotationType]
	if ok {
		mergo.Merge(&result.Spec, template.Spec)
		mergo.Merge(&result.ObjectMeta, template.ObjectMeta)
	}

	return result
}
