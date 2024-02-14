package main

import (
	"context"
	"fmt"
	"strings"

	iampb "cloud.google.com/go/iam/apiv1/iampb"
	run "cloud.google.com/go/run/apiv2"
	runpb "cloud.google.com/go/run/apiv2/runpb"
	"github.com/docker/docker/pkg/namesgenerator"
	"google.golang.org/api/option"
)

type GoogleCloudRun struct {}

// dagger -m github.com/vikram-dagger/daggerverse/google-cloud-run call deploy --project vikram-experiments --location us-central1 --image docker.io/nginx --http-port 80 --credential env:GOOGLE_CREDENTIAL

// dagger -m github.com/vikram-dagger/daggerverse/google-cloud-run call deploy --project vikram-experiments --location us-central1 --image docker.io/httpd --http-port 80 --credential env:GOOGLE_CREDENTIAL
func (m *GoogleCloudRun) Deploy(project string, location string, image string, httpPort int32, credential *Secret) (string, error) {
	ctx := context.Background()
	json, err := credential.Plaintext(ctx)
	b := []byte(json)
	gcrClient, err := run.NewServicesClient(ctx, option.WithCredentialsJSON(b))
	if err != nil {
		panic(err)
	}
	defer gcrClient.Close()

	name := strings.Replace(namesgenerator.GetRandomName(0), "_", "-", -1)

	gcrServiceRequest := &runpb.CreateServiceRequest{
		Parent:    fmt.Sprintf("projects/%s/locations/%s", project, location),
		ServiceId: name,
		Service: &runpb.Service{
			Ingress: runpb.IngressTraffic_INGRESS_TRAFFIC_ALL,
			Template: &runpb.RevisionTemplate{
				Containers: []*runpb.Container{
					{
						Image: image,
						Ports: []*runpb.ContainerPort{
							{
								Name:          "http1",
								ContainerPort: httpPort,
							},
						},
					},
				},
			},
		},
	}

	op, err := gcrClient.CreateService(ctx, gcrServiceRequest)
	if err != nil {
		panic(err)
	}

	resp, err := op.Wait(ctx)
	if err != nil {
		panic(err)
	}

	gcrIamRequest := &iampb.SetIamPolicyRequest{
		Resource: resp.Name,
		Policy: &iampb.Policy{
			Bindings: []*iampb.Binding{
				{
					Members: []string{"allUsers"},
					Role:    "roles/run.invoker",
				},
			},
		},
	}
	_, err = gcrClient.SetIamPolicy(ctx, gcrIamRequest)
	if err != nil {
		panic(err)
	}

	return resp.Uri, err

}
