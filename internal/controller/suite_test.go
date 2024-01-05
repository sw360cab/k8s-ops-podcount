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

	appsv1 "k8s.io/api/apps/v1"
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

const (
	PodCountName     = "podcountcr"
	DefaultNamespace = "default"
	// PodCountNamespace = "default"

	ApiVersion = "github.com.k8s-ops-hello/v1"
	JobName    = "test-job"

	timeout  = time.Second * 10
	duration = time.Second * 10
	interval = time.Millisecond * 250
)

var cfg *rest.Config
var k8sClient client.Client
var testEnv *envtest.Environment

func TestControllers(t *testing.T) {
	RegisterFailHandler(Fail)

	RunSpecs(t, "Controller Suite")
}

func createPodCountCr(maxPodPerNode int, podName string, podNs string) podcountv1.PodCount {
	return podcountv1.PodCount{
		TypeMeta: metav1.TypeMeta{
			APIVersion: ApiVersion,
			Kind:       "PodCount",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      podName,
			Namespace: podNs,
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
			PodsPerNode: maxPodPerNode,
		},
	}
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
				APIVersion: ApiVersion,
				Kind:       "Pod",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:      "whoami",
				Namespace: DefaultNamespace,
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

	It("Create a Pod Count CR", func() {
		ctx := context.Background()
		podCountCreated := createPodCountCr(2, PodCountName, DefaultNamespace)
		Expect(k8sClient.Create(ctx, &podCountCreated)).Should(Succeed())

		podCountLookupKey := types.NamespacedName{Name: PodCountName, Namespace: DefaultNamespace}
		By("By checking that the PodCount is active")

		var podCountFound podcountv1.PodCount
		Eventually(func() (int, error) {
			err := k8sClient.Get(ctx, podCountLookupKey, &podCountFound)
			if err != nil {
				return -1, err
			}

			return podCountFound.Spec.PodsPerNode, nil
		}, timeout, interval).Should(BeNumerically("==", 2), "should value of `PodsPerNode` spec: %d", 2)
	})

	It("Create a deployment with multiple pods", func() {
		var replicas int32 = 2
		var deploymentFound appsv1.Deployment

		By("By creating a new Deployment")
		ctx := context.Background()
		deployment := appsv1.Deployment{
			// TypeMeta: metav1.TypeMeta{
			// 	APIVersion: apiVersion,
			// 	Kind:       "Deployment",
			// },
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-deployment",
				Namespace: DefaultNamespace,
			},
			Spec: appsv1.DeploymentSpec{
				Replicas: &replicas,
				Selector: &metav1.LabelSelector{
					MatchLabels: map[string]string{
						"app": "whoami",
					},
				},
				Template: v1.PodTemplateSpec{
					ObjectMeta: metav1.ObjectMeta{
						Labels: map[string]string{
							"app": "whoami",
						},
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
				},
			},
		}
		Expect(k8sClient.Create(ctx, &deployment)).Should(Succeed())
		deploymentLookupKey := types.NamespacedName{Name: "test-deployment", Namespace: DefaultNamespace}
		Eventually(func() (int32, error) {
			err := k8sClient.Get(ctx, deploymentLookupKey, &deploymentFound)
			if err != nil {
				return -1, err
			}

			return *deploymentFound.Spec.Replicas, nil
		}, timeout, interval).Should(BeNumerically("==", replicas), "should value of Deployment `replicas` spec: %d", replicas)
	})

})
