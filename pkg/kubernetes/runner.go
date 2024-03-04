package kubernetes

import (
	"bufio"
	"context"
	"os"
	"time"

	"github.com/gg-mike/ccli/pkg/runner"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/tools/remotecommand"
)

func (client Client) createPod(namespace, name, imageName string) error {
	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{
					Name:  "worker",
					Image: imageName,
					Stdin: true,
					TTY:   true,
				},
			},
		},
	}

	if _, err := client.clientset.CoreV1().Pods(namespace).Create(context.Background(), pod, metav1.CreateOptions{}); err != nil {
		return err
	}
	if err := wait.PollUntilContextTimeout(context.TODO(), 1*time.Second, 5*time.Minute, true, func(ctx context.Context) (done bool, err error) {
		pod, err := client.clientset.CoreV1().Pods(namespace).Get(ctx, name, metav1.GetOptions{})
		if err != nil {
			return false, err
		}

		if pod.Status.Phase != corev1.PodRunning {
			return false, nil
		}

		return true, nil
	}); err != nil {
		return err
	}

	return nil
}

func (client Client) NewRunner(namespace, name, imageName, shell string) (*runner.Runner, error) {
	if err := client.createPod(namespace, name, imageName); err != nil {
		return &runner.Runner{}, err
	}

	stdinReader, stdinWriter, err := os.Pipe()
	if err != nil {
		return &runner.Runner{}, err
	}

	stdoutReader, stdoutWriter, err := os.Pipe()
	if err != nil {
		return &runner.Runner{}, err
	}

	req := client.clientset.CoreV1().RESTClient().
		Post().
		Namespace(namespace).
		Resource("pods").
		Name(name).
		SubResource("exec").
		Param("container", "worker").
		Param("stdin", "true").
		Param("stdout", "true").
		Param("stderr", "false").
		Param("tty", "false").
		Param("command", shell)

	exec, err := remotecommand.NewSPDYExecutor(client.config, "POST", req.URL())
	if err != nil {
		return &runner.Runner{}, err
	}

	go func() {
		exec.StreamWithContext(context.Background(), remotecommand.StreamOptions{
			Stdin:  stdinReader,
			Stdout: stdoutWriter,
			Stderr: nil,
			Tty:    false,
		})
	}()

	_runner := runner.NewRunner(bufio.NewWriter(stdinWriter), bufio.NewReader(stdoutReader))
	_runner.OnShutdown = func() error {
		return client.clientset.CoreV1().Pods(namespace).Delete(context.Background(), name, metav1.DeleteOptions{})
	}

	return _runner, nil
}
