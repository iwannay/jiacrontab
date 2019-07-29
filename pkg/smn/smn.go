package smn

import (
	"fmt"
        "errors"
	"github.com/SimpleMessageNotification/smn-sdk-go/smn-sdk-go/client"
)

func PublishMessageTemplate(domainName string, userName string, userPass string, region string, topicUrn string, templateName string, subject string, tags map[string]string) error {
	smnClient, err := client.NewClient(userName, domainName, userPass, region)
	if err != nil {
		panic(err)
	}

	request := smnClient.NewPublishMessageTemplateRequest()
	request.TopicUrn = topicUrn
	request.MessageTemplateName = templateName
        request.Subject = subject
        for k,v := range(tags) {
          request.Tags[k] = v
        }

	response, err := smnClient.PublishMessageTemplate(request)
	if err != nil {
		fmt.Println("the request is error ", err)
		return err
	}

	if !response.IsSuccess() {
		fmt.Printf("%#v\n", response.ErrorResponse)
		return errors.New("smn not success")
	}

	fmt.Printf("%#v\n", response)
        return nil
}

