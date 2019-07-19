/*
 * Copyright (C) 2017. Huawei Technologies Co., LTD. All rights reserved.
 *
 * This program is free software; you can redistribute it and/or modify
 * it under the terms of Apache License, Version 2.0.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
 * Apache License, Version 2.0 for more details.
 */
package util

import (
	"regexp"
	"encoding/json"
)

const (
	MaxTopicDisplayName = 192
	MaxTemplateContent  = 256 * 1024
	MaxSubjectLength    = 512
	MaxMessageLength    = 256 * 1024
	PhoneMatchPattern   = "^\\+?[0-9]{1}[0-9 /\\-]{1,31}$"
	TopicNamePattern    = "^[a-zA-Z0-9]{1}[-_a-zA-Z0-9]{0,255}$"
	TemplateNamePattern = "^[a-zA-Z0-9]{1}([-_a-zA-Z0-9]){0,64}"
	SubjectPattern      = "^[^\\r\\n\\t\\f]+$"
)

func ValidatePhone(phone string) bool {
	reg := regexp.MustCompile(PhoneMatchPattern)
	return reg.MatchString(phone)
}

func ValidateSmsEventType(eventType string) bool {
	return eventType == SmsFailEvent || eventType == SmsReplyEvent ||
		eventType == SmsSuccessEvent
}

func ValidateTopicName(topicName string) bool {
	if topicName == "" {
		return false
	}
	reg := regexp.MustCompile(TopicNamePattern)
	return reg.MatchString(topicName)
}

func ValidateTopicDiplayName(displayName string) bool {
	if displayName == "" {
		return true
	}
	bytes := []byte(displayName)
	return len(bytes) < MaxTopicDisplayName
}

func ValidateSubject(subject string) bool {
	if subject == "" {
		return true
	}
	bytes := []byte(subject)

	if len(bytes) > MaxSubjectLength {
		return false
	}

	reg := regexp.MustCompile(SubjectPattern)
	return reg.MatchString(subject)
}

func ValidateMessage(message string) bool {
	if message == "" {
		return false
	}
	bytes := []byte(message)
	return len(bytes) < MaxMessageLength
}

func ValidateMessageStructureLength(message string) bool {
	bytes := []byte(message)
	return len(bytes) < MaxMessageLength
}

func ValidateMessageStructureDefault(message string) bool {
	var result map[string]interface{}
	if err := json.Unmarshal([]byte(message), &result); err != nil {
		return false
	}

	defaultValue := result[Default]
	val, ok := defaultValue.(string)
	if !ok {
		return false
	}

	return val != ""
}

func ValidateMessageTemplateName(templateName string) bool {
	reg := regexp.MustCompile(TemplateNamePattern)
	return reg.MatchString(templateName)
}

func ValidateMessageTemplateContent(content string) bool {
	if content == "" {
		return false
	}
	bytes := []byte(content)
	return len(bytes) < MaxTemplateContent
}
