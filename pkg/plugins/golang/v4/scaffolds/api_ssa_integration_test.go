//go:build integration

/*
Copyright 2026 The Kubernetes Authors.

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

package scaffolds

import (
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	pluginutil "sigs.k8s.io/kubebuilder/v4/pkg/plugin/util"
	"sigs.k8s.io/kubebuilder/v4/test/e2e/utils"
)

var _ = Describe("Server-Side Apply (--ssa) Scaffolding", func() {
	var kbc *utils.TestContext

	BeforeEach(func() {
		var err error
		kbc, err = utils.NewTestContext(pluginutil.KubebuilderBinName, "GO111MODULE=on")
		Expect(err).NotTo(HaveOccurred())
		Expect(kbc.Prepare()).To(Succeed())

		By("initializing a project")
		Expect(kbc.Init(
			"--domain", "test.io",
			"--repo", "test.io/ssatest",
		)).To(Succeed())
	})

	AfterEach(func() {
		By("removing working directory")
		kbc.Destroy()
	})

	It("should scaffold a webhook for an SSA resource and run 'make manifests' successfully", func() {
		By("creating an API with Server-Side Apply enabled")
		Expect(kbc.CreateAPI(
			"--group", "crew",
			"--version", "v1",
			"--kind", "Captain",
			"--resource", "--controller",
			"--ssa",
			"--make=false",
		)).To(Succeed())

		By("verifying the SSA applyconfiguration generator was wired into the Makefile")
		makefile, err := os.ReadFile(filepath.Join(kbc.Dir, "Makefile"))
		Expect(err).NotTo(HaveOccurred())
		Expect(string(makefile)).To(ContainSubstring("applyconfiguration:headerFile"))

		By("creating defaulting and validation webhooks for the SSA resource")
		Expect(kbc.CreateWebhook(
			"--group", "crew",
			"--version", "v1",
			"--kind", "Captain",
			"--defaulting",
			"--programmatic-validation",
			"--make=false",
		)).To(Succeed())

		By("running 'make manifests'")
		Expect(kbc.Make("manifests")).To(Succeed())

		By("verifying webhook manifests were generated for the SSA resource")
		webhookManifest, err := os.ReadFile(filepath.Join(kbc.Dir, "config", "webhook", "manifests.yaml"))
		Expect(err).NotTo(HaveOccurred())
		Expect(string(webhookManifest)).To(ContainSubstring("captains"))

		By("verifying applyconfiguration code was generated for the SSA resource")
		acFile := filepath.Join(kbc.Dir, "api", "v1", "applyconfiguration", "api", "v1", "captain.go")
		_, err = os.Stat(acFile)
		Expect(err).NotTo(HaveOccurred(), "applyconfiguration should be generated for the SSA resource")
	})

	It("should support splitting the SSA API and its controller across two commands", func() {
		typesFile := filepath.Join(kbc.Dir, "api", "v1", "admiral_types.go")
		controllerFile := filepath.Join(kbc.Dir, "internal", "controller", "admiral_controller.go")

		By("creating the SSA API resource WITHOUT a controller")
		Expect(kbc.CreateAPI(
			"--group", "crew",
			"--version", "v1",
			"--kind", "Admiral",
			"--resource=true",
			"--controller=false",
			"--ssa",
			"--make=false",
		)).To(Succeed())

		By("verifying the SSA marker was scaffolded on the resource")
		typesContent, err := os.ReadFile(typesFile)
		Expect(err).NotTo(HaveOccurred())
		Expect(string(typesContent)).To(ContainSubstring("+genclient"))

		By("verifying no controller was scaffolded yet")
		_, err = os.Stat(controllerFile)
		Expect(os.IsNotExist(err)).To(BeTrue(), "controller should not exist before it is requested")

		By("adding the controller for the existing SSA resource WITHOUT recreating the API")
		Expect(kbc.CreateAPI(
			"--group", "crew",
			"--version", "v1",
			"--kind", "Admiral",
			"--resource=false",
			"--controller=true",
			"--make=false",
		)).To(Succeed())

		By("verifying the controller was scaffolded and wired to the SSA resource")
		controllerContent, err := os.ReadFile(controllerFile)
		Expect(err).NotTo(HaveOccurred())
		Expect(string(controllerContent)).To(ContainSubstring("type AdmiralReconciler struct"))
		Expect(string(controllerContent)).To(ContainSubstring("crewv1.Admiral"))

		By("verifying the controller is registered with the manager in main.go")
		mainContent, err := os.ReadFile(filepath.Join(kbc.Dir, "cmd", "main.go"))
		Expect(err).NotTo(HaveOccurred())
		Expect(string(mainContent)).To(ContainSubstring("AdmiralReconciler{"))

		By("verifying the SSA marker on the resource was preserved after adding the controller")
		typesContent, err = os.ReadFile(typesFile)
		Expect(err).NotTo(HaveOccurred())
		Expect(string(typesContent)).To(ContainSubstring("+genclient"))

		By("verifying the PROJECT file tracks both ssa and the controller for the resource")
		projectContent, err := os.ReadFile(filepath.Join(kbc.Dir, "PROJECT"))
		Expect(err).NotTo(HaveOccurred())
		Expect(string(projectContent)).To(ContainSubstring("ssa: true"),
			"adding the controller must not drop the ssa flag from the PROJECT file")
		Expect(string(projectContent)).To(ContainSubstring("controller: true"))
	})
})
