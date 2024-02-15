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

type GoogleCloudRun struct{}

// dagger -m github.com/vvaswani/daggerverse/google-cloud-run call create-service --project vikram-experiments --location us-central1 --image docker.io/nginx --http-port 80 --credential env:GOOGLE_CREDENTIAL

// dagger -m github.com/vvaswani/daggerverse/google-cloud-run call create-service --project vikram-experiments --location us-central1 --image docker.io/httpd --http-port 80 --credential env:GOOGLE_CREDENTIAL
func (m *GoogleCloudRun) CreateService(project string, location string, image string, httpPort int32, credential *Secret) (string, error) {
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

	gcrOperation, err := gcrClient.CreateService(ctx, gcrServiceRequest)
	if err != nil {
		panic(err)
	}

	gcrResponse, err := gcrOperation.Wait(ctx)
	if err != nil {
		panic(err)
	}

	gcrIamRequest := &iampb.SetIamPolicyRequest{
		Resource: gcrResponse.Name,
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

	return gcrResponse.Uri, err

}

// dagger -m github.com/vikram-dagger/daggerverse/google-cloud-run call update-service --project vikram-experiments --location us-central1 --service myservice --image docker.io/nginx --http-port 80 --credential env:GOOGLE_CREDENTIAL

func (m *GoogleCloudRun) UpdateService(project string, location string, service string, image string, httpPort int32, credential *Secret) (string, error) {
	ctx := context.Background()
	json, err := credential.Plaintext(ctx)
	b := []byte(json)
	gcrClient, err := run.NewServicesClient(ctx, option.WithCredentialsJSON(b))
	if err != nil {
		panic(err)
	}
	defer gcrClient.Close()

	gcrServiceRequest := &runpb.UpdateServiceRequest{
		Service: &runpb.Service{
			Name:    fmt.Sprintf("projects/%s/locations/%s/services/%s", project, location, service),
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

	gcrOperation, err := gcrClient.UpdateService(ctx, gcrServiceRequest)
	if err != nil {
		panic(err)
	}

	gcrResponse, err := gcrOperation.Wait(ctx)
	if err != nil {
		panic(err)
	}

	return gcrResponse.Uri, err

}
