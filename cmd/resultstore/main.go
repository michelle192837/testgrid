package main

import (
	"context"
	"fmt"
	"time"

	tgresultstore "github.com/GoogleCloudPlatform/testgrid/resultstore"
	durationpb "github.com/golang/protobuf/ptypes/duration"
	timestamppb "github.com/golang/protobuf/ptypes/timestamp"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"google.golang.org/genproto/googleapis/devtools/resultstore/v2"
)

// fakeUpload uploads some demo results to verify API access.
func fakeUpload(ctx context.Context, client resultstore.ResultStoreUploadClient,statuses []resultstore.Status) error {
	now := time.Now()
	for i, status := range statuses {
		start := now.Add(time.Duration(-1 * (i+1)) * time.Hour)
		inv, err := createInvocation(ctx, client, status, start)
		if err != nil {
			return err
		}
		logrus.WithField("inv", inv).Info("Uploaded invocation.")
	}
	return nil
}

type invocUploadToken struct {
	invID, authToken string
}

func createInvocation(ctx context.Context, client resultstore.ResultStoreUploadClient, status resultstore.Status, start time.Time) (*resultstore.Invocation, error) {
	if client == nil {
		return nil, fmt.Errorf("no ResultStore upload client provided")
	}
	uploadToken := invocUploadToken{
		invID: uuid.New().String(),
		authToken: uuid.New().String(),
	}
	req := &resultstore.CreateInvocationRequest{
		InvocationId: uploadToken.invID,
		AuthorizationToken: uploadToken.authToken,
		Invocation: &resultstore.Invocation{
			Id: &resultstore.Invocation_Id{
				InvocationId: uploadToken.invID,
			},
			Timing: &resultstore.Timing{
				StartTime: &timestamppb.Timestamp{
					Seconds: start.Unix(),
				},
				Duration: &durationpb.Duration{
					Seconds: 10 * 60, // 10 minutes
				},
			},
			StatusAttributes: &resultstore.StatusAttributes{
				Status: status,
			},
			InvocationAttributes: &resultstore.InvocationAttributes{
				ProjectId:   "k8s-testgrid",
				Labels: []string{"testgrid", "demo"},
				Description: "Demo invocation for testing uploads.",
			},
		},
	}
	logrus.WithField("req", req).Info("Creating invocation.")
	inv, err := client.CreateInvocation(ctx, req)
	if err != nil {
		return nil, err
	}

	finalReq := &resultstore.FinalizeInvocationRequest{
		Name: inv.GetName(),
		AuthorizationToken: uploadToken.authToken,
	}
	_, err = client.FinalizeInvocation(ctx, finalReq)
	return inv, err
}

func main() {
	ctx := context.Background()
	conn, err := tgresultstore.Connect(ctx, "")
	if err != nil {
		logrus.WithError(err).Fatal("Failed to create ResultStore client.")
	}
	client := resultstore.NewResultStoreUploadClient(conn)
	statuses := []resultstore.Status{
		resultstore.Status_PASSED, 
		resultstore.Status_PASSED, 
		resultstore.Status_FAILED,
	}
	err = fakeUpload(ctx, client, statuses)
	if err != nil {
		logrus.WithError(err).Fatal("Failed to upload all invocations to ResultStore.")
	}
}
