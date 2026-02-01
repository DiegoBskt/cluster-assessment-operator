package resourcequotas

import (
	"context"
	"fmt"
	"testing"

	"github.com/openshift-assessment/cluster-assessment-operator/pkg/profiles"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type benchmarkMockClient struct {
	client.Client
	limitRanges *corev1.LimitRangeList
}

func (m *benchmarkMockClient) List(ctx context.Context, list client.ObjectList, opts ...client.ListOption) error {
	if l, ok := list.(*corev1.LimitRangeList); ok {
		*l = *m.limitRanges
		return nil
	}
	// Handle other lists minimally to avoid errors
	if nl, ok := list.(*metav1.PartialObjectMetadataList); ok {
		nl.Items = []metav1.PartialObjectMetadata{}
		return nil
	}
	if rql, ok := list.(*corev1.ResourceQuotaList); ok {
		rql.Items = []corev1.ResourceQuota{}
		return nil
	}
	return nil
}

func BenchmarkCheckLimitRanges(b *testing.B) {
	// Setup large LimitRange list
	limitRanges := &corev1.LimitRangeList{}
	for i := 0; i < 1000; i++ {
		lr := corev1.LimitRange{
			ObjectMeta: metav1.ObjectMeta{
				Name:      fmt.Sprintf("lr-%d", i),
				Namespace: fmt.Sprintf("ns-%d", i),
			},
			Spec: corev1.LimitRangeSpec{
				Limits: []corev1.LimitRangeItem{
					{
						Type: corev1.LimitTypeContainer,
						Default: corev1.ResourceList{
							corev1.ResourceMemory: resource.MustParse("4Gi"), // Triggers comparison logic but < 8Gi
						},
					},
					{
						Type: corev1.LimitTypeContainer,
						Default: corev1.ResourceList{
							corev1.ResourceMemory: resource.MustParse("16Gi"), // > 8Gi
						},
					},
				},
			},
		}
		limitRanges.Items = append(limitRanges.Items, lr)
	}

	c := &benchmarkMockClient{limitRanges: limitRanges}
	v := &ResourceQuotasValidator{}
	profile := profiles.Profile{}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = v.Validate(context.Background(), c, profile)
	}
}
