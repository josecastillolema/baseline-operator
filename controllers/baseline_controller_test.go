package controllers

import (
	"context"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	// "sigs.k8s.io/controller-runtime/pkg/envtest/komega"

	perfv1 "github.com/josecastillolema/baseline-operator/api/v1"
)

var _ = Describe("Baseline controller", func() {

	// Define utility constants for object names and testing timeouts/durations and intervals.
	const (
		BaselineName      = "test-baseline"
		BaselineNamespace = "default"

		timeout  = time.Second * 10
		duration = time.Second * 10
		interval = time.Millisecond * 250
	)

	Context("When updating CronJob Status", func() {
		It("Should increase CronJob Status.Active count when new Jobs are created", func() {
			By("By creating a new Baseline")
			ctx := context.Background()
			baseline := &perfv1.Baseline{
				TypeMeta: metav1.TypeMeta{
					APIVersion: "perf.baseline.io/v1",
					Kind:       "Baseline",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      BaselineName,
					Namespace: BaselineNamespace,
				},
				Spec: perfv1.BaselineSpec{
					Cpu:   1,
					Io:    1,
					Sock:  1,
					Image: "quay.io/jcastillolema/stressng:0.14.01",
				},
			}
			Expect(k8sClient.Create(ctx, baseline)).Should(Succeed())

			baselineLookupKey := types.NamespacedName{Name: BaselineName, Namespace: BaselineNamespace}
			createdBaseline := &perfv1.Baseline{}

			// We'll need to retry getting this newly created Baseline, given that creation may not immediately happen.
			Eventually(func() (bool, error) {
				err := k8sClient.Get(ctx, baselineLookupKey, createdBaseline)
				if err != nil {
					return false, err
				}
				return createdBaseline.Status.Command != "", nil
			}, timeout, interval).Should(BeTrue())

			//Eventually(komega.Object(baseline)).Should(HaveField("Status.Command", Not(BeEmpty())))
			// Let's make sure our command string value was properly converted/handled.
			Expect(createdBaseline.Status.Command).Should(Equal("stress-ng -t 0 --cpu 1 --io 1 --sock 1 --sock-if eth0"))
		})
	})

})
