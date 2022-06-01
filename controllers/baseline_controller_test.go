package controllers

import (
	"context"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/envtest/komega"

	perfv1 "github.com/josecastillolema/baseline-operator/api/v1"
)

var _ = Describe("Baseline controller", func() {

	// Define utility constants for object names and testing timeouts/durations and intervals.
	const (
		BaselineName      = "test-baseline"
		BaselineNamespace = "default"
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
					Cpu:    new(int32),
					Memory: "1G",
					Io:     1,
					Sock:   1,
					Custom: "--timer 1",
					Image:  "quay.io/jcastillolema/stressng:0.14.01",
				},
			}
			Expect(k8sClient.Create(ctx, baseline)).Should(Succeed())
			Eventually(komega.Object(baseline)).Should(HaveField("Status.Command", Equal("stress-ng -t 0 --cpu 0 --vm 1 --vm-bytes 1G --io 1 --sock 1 --sock-if eth0 --timer 1")))
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

			*createdBaseline.Spec.Cpu = 2
			createdBaseline.Spec.Memory = "2G"
			createdBaseline.Spec.Custom = "--timer 2"
			Expect(k8sClient.Update(ctx, createdBaseline)).Should(Succeed())
			Eventually(komega.Object(createdBaseline)).Should(HaveField("Status.Command", Equal("stress-ng -t 0 --cpu 2 --vm 1 --vm-bytes 2G --io 1 --sock 1 --sock-if eth0 --timer 2")))
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
			Eventually(komega.Object(createdBaseline)).Should(HaveField("Status.Command", Equal("stress-ng -t 0 --cpu 2 --vm 1 --vm-bytes 2G --io 1 --sock 1 --sock-if eth0")))
		})
	})

	Context("Removing a Baseline CRD CPU field", func() {
		It("Should accordingly update the status field", func() {
			By("By removing the existing Baseline CPU field")
			baselineLookupKey := types.NamespacedName{Name: BaselineName, Namespace: BaselineNamespace}
			createdBaseline := &perfv1.Baseline{}
			err := k8sClient.Get(ctx, baselineLookupKey, createdBaseline)
			if err != nil {
				panic("Baseline object should exist from previous test")
			}

			createdBaseline.Spec.Cpu = nil
			Expect(k8sClient.Update(ctx, createdBaseline)).Should(Succeed())
			Eventually(komega.Object(createdBaseline)).Should(HaveField("Status.Command", Equal("stress-ng -t 0 --vm 1 --vm-bytes 2G --io 1 --sock 1 --sock-if eth0")))
		})
	})

	Context("Removing a Baseline CRD IO field", func() {
		It("Should accordingly update the status field", func() {
			By("By removing the existing Baseline IO field")
			baselineLookupKey := types.NamespacedName{Name: BaselineName, Namespace: BaselineNamespace}
			createdBaseline := &perfv1.Baseline{}
			err := k8sClient.Get(ctx, baselineLookupKey, createdBaseline)
			if err != nil {
				panic("Baseline object should exist from previous test")
			}

			createdBaseline.Spec.Io = 0
			Expect(k8sClient.Update(ctx, createdBaseline)).Should(Succeed())
			Eventually(komega.Object(createdBaseline)).Should(HaveField("Status.Command", Equal("stress-ng -t 0 --vm 1 --vm-bytes 2G --sock 1 --sock-if eth0")))
		})
	})

	Context("Removing a Baseline CRD Mem field", func() {
		It("Should accordingly update the status field", func() {
			By("By removing the existing Baseline Mem field")
			baselineLookupKey := types.NamespacedName{Name: BaselineName, Namespace: BaselineNamespace}
			createdBaseline := &perfv1.Baseline{}
			err := k8sClient.Get(ctx, baselineLookupKey, createdBaseline)
			if err != nil {
				panic("Baseline object should exist from previous test")
			}

			createdBaseline.Spec.Memory = ""
			Expect(k8sClient.Update(ctx, createdBaseline)).Should(Succeed())
			Eventually(komega.Object(createdBaseline)).Should(HaveField("Status.Command", Equal("stress-ng -t 0 --sock 1 --sock-if eth0")))
		})
	})

	Context("Adding a Baseline CRD CPU field", func() {
		It("Should accordingly update the status field", func() {
			By("By adding the existing Baseline CPU field")
			baselineLookupKey := types.NamespacedName{Name: BaselineName, Namespace: BaselineNamespace}
			createdBaseline := &perfv1.Baseline{}
			err := k8sClient.Get(ctx, baselineLookupKey, createdBaseline)
			if err != nil {
				panic("Baseline object should exist from previous test")
			}

			createdBaseline.Spec.Io = 1
			Expect(k8sClient.Update(ctx, createdBaseline)).Should(Succeed())
			Eventually(komega.Object(createdBaseline)).Should(HaveField("Status.Command", Equal("stress-ng -t 0 --io 1 --sock 1 --sock-if eth0")))
		})
	})

	Context("Removing a Baseline CRD Sock field", func() {
		It("Should accordingly update the status field", func() {
			By("By removing the existing Baseline Sock field")
			baselineLookupKey := types.NamespacedName{Name: BaselineName, Namespace: BaselineNamespace}
			createdBaseline := &perfv1.Baseline{}
			err := k8sClient.Get(ctx, baselineLookupKey, createdBaseline)
			if err != nil {
				panic("Baseline object should exist from previous test")
			}

			createdBaseline.Spec.Sock = 0
			Expect(k8sClient.Update(ctx, createdBaseline)).Should(Succeed())
			Eventually(komega.Object(createdBaseline)).Should(HaveField("Status.Command", Equal("stress-ng -t 0 --io 1")))
		})
	})
})
