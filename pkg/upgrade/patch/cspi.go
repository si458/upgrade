/*
Copyright 2020 The OpenEBS Authors

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

package patch

import (
	"context"
	"strings"

	apis "github.com/openebs/api/v3/pkg/apis/cstor/v1"
	clientset "github.com/openebs/api/v3/pkg/client/clientset/versioned"
	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/klog"
)

// CSPI ...
type CSPI struct {
	Object *apis.CStorPoolInstance
	Data   []byte
	Client clientset.Interface
}

// CSPIOptions ...
type CSPIOptions func(*CSPI)

// NewCSPI ...
func NewCSPI(opts ...CSPIOptions) *CSPI {
	obj := &CSPI{}
	for _, o := range opts {
		o(obj)
	}
	return obj
}

// WithCSPIClient ...
func WithCSPIClient(c clientset.Interface) CSPIOptions {
	return func(obj *CSPI) {
		obj.Client = c
	}
}

// PreChecks ...
func (c *CSPI) PreChecks(from, to string) error {
	if c.Object == nil {
		return errors.Errorf("nil cspi object")
	}
	version := strings.Split(c.Object.Labels["openebs.io/version"], "-")[0]
	if version != strings.Split(from, "-")[0] && version != strings.Split(to, "-")[0] {
		return errors.Errorf(
			"cspi version %s is neither %s nor %s",
			c.Object.Labels["openebs.io/version"],
			from,
			to,
		)
	}
	return nil
}

// Patch ...
func (c *CSPI) Patch(from, to string) error {
	klog.Info("patching cspi ", c.Object.Name)
	version := c.Object.Labels["openebs.io/version"]
	if version == to {
		klog.Infof("cspi already in %s version", to)
		return nil
	}
	if version == from {
		patch := c.Data
		_, err := c.Client.CstorV1().CStorPoolInstances(c.Object.Namespace).Patch(
			context.TODO(),
			c.Object.Name,
			types.MergePatchType,
			[]byte(patch),
			metav1.PatchOptions{},
		)
		if err != nil {
			return errors.Wrapf(
				err,
				"failed to patch cspi %s",
				c.Object.Name,
			)
		}
		klog.Infof("cspi %s patched", c.Object.Name)
	}
	return nil
}

// Get ...
func (c *CSPI) Get(name, namespace string) error {
	cspi, err := c.Client.CstorV1().CStorPoolInstances(namespace).
		Get(context.TODO(), name, metav1.GetOptions{})
	if err != nil {
		return errors.Wrapf(err, "failed to get cspi %s in %s namespace", name, namespace)
	}
	c.Object = cspi
	return nil
}
