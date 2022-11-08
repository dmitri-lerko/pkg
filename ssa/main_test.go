/*
Copyright 2021 Stefan Prodan
Copyright 2021 The Flux authors

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

package ssa

import (
	"fmt"
	"os"
	"strings"
	"sync/atomic"
	"testing"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/cli-utils/pkg/kstatus/polling"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/apiutil"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
)

var manager *ResourceManager

func TestMain(m *testing.M) {
	testEnv := &envtest.Environment{}

	cfg, err := testEnv.Start()
	if err != nil {
		panic(err)
	}

	restMapper, err := apiutil.NewDynamicRESTMapper(cfg)
	if err != nil {
		panic(err)
	}

	kubeClient, err := client.New(cfg, client.Options{
		Mapper: restMapper,
	})
	if err != nil {
		panic(err)
	}

	poller := polling.NewStatusPoller(kubeClient, restMapper, polling.Options{})

	manager = &ResourceManager{
		client: kubeClient,
		poller: poller,
		owner: Owner{
			Field: "resource-manager",
			Group: "resource-manager.io",
		},
	}

	code := m.Run()

	testEnv.Stop()

	os.Exit(code)
}

func readManifest(manifest, namespace string) ([]*unstructured.Unstructured, error) {
	data, err := os.ReadFile(manifest)
	if err != nil {
		return nil, err
	}
	yml := fmt.Sprintf(string(data), namespace)

	objects, err := ReadObjects(strings.NewReader(yml))
	if err != nil {
		return nil, err
	}

	return objects, nil
}

func setNamespace(objects []*unstructured.Unstructured, namespace string) {
	for _, object := range objects {
		object.SetNamespace(namespace)
	}

	u := &unstructured.Unstructured{}
	u.SetGroupVersionKind(schema.GroupVersionKind{
		Group:   "",
		Kind:    "Namespace",
		Version: "v1",
	})
	u.SetName(namespace)
	objects = append(objects, u)
}

var nextNameId int64

func generateName(prefix string) string {
	id := atomic.AddInt64(&nextNameId, 1)
	return fmt.Sprintf("%s-%d", prefix, id)
}

func getFirstObject(objects []*unstructured.Unstructured, kind, name string) (string, *unstructured.Unstructured) {
	for _, object := range objects {
		if object.GetKind() == kind && object.GetName() == name {
			return FmtUnstructured(object), object
		}
	}
	return "", nil
}

func removeObject(s []*unstructured.Unstructured, index int) []*unstructured.Unstructured {
	return append(s[:index], s[index+1:]...)
}
