package helper

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	appsv1 "k8s.io/api/apps/v1" // Import appsv1 package
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"os"
	"os/exec"
	"strings"
)

func UnusedPort() (string, error) {
	scriptPath := "unused_port.sh"
	cmd := exec.Command("/bin/bash", scriptPath)

	outputChan := make(chan string)
	errChan := make(chan error)

	go func() {
		output, err := cmd.Output()
		if err != nil {
			errChan <- err
			return
		}
		unusedPort := strings.TrimSpace(string(output))
		outputChan <- unusedPort
	}()

	select {
	case unusedPort := <-outputChan:
		return unusedPort, nil
	case err := <-errChan:
		fmt.Printf("Error executing script: %v\n", err)
		return "", err
	}
}

func Exec(bashCommand string, args []string) error {
	cmd := exec.Command(bashCommand, args...)
	var stdoutBuf, stderrBuf bytes.Buffer
	cmd.Stdout = &stdoutBuf
	cmd.Stderr = &stderrBuf

	if err := cmd.Run(); err != nil {
		return errors.New("exec err: " + err.Error())
	}

	fmt.Println(stdoutBuf.String())
	fmt.Println(stderrBuf.String())

	return nil
}

func CreateClient(kubeconfigPath string) (kubernetes.Interface, error) {
	var kubeconfig *rest.Config

	if kubeconfigPath != "" {
		config, err := clientcmd.BuildConfigFromFlags("", kubeconfigPath)
		if err != nil {
			return nil, fmt.Errorf("unable to load kubeconfig from %s: %v", kubeconfigPath, err)
		}
		kubeconfig = config
	} else {
		config, err := rest.InClusterConfig()
		if err != nil {
			return nil, fmt.Errorf("unable to load in-cluster config: %v", err)
		}
		kubeconfig = config
	}

	client, err := kubernetes.NewForConfig(kubeconfig)
	if err != nil {
		return nil, fmt.Errorf("unable to create a client: %v", err)
	}

	return client, nil
}

func int32Ptr(i int32) *int32 {
	return &i
}

func CreateDeployment(
	clientset kubernetes.Interface,
	namespace,
	name,
	containerName,
	image string,
	replicas int32,
	labels map[string]string,
	ports []apiv1.ContainerPort,
) error {
	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: int32Ptr(replicas),
			Selector: &metav1.LabelSelector{
				MatchLabels: labels,
			},
			Template: apiv1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: labels,
				},
				Spec: apiv1.PodSpec{
					Containers: []apiv1.Container{
						{
							Name:  containerName,
							Image: image,
							Ports: ports,
						},
					},
				},
			},
		},
	}

	_, err := clientset.AppsV1().Deployments(namespace).Create(
		context.Background(),
		deployment,
		metav1.CreateOptions{},
	)
	if err != nil {
		fmt.Printf("Error creating Deployment: %s\n", err.Error())
		os.Exit(1)
	}

	fmt.Println("Deployment created successfully!")
	return nil
}

func DeleteDeployment(clientset kubernetes.Interface, namespace, name string) error {
	err := clientset.AppsV1().Deployments(namespace).Delete(
		context.Background(),
		name,
		metav1.DeleteOptions{},
	)
	if err != nil {
		return err
	}

	return nil
}

func CreateStatefulSet(
	clientset kubernetes.Interface,
	namespace,
	name,
	serviceName,
	containerName,
	image string,
	replicas int32,
	labels map[string]string,
	ports []apiv1.ContainerPort,
	volumeClaimTemplates []apiv1.PersistentVolumeClaim,
) error {
	statefulset := &appsv1.StatefulSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: appsv1.StatefulSetSpec{
			Replicas: int32Ptr(replicas),
			Selector: &metav1.LabelSelector{
				MatchLabels: labels,
			},
			ServiceName: serviceName,
			Template: apiv1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: labels,
				},
				Spec: apiv1.PodSpec{
					Affinity: &apiv1.Affinity{
						NodeAffinity: &apiv1.NodeAffinity{
							RequiredDuringSchedulingIgnoredDuringExecution: &apiv1.NodeSelector{
								NodeSelectorTerms: []apiv1.NodeSelectorTerm{
									{
										MatchExpressions: []apiv1.NodeSelectorRequirement{
											{
												Key:      "kubernetes.io/hostname",
												Operator: apiv1.NodeSelectorOpIn,
												Values:   []string{"k8s-worker3"},
											},
										},
									},
								},
							},
						},
					},
					Containers: []apiv1.Container{
						{
							Name:  containerName,
							Image: image,
							Command: []string{
								"/bin/sh",
							},
							Args: []string{
								"-c",
								`echo "jupyter:jupyter" | chpasswd &&
								cp /etc/jupyterhub/jupyterhub_config.py jupyterhub_config.py &&
								jupyterhub --config jupyterhub_config.py`,
							},
							Ports: ports,
							EnvFrom: []apiv1.EnvFromSource{
								{
									ConfigMapRef: &apiv1.ConfigMapEnvSource{
										LocalObjectReference: apiv1.LocalObjectReference{
											Name: "aws-env-g45526fc96",
										},
									},
								},
								{
									ConfigMapRef: &apiv1.ConfigMapEnvSource{
										LocalObjectReference: apiv1.LocalObjectReference{
											Name: "hadoop-env-544cf4678g",
										},
									},
								},
							},
							Env: []apiv1.EnvVar{
								{
									Name:  "HADOOP_USER_NAME",
									Value: "hdfs",
								},
								// Add other environment variables as needed
							},
							VolumeMounts: []apiv1.VolumeMount{
								// Add volume mounts as needed
							},
							WorkingDir: "/home/zookage",
						},
					},
				},
			},
			VolumeClaimTemplates: volumeClaimTemplates,
		},
	}

	_, err := clientset.AppsV1().StatefulSets(namespace).Create(
		context.Background(),
		statefulset,
		metav1.CreateOptions{},
	)
	if err != nil {
		fmt.Printf("Error creating StatefulSet: %s\n", err.Error())
		return err
	}

	fmt.Println("StatefulSet created successfully!")
	return nil
}
