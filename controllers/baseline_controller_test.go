package controllers

import (
	"context"
	"strings"
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

	Context("Creating a Baseline CRD", func() {
		It("Should accordingly create the status field", func() {
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
					Cpu:    1,
					Memory: "1G",
					Io:     1,
					Sock:   1,
					Custom: "--timer 1",
					Image:  "quay.io/jcastillolema/stressng:0.14.01",
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
			Expect(createdBaseline.Status.Command).Should(Equal("stress-ng -t 0 --cpu 1 --vm 1 --vm-bytes 1G --io 1 --sock 1 --sock-if eth0 --timer 1"))
		})
	})

	Context("Updating some Baseline CRD fields (CPU(int), Mem(string), custom)", func() {
		It("Should accordingly update the status field", func() {
			By("By updating the existing Baseline fields")
			baselineLookupKey := types.NamespacedName{Name: BaselineName, Namespace: BaselineNamespace}
			createdBaseline := &perfv1.Baseline{}
			err := k8sClient.Get(ctx, baselineLookupKey, createdBaseline)
			if err != nil {
				panic("Baseline object should exist from previous test")
			}

			createdBaseline.Spec.Cpu = 2
			createdBaseline.Spec.Memory = "2G"
			createdBaseline.Spec.Custom = "--timer 2"
			Expect(k8sClient.Update(ctx, createdBaseline)).Should(Succeed())

			// We'll need to retry getting this newly updated Baseline, given that update may not immediately happen.
			Eventually(func() (bool, error) {
				err := k8sClient.Get(ctx, baselineLookupKey, createdBaseline)
				if err != nil {
					return false, err
				}
				return strings.Contains(createdBaseline.Status.Command, "--cpu 2 --vm 1 --vm-bytes 2G"), nil
			}, timeout, interval).Should(BeTrue())

			// Let's make sure our command string value was properly converted/handled.
			Expect(createdBaseline.Status.Command).Should(Equal("stress-ng -t 0 --cpu 2 --vm 1 --vm-bytes 2G --io 1 --sock 1 --sock-if eth0 --timer 2"))
		})
	})

	Context("Removing a Baseline CRD custom field", func() {
		It("Should accordingly update the status field", func() {
			By("By removing the existing Baseline custom field")
			baselineLookupKey := types.NamespacedName{Name: BaselineName, Namespace: BaselineNamespace}
			createdBaseline := &perfv1.Baseline{}
			err := k8sClient.Get(ctx, baselineLookupKey, createdBaseline)
			if err != nil {
				panic("Baseline object should exist from previous test")
			}

			createdBaseline.Spec.Custom = ""
			Expect(k8sClient.Update(ctx, createdBaseline)).Should(Succeed())

			// We'll need to retry getting this newly updated Baseline, given that update may not immediately happen.
			Eventually(func() (bool, error) {
				err := k8sClient.Get(ctx, baselineLookupKey, createdBaseline)
				if err != nil {
					return false, err
				}
				return !strings.Contains(createdBaseline.Status.Command, "--timer"), nil
			}, timeout, interval).Should(BeTrue())

			// Let's make sure our command string value was properly converted/handled.
			Expect(createdBaseline.Status.Command).Should(Equal("stress-ng -t 0 --cpu 2 --vm 1 --vm-bytes 2G --io 1 --sock 1 --sock-if eth0"))
		})
	})

	/* Context("Removing a Baseline CRD CPU field", func() {
		It("Should accordingly update the status field", func() {
			By("By setting to zero the existing Baseline CPU field")
			baselineLookupKey := types.NamespacedName{Name: BaselineName, Namespace: BaselineNamespace}
			createdBaseline := &perfv1.Baseline{}
			err := k8sClient.Get(ctx, baselineLookupKey, createdBaseline)
			if err != nil {
				panic("Baseline object should exist from previous test")
			}

			createdBaseline.Spec.Cpu = 0
			Expect(k8sClient.Update(ctx, createdBaseline)).Should(Succeed())

			// We'll need to retry getting this newly updated Baseline, given that update may not immediately happen.
			Eventually(func() (bool, error) {
				err := k8sClient.Get(ctx, baselineLookupKey, createdBaseline)
				if err != nil {
					return false, err
				}
				return !strings.Contains(createdBaseline.Status.Command, "--cpu"), nil
			}, timeout, interval).Should(BeTrue())

			// Let's make sure our command string value was properly converted/handled.
			Expect(createdBaseline.Status.Command).Should(Equal("stress-ng -t 0 --io 1 --sock 1 --sock-if eth0"))
		})
	}) */

})
