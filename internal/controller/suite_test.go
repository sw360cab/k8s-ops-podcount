/*
Copyright 2023.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controller

import (
	"context"
	"fmt"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	podcountv1 "github.com/k8s-ops-hello/api/v1"
	//+kubebuilder:scaffold:imports
)

// These tests use Ginkgo (BDD-style Go testing framework). Refer to
// http://onsi.github.io/ginkgo/ to learn more about Ginkgo.

var cfg *rest.Config
var k8sClient client.Client
var testEnv *envtest.Environment

func TestControllers(t *testing.T) {
	RegisterFailHandler(Fail)

	RunSpecs(t, "Controller Suite")
}

var _ = BeforeSuite(func() {
	logf.SetLogger(zap.New(zap.WriteTo(GinkgoWriter), zap.UseDevMode(true)))

	By("bootstrapping test environment")
	testEnv = &envtest.Environment{
		CRDDirectoryPaths:     []string{filepath.Join("..", "..", "config", "crd", "bases")},
		ErrorIfCRDPathMissing: true,

		// The BinaryAssetsDirectory is only required if you want to run the tests directly
		// without call the makefile target test. If not informed it will look for the
		// default path defined in controller-runtime which is /usr/local/kubebuilder/.
		// Note that you must have the required binaries setup under the bin directory to perform
		// the tests directly. When we run make test it will be setup and used automatically.
		BinaryAssetsDirectory: filepath.Join("..", "..", "bin", "k8s",
			fmt.Sprintf("1.28.3-%s-%s", runtime.GOOS, runtime.GOARCH)),
	}

	var err error
	// cfg is defined in this file globally.
	cfg, err = testEnv.Start()
	Expect(err).NotTo(HaveOccurred())
	Expect(cfg).NotTo(BeNil())

	err = podcountv1.AddToScheme(scheme.Scheme)
	Expect(err).NotTo(HaveOccurred())

	//+kubebuilder:scaffold:scheme

	k8sClient, err = client.New(cfg, client.Options{Scheme: scheme.Scheme})
	Expect(err).NotTo(HaveOccurred())
	Expect(k8sClient).NotTo(BeNil())

})

var _ = AfterSuite(func() {
	By("tearing down the test environment")
	err := testEnv.Stop()
	Expect(err).NotTo(HaveOccurred())
})

var _ = Context("Creating a pod from scratch", func() {
	It("Create a pod from scratch", func() {
		By("By creating a new Pod")
		ctx := context.Background()
		pod := v1.Pod{
			TypeMeta: metav1.TypeMeta{
				APIVersion: "batch.tutorial.kubebuilder.io/v1",
				Kind:       "Pod",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:      "whoami",
				Namespace: "default",
			},
			Spec: v1.PodSpec{
				Containers: []v1.Container{
					{
						Name:            "whoami",
						Image:           "traefik/whoami",
						ImagePullPolicy: v1.PullIfNotPresent,
					},
				},
			},
		}
		Expect(k8sClient.Create(ctx, &pod)).Should(Succeed())
	})
})

var _ = Context("Creating a Pod Count", func() {
	const (
		PodCountName      = "podcountcr"
		PodCountNamespace = "default"
		JobName           = "test-job"

		timeout  = time.Second * 10
		duration = time.Second * 10
		interval = time.Millisecond * 250
	)

	It("Create a Pod Count CR", func() {
		ctx := context.Background()
		podCountCreated := podcountv1.PodCount{
			TypeMeta: metav1.TypeMeta{
				APIVersion: "github.com.k8s-ops-hello//v1",
				Kind:       "PodCount",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:      PodCountName,
				Namespace: PodCountNamespace,
				Labels: map[string]string{
					"app.kubernetes.io/name":       "podcount",
					"app.kubernetes.io/instance":   "podcount-sample",
					"app.kubernetes.io/part-of":    "k8s-hello-operator",
					"app.kubernetes.io/managed-by": "kustomize",
					"app.kubernetes.io/created-by": "k8s-hello-operator",
				},
			},
			Spec: podcountv1.PodCountSpec{
				Name:        "whoami",
				PodsPerNode: 1,
			},
		}
		Expect(k8sClient.Create(ctx, &podCountCreated)).Should(Succeed())

		podCountLookupKey := types.NamespacedName{Name: PodCountName, Namespace: PodCountNamespace}
		By("By checking that the PodCount is active")

		Eventually(func() (int, error) {
			err := k8sClient.Get(ctx, podCountLookupKey, &podCountCreated)
			if err != nil {
				return -1, err
			}

			return podCountCreated.Spec.PodsPerNode, nil
		}, timeout, interval).Should(BeNumerically("==", 1), "should value of `PodsPerNode` spec: %d", 1)
	})
})
