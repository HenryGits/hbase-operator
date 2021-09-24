/*
Copyright 2021.

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

package hbase

import (
	"context"
	"gitee.com/dmcca/gotools/kubernetes"
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/util/json"
	"k8s.io/klog/v2"
	"reflect"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	"gitee.com/dmcca/gotools/os"
	"gitee.com/dmcca/gotools/template"
	hbasev1 "github.com/HenryGits/hbase-operator/apis/hbase/v1"
)

const FieldManager string = "hbase-operator"
const Finalizer string = "finalizer.hbase.operator.dameng.com"

// HbaseReconciler reconciles a Hbase object
type HbaseReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

var HBaseTpl = template.Parser{
	Directory: os.EnvVar("GT_TEMPLATE_PATH", "/etc/operator/templates"),
	Pattern:   "\\.gotmpl$",
}

//+kubebuilder:rbac:groups=hbase.dameng.com,resources=hbases,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=hbase.dameng.com,resources=hbases/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=hbase.dameng.com,resources=hbases/finalizers,verbs=update
//+kubebuilder:rbac:groups=apps,resources=statefulsets,verbs=get;list;watch;create;update;patch;delete

func (r *HbaseReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)
	logger.Info("==========HbaseReconciler begin to Reconcile==========")

	var origin = &hbasev1.Hbase{}
	if err := r.Get(ctx, req.NamespacedName, origin); err != nil {
		logger.Error(err, "unable to fetch hbase")
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}
	hbase := origin.DeepCopy()
	//hadoop.Status.Phase = hadoopv1.Reconciling

	// examine DeletionTimestamp to determine if object is under deletion
	if hbase.ObjectMeta.DeletionTimestamp.IsZero() {
		// The object is not being deleted, so if it does not have our finalizer,
		// then lets add the finalizer and update the object. This is equivalent
		// registering our finalizer.
		if !containsString(hbase.GetFinalizers(), Finalizer) {
			controllerutil.AddFinalizer(hbase, Finalizer)
			if err := r.Update(ctx, hbase); err != nil {
				logger.Error(err, "unable to update hadoop")
				return ctrl.Result{}, err
			}
		}

		runtimeObjects, err := r.generateRuntimeObjects(hbase)
		if err != nil {
			logger.Error(err, "decode error")
			return ctrl.Result{}, err
		}

		for _, object := range runtimeObjects {
			// set spec annotation
			object, err := setSpecAnnotation((*object).(*unstructured.Unstructured))
			if err != nil {
				logger.Error(err, "set annotation error")
				return ctrl.Result{}, err
			}
			// set namespace
			object.SetNamespace(hbase.Namespace)
			// set controller reference
			if err := controllerutil.SetControllerReference(hbase, object, r.Scheme); err != nil {
				logger.Error(err, "maintain hadoop controller reference error")
				return ctrl.Result{}, err
			}

			logger.V(6).Info("object content:", "object", object)

			// retrieve hbase
			var originObject unstructured.Unstructured
			originObject.SetGroupVersionKind(object.GroupVersionKind())
			if err := r.Get(ctx, client.ObjectKey{Namespace: hbase.Namespace, Name: object.GetName()}, &originObject); err != nil {
				// create object if not found
				if errors.IsNotFound(err) {
					if err := r.Create(ctx, object, &client.CreateOptions{FieldManager: FieldManager}); err != nil {
						logger.Error(err, "Object create error", "Object", object)

						return ctrl.Result{}, err
					}
				} else {
					logger.Error(err, "get hbase Object error")
					return ctrl.Result{}, err
				}
			} else {
				// continue if hbase equal new
				equal, err := objectEqual(&originObject, object)
				if err != nil {
					logger.Error(err, "Object equal error")
					return ctrl.Result{}, err
				}
				logger.V(4).Info("deep equal", "kind", object.GroupVersionKind(), "result", equal)
				if equal {
					continue
				}

				// patch if hbase not equal new
				if err := r.Patch(ctx, object, client.Merge, &client.PatchOptions{FieldManager: FieldManager}); err != nil {
					logger.Error(err, "Object patch error", "Object", object)

					return ctrl.Result{}, err
				}
			}
		}
	} else {
		// The object is being deleted
		if containsString(hbase.GetFinalizers(), Finalizer) {
			// our finalizer is present, so lets handle any external dependency
			//if err := r.deleteExternalResources(ctx, hbase); err != nil {
			//	// if fail to delete the external dependency here, return with error
			//	// so that it can be retried
			//	return ctrl.Result{}, err
			//}

			// remove our finalizer from the list and update it.
			controllerutil.RemoveFinalizer(hbase, Finalizer)
			if err := r.Update(ctx, hbase); err != nil {
				return ctrl.Result{}, err
			}
		}
	}

	if !reflect.DeepEqual(hbase.TypeMeta, hbase.TypeMeta) || !reflect.DeepEqual(hbase.ObjectMeta, hbase.ObjectMeta) || !reflect.DeepEqual(hbase.Spec, hbase.Spec) {
		if err := r.Update(ctx, hbase); err != nil {
			logger.Error(err, "update without status error")
			return ctrl.Result{}, err
		}
	}

	if !reflect.DeepEqual(hbase.Status, hbase.Status) {
		if err := r.Status().Update(ctx, hbase); err != nil {
			logger.Error(err, "update status error")
			return ctrl.Result{}, err
		}
	}

	return ctrl.Result{}, nil
}

// Helper functions to check and remove string from a slice of strings.
func containsString(slice []string, s string) bool {
	for _, item := range slice {
		if item == s {
			return true
		}
	}
	return false
}

func objectEqual(hbase, newer *unstructured.Unstructured) (bool, error) {
	newSpec, found, err := unstructured.NestedFieldNoCopy(newer.Object, "spec")
	if err != nil {
		klog.Errorf("nested spec error: %v", err)
		return false, err
	}
	if !found {
		return false, nil
	}

	originSpecJSON, ok := hbase.GetAnnotations()["operator.dameng.com/spec"]
	if !ok {
		return false, nil
	}

	klog.V(6).Infof("hbase spec json: %v", originSpecJSON)

	var originSpec map[string]interface{}
	if err := json.Unmarshal([]byte(originSpecJSON), &originSpec); err != nil {
		klog.Errorf("unmarshal error: %v", err)
		return false, err
	}

	if reflect.DeepEqual(newSpec, originSpec) {
		return true, nil
	}
	return false, nil
}

func setSpecAnnotation(object *unstructured.Unstructured) (*unstructured.Unstructured, error) {
	spec, found, err := unstructured.NestedFieldNoCopy(object.Object, "spec")
	if err != nil {
		klog.Errorf("nested spec error: %v", err)
		return nil, err
	}
	if !found {
		return object, nil
	}
	annotation, err := json.Marshal(spec)
	if err != nil {
		klog.Errorf("marshal spec error: %v", err)
		return nil, err
	}
	annotations := object.GetAnnotations()
	if annotations == nil {
		annotations = make(map[string]string)
	}
	annotations["operator.dameng.com/spec"] = string(annotation)
	object.SetAnnotations(annotations)
	return object, nil
}

func (r *HbaseReconciler) generateRuntimeObjects(hbase *hbasev1.Hbase) (runtimeObjects []*runtime.Object, err error) {
	templates, err := HBaseTpl.ParseTemplate("hbase.dameng.com_hbase.gotmpl", hbase)
	if err != nil {
		klog.Errorf("generate hbase runtime error: %v", err)
		return runtimeObjects, err
	}
	klog.V(8).Infof("template: %s", templates)
	return kubernetes.ParseYaml(templates)
}

// SetupWithManager sets up the controller with the Manager.
func (r *HbaseReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&hbasev1.Hbase{}).
		Owns(&appsv1.StatefulSet{}).
		Complete(r)
}
