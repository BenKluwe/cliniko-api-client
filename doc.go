/*
The package cliniko provides the means to interact with the cliniko api
from golang.

To get started create a new ClinikoClient:

	client, err := NewClinikoClient(
		"token",
		"vendor",
		"vendor email",
	)

Where token is the the token copied directly from the Cliniko API. The shard is deduced from the token.
The vendor name and email will be passed in the User-Agent field with each outgoing request.

Use any *WithResponse function to execute a query and get a parsed response:

	page, perPage, sort, order :=
		0, 10, []string{"id"},
		ListAppointmentTypesGetParamsOrder("desc")

	appointments, err :=
		client.ListAppointmentTypesGetWithResponse(
			context.TODO(),
			&ListAppointmentTypesGetParams{
				Page:    &page,
				PerPage: &perPage,
				Sort:    &sort,
				Order:   &order,
			})

	log.Println(appointments.JSON200.TotalEntries)

One special case exists for creating an attachment as this is a multi-step process:

	contents := []byte{0}
	fileDescription := "file description as show by Cliniko"
	presignedURLResponse,
	AmazonS3BucketResponse,
	createAttachmentResponse,
	err :=
		client.CreateAttachment(
			context.TODO(),
			"patientId",
			&fileDescription,
			"filename",
			bytes.NewReader(contents))

*/

package cliniko
