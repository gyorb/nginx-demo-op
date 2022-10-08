/*
Copyright 2022.

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

package controllers

import (
	"context"
	"fmt"
	"net"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	networkv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/intstr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	nginxopv1 "my.domain/nginxop/api/v1"
)

// NginxOpReconciler reconciles a NginxOp object
type NginxOpReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=nginxop.my.domain,resources=nginxops,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=nginxop.my.domain,resources=nginxops/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=nginxop.my.domain,resources=nginxops/finalizers,verbs=update
//+kubebuilder:rbac:groups=apps,resources=deployments,verbs=patch
//+kubebuilder:rbac:groups="",resources=services,verbs=patch
//+kubebuilder:rbac:groups=networking.k8s.io,resources=ingresses,verbs=patch

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the NginxOp object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.13.0/pkg/reconcile
func (r *NginxOpReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := log.FromContext(ctx)

	log.Info("new request", "request", req.String())
	log.Info(req.String())

	nginxop := &nginxopv1.NginxOp{}

	if err := r.Client.Get(ctx, req.NamespacedName, nginxop); err != nil {
		log.Error(err, "no nginxop CRD object")
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	msg := fmt.Sprintf("received reconcile request for %q (namespace: %q)", nginxop.GetName(), nginxop.GetNamespace())
	log.Info(msg)

	// Create Deployment
	deployment, err := r.desiredDeployment(*nginxop)
	if err != nil {
		return ctrl.Result{}, err
	}

	applyOpts := []client.PatchOption{client.ForceOwnership, client.FieldOwner(nginxop.Name)}

	err = r.Patch(ctx, &deployment, client.Apply, applyOpts...)
	if err != nil {
		return ctrl.Result{}, err
	}

	// Create service
	svc, err := r.desiredService(*nginxop)
	if err != nil {
		return ctrl.Result{}, err
	}
	err = r.Patch(ctx, &svc, client.Apply, applyOpts...)
	if err != nil {
		return ctrl.Result{}, err
	}

	// Create ingress
	ingressRule, err := r.desiredIngressRule(*nginxop)
	if err != nil {
		return ctrl.Result{}, client.IgnoreAlreadyExists(err)
	}
	err = r.Patch(ctx, &ingressRule, client.Apply, applyOpts...)
	if err != nil {
		return ctrl.Result{}, err
	}

	// Update status
	nginxop.Status.URL = urlForService(ingressRule, 80)
	err = r.Status().Update(ctx, nginxop)
	if err != nil {
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

func (r *NginxOpReconciler) desiredDeployment(nginxop nginxopv1.NginxOp) (appsv1.Deployment, error) {

	depl := appsv1.Deployment{
		TypeMeta: metav1.TypeMeta{APIVersion: appsv1.SchemeGroupVersion.String(), Kind: "Deployment"},
		ObjectMeta: metav1.ObjectMeta{
			Name:      nginxop.Name,
			Namespace: nginxop.Namespace,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &nginxop.Spec.Replicas, // won't be nil because defaulting
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{"nginx": nginxop.Name},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{"nginx": nginxop.Name, "app": "nginx"},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  "nginx",
							Image: nginxop.Spec.Image,
							Ports: []corev1.ContainerPort{
								{ContainerPort: 80, Name: "http", Protocol: "TCP"},
								//{ContainerPort: 443, Name: "https", Protocol: "TCP"},
							},
						},
					},
				},
			},
		},
	}

	if err := ctrl.SetControllerReference(&nginxop, &depl, r.Scheme); err != nil {
		return depl, err
	}

	return depl, nil
}

func (r *NginxOpReconciler) desiredService(nginxop nginxopv1.NginxOp) (corev1.Service, error) {
	svc := corev1.Service{
		TypeMeta: metav1.TypeMeta{APIVersion: corev1.SchemeGroupVersion.String(), Kind: "Service"},
		ObjectMeta: metav1.ObjectMeta{
			Name:      nginxop.Name,
			Namespace: nginxop.Namespace,
		},
		Spec: corev1.ServiceSpec{
			Ports: []corev1.ServicePort{
				{Name: "http", Port: 80, Protocol: "TCP", TargetPort: intstr.FromString("http")},
			},
			Selector: map[string]string{"nginx": nginxop.Name},
			Type:     corev1.ServiceTypeLoadBalancer,
		},
	}

	// always set the controller reference so that we know which object owns this.
	if err := ctrl.SetControllerReference(&nginxop, &svc, r.Scheme); err != nil {
		return svc, err
	}

	return svc, nil
}

func (r *NginxOpReconciler) desiredIngressRule(nginxop nginxopv1.NginxOp) (networkv1.Ingress, error) {

	var defaultPathType = networkv1.PathTypePrefix

	ingress := networkv1.Ingress{
		TypeMeta: metav1.TypeMeta{APIVersion: networkv1.SchemeGroupVersion.String(), Kind: "Ingress"},
		ObjectMeta: metav1.ObjectMeta{
			Name:      nginxop.Name,
			Namespace: nginxop.Namespace,
			Labels:    map[string]string{"cert-manager.io/issuer": "test-selfsigned"},
		},
		Spec: networkv1.IngressSpec{
			TLS: []networkv1.IngressTLS{
				{Hosts: []string{nginxop.Spec.Host}, SecretName: "selfsigned-cert-tls"}},
			Rules: []networkv1.IngressRule{
				{Host: nginxop.Spec.Host,
					IngressRuleValue: networkv1.IngressRuleValue{
						HTTP: &networkv1.HTTPIngressRuleValue{
							Paths: []networkv1.HTTPIngressPath{
								{
									Path:     "/",
									PathType: &defaultPathType,
									Backend: networkv1.IngressBackend{
										Service: &networkv1.IngressServiceBackend{
											Name: nginxop.Name,
											Port: networkv1.ServiceBackendPort{
												Number: 80,
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}

	// always set the controller reference so that we know which object owns this.
	if err := ctrl.SetControllerReference(&nginxop, &ingress, r.Scheme); err != nil {
		return ingress, err
	}

	return ingress, nil
}

func urlForService(ing networkv1.Ingress, port int32) string {
	if len(ing.Status.LoadBalancer.Ingress) == 0 {
		return ""
	}

	host := ing.Status.LoadBalancer.Ingress[0].Hostname
	if host == "" {
		host = ing.Status.LoadBalancer.Ingress[0].IP
	}

	return fmt.Sprintf("http://%s", net.JoinHostPort(host, fmt.Sprintf("%v", port)))
}

// SetupWithManager sets up the controller with the Manager.
func (r *NginxOpReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&nginxopv1.NginxOp{}).
		Complete(r)
}
