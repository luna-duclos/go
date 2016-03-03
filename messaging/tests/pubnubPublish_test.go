// Package tests has the unit tests of package messaging.
// pubnubPublish_test.go contains the tests related to the publish requests on pubnub Api
package tests

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/pubnub/go/messaging"
	"github.com/stretchr/testify/assert"
)

// TestPublishStart prints a message on the screen to mark the beginning of
// publish tests.
// PrintTestMessage is defined in the common.go file.
func TestPublishStart(t *testing.T) {
	PrintTestMessage("==========Publish tests start==========")
}

// TestNullMessage sends out a null message to a pubnub channel. The response should
// be an "Invalid Message".
func TestNullMessage(t *testing.T) {
	assert := assert.New(t)
	pubnubInstance := messaging.NewPubnub(PubKey, SubKey, "", "", false, "")
	channel := "nullMessage"
	var message interface{}
	message = nil

	successChannel := make(chan []byte)
	errorChannel := make(chan []byte)

	go pubnubInstance.Publish(channel, message, successChannel, errorChannel)
	select {
	case msg := <-successChannel:
		assert.Fail("Response on success channel while expecting an error", string(msg))
	case err := <-errorChannel:
		assert.Contains(string(err), "Invalid Message")
	case <-timeout():
		assert.Fail("Publish timeout")
	}
}

// TestSuccessCodeAndInfo sends out a message to the pubnub channel
// The response is parsed and should match the 'sent' status.
// _publishSuccessMessage is defined in the common.go file
func TestSuccessCodeAndInfo(t *testing.T) {
	assert := assert.New(t)

	stop := NewVCRNonSubscribe("fixtures/publish/successCodeAndInfo",
		[]string{"uuid"}, 1)
	defer stop()

	pubnubInstance := messaging.NewPubnub(PubKey, SubKey, "", "", false, "")
	channel := "successCodeAndInfo"
	message := "Pubnub API Usage Example"

	successChannel := make(chan []byte)
	errorChannel := make(chan []byte)

	go pubnubInstance.Publish(channel, message, successChannel, errorChannel)
	select {
	case msg := <-successChannel:
		assert.Contains(string(msg), "1,")
		assert.Contains(string(msg), "\"Sent\",")
	case err := <-errorChannel:
		assert.Fail(string(err))
	case <-timeout():
		assert.Fail("Publish timeout")
	}
}

// TestSuccessCodeAndInfoWithEncryption sends out an encrypted
// message to the pubnub channel
// The response is parsed and should match the 'sent' status.
// _publishSuccessMessage is defined in the common.go file
func TestSuccessCodeAndInfoWithEncryption(t *testing.T) {
	assert := assert.New(t)

	stop := NewVCRNonSubscribe(
		"fixtures/publish/successCodeAndInfoWithEncryption", []string{"uuid"}, 1)
	defer stop()

	pubnubInstance := messaging.NewPubnub(PubKey, SubKey, "", "enigma", false, "")
	channel := "successCodeAndInfo"
	message := "Pubnub API Usage Example"

	successChannel := make(chan []byte)
	errorChannel := make(chan []byte)

	go pubnubInstance.Publish(channel, message, successChannel, errorChannel)
	select {
	case msg := <-successChannel:
		assert.Contains(string(msg), "1,")
		assert.Contains(string(msg), "\"Sent\",")
	case err := <-errorChannel:
		assert.Fail(string(err))
	case <-timeout():
		assert.Fail("Publish timeout")
	}
}

// TestSuccessCodeAndInfoForComplexMessage sends out a complex message to the pubnub channel
// The response is parsed and should match the 'sent' status.
// _publishSuccessMessage and customstruct is defined in the common.go file
func TestSuccessCodeAndInfoForComplexMessage(t *testing.T) {
	pubnubInstance := messaging.NewPubnub(PubKey, SubKey, "", "", false, "")
	channel := "testChannel"

	customStruct := CustomStruct{
		Foo: "hi!",
		Bar: []int{1, 2, 3, 4, 5},
	}

	returnChannel := make(chan []byte)
	errorChannel := make(chan []byte)
	responseChannel := make(chan string)
	waitChannel := make(chan string)

	go pubnubInstance.Publish(channel, customStruct, returnChannel, errorChannel)
	go ParsePublishResponse(returnChannel, channel, publishSuccessMessage, "SuccessCodeAndInfoForComplexMessage", responseChannel)

	go ParseErrorResponse(errorChannel, responseChannel)
	go WaitForCompletion(responseChannel, waitChannel)
	ParseWaitResponse(waitChannel, t, "SuccessCodeAndInfoForComplexMessage")
	time.Sleep(2 * time.Second)
}

// TestSuccessCodeAndInfoForComplexMessage2 sends out a complex message to the pubnub channel
// The response is parsed and should match the 'sent' status.
// _publishSuccessMessage and InitComplexMessage is defined in the common.go file
func TestSuccessCodeAndInfoForComplexMessage2(t *testing.T) {
	pubnubInstance := messaging.NewPubnub(PubKey, SubKey, "", "", false, "")
	channel := "testChannel"

	customComplexMessage := InitComplexMessage()

	returnChannel := make(chan []byte)
	errorChannel := make(chan []byte)
	responseChannel := make(chan string)
	waitChannel := make(chan string)

	go pubnubInstance.Publish(channel, customComplexMessage, returnChannel, errorChannel)
	go ParsePublishResponse(returnChannel, channel, publishSuccessMessage, "SuccessCodeAndInfoForComplexMessage2", responseChannel)

	go ParseErrorResponse(errorChannel, responseChannel)
	go WaitForCompletion(responseChannel, waitChannel)
	ParseWaitResponse(waitChannel, t, "SuccessCodeAndInfoForComplexMessage2")
	time.Sleep(2 * time.Second)
}

// TestSuccessCodeAndInfoForComplexMessage2WithSecretAndEncryption sends out an
// encypted and secret keyed complex message to the pubnub channel
// The response is parsed and should match the 'sent' status.
// _publishSuccessMessage and InitComplexMessage is defined in the common.go file
func TestSuccessCodeAndInfoForComplexMessage2WithSecretAndEncryption(t *testing.T) {
	pubnubInstance := messaging.NewPubnub(PubKey, SubKey, "secret", "enigma", false, "")
	channel := "testChannel"

	customComplexMessage := InitComplexMessage()

	returnChannel := make(chan []byte)
	errorChannel := make(chan []byte)
	responseChannel := make(chan string)
	waitChannel := make(chan string)

	go pubnubInstance.Publish(channel, customComplexMessage, returnChannel, errorChannel)
	go ParsePublishResponse(returnChannel, channel, publishSuccessMessage, "SuccessCodeAndInfoForComplexMessage2WithSecretAndEncryption", responseChannel)

	go ParseErrorResponse(errorChannel, responseChannel)
	go WaitForCompletion(responseChannel, waitChannel)
	ParseWaitResponse(waitChannel, t, "SuccessCodeAndInfoForComplexMessage2WithSecretAndEncryption")
	time.Sleep(2 * time.Second)
}

// TestSuccessCodeAndInfoForComplexMessage2WithEncryption sends out an
// encypted complex message to the pubnub channel
// The response is parsed and should match the 'sent' status.
// _publishSuccessMessage and InitComplexMessage is defined in the common.go file
func TestSuccessCodeAndInfoForComplexMessage2WithEncryption(t *testing.T) {
	pubnubInstance := messaging.NewPubnub(PubKey, SubKey, "", "enigma", false, "")
	channel := "testChannel"

	customComplexMessage := InitComplexMessage()

	returnChannel := make(chan []byte)
	errorChannel := make(chan []byte)
	responseChannel := make(chan string)
	waitChannel := make(chan string)

	go pubnubInstance.Publish(channel, customComplexMessage, returnChannel, errorChannel)
	go ParsePublishResponse(returnChannel, channel, publishSuccessMessage, "SuccessCodeAndInfoForComplexMessage2WithEncryption", responseChannel)

	go ParseErrorResponse(errorChannel, responseChannel)
	go WaitForCompletion(responseChannel, waitChannel)
	ParseWaitResponse(waitChannel, t, "SuccessCodeAndInfoForComplexMessage2WithEncryption")
	time.Sleep(2 * time.Second)
}

func TestPublishStringWithSerialization(t *testing.T) {
	assert := assert.New(t)
	pubnubInstance := messaging.NewPubnub(PubKey, SubKey, "", "", false, "")
	channel := "testChannel"
	messageToPost := "{\"name\": \"Alex\", \"age\": \"123\"}"

	successChannel := make(chan []byte)
	errorChannel := make(chan []byte)

	subscribeSuccessChannel := make(chan []byte)
	subscribeErrorChannel := make(chan []byte)

	await := make(chan bool)

	go pubnubInstance.Subscribe(channel, "", subscribeSuccessChannel, false,
		subscribeErrorChannel)
	ExpectConnectedEvent(t, channel, "", subscribeSuccessChannel,
		subscribeErrorChannel)

	go func() {
		select {
		case message := <-subscribeSuccessChannel:
			var response []interface{}
			var msgs []interface{}
			var err error

			err = json.Unmarshal(message, &response)
			if err != nil {
				assert.Fail(err.Error())
			}

			switch t := response[0].(type) {
			case []interface{}:
				var messageToPostMap map[string]interface{}

				msgs = response[0].([]interface{})
				err := json.Unmarshal([]byte(messageToPost), &messageToPostMap)
				if err != nil {
					assert.Fail(err.Error())
				}

				assert.Equal(messageToPost, msgs[0])
			default:
				assert.Fail("Unexpected response type%s: ", t)
			}

			await <- true
		case err := <-subscribeErrorChannel:
			assert.Fail(string(err))
			await <- false
		case <-timeouts(10):
			assert.Fail("Timeout")
			await <- false
		}
	}()

	go pubnubInstance.Publish(channel, messageToPost, successChannel, errorChannel)

	<-await
}

func TestPublishStringWithoutSerialization(t *testing.T) {
	assert := assert.New(t)
	pubnubInstance := messaging.NewPubnub(PubKey, SubKey, "", "", false, "")
	channel := "testChannel"
	messageToPost := "{\"name\": \"Alex\", \"age\": \"123\"}"

	successChannel := make(chan []byte)
	errorChannel := make(chan []byte)

	subscribeSuccessChannel := make(chan []byte)
	subscribeErrorChannel := make(chan []byte)

	await := make(chan bool)

	go pubnubInstance.Subscribe(channel, "", subscribeSuccessChannel, false,
		subscribeErrorChannel)
	ExpectConnectedEvent(t, channel, "", subscribeSuccessChannel,
		subscribeErrorChannel)

	go func() {
		select {
		case message := <-subscribeSuccessChannel:
			var response []interface{}
			var msgs []interface{}
			var err error

			err = json.Unmarshal(message, &response)
			if err != nil {
				assert.Fail(err.Error())
			}

			switch t := response[0].(type) {
			case []interface{}:
				var messageToPostMap map[string]interface{}

				msgs = response[0].([]interface{})
				err := json.Unmarshal([]byte(messageToPost), &messageToPostMap)
				if err != nil {
					assert.Fail(err.Error())
				}

				assert.Equal(messageToPostMap, msgs[0])
			default:
				assert.Fail("Unexpected response type%s: ", t)
			}

			await <- true
		case err := <-subscribeErrorChannel:
			assert.Fail(string(err))
			await <- false
		case <-timeouts(10):
			assert.Fail("Timeout")
			await <- false
		}
	}()

	go pubnubInstance.PublishExtended(channel, messageToPost, false, true,
		successChannel, errorChannel)

	<-await
}

// ParsePublishResponse parses the response from the pubnub api to validate the
// sent status.
func ParsePublishResponse(returnChannel chan []byte, channel string, message string, testname string, responseChannel chan string) {
	for {
		value, ok := <-returnChannel
		if !ok {
			break
		}
		if string(value) != "[]" {
			response := fmt.Sprintf("%s", value)
			//fmt.Println("Test '" + testname + "':" +response)
			if strings.Contains(response, message) {
				responseChannel <- "Test '" + testname + "': passed."
				break
			} else {
				responseChannel <- "Test '" + testname + "': failed."
				break
			}
		}
	}
}

// TestLargeMessage tests the client by publshing a large message
// An error "message to large" should be returned from the server
/*func TestLargeMessage(t *testing.T) {
	pubnubInstance := messaging.NewPubnub(PubKey, SubKey, SecKey, "", false, "")
	channel := "testChannel"
	message := "This is a large message test which will return an error message. This is a large message test which will return an error message. This is a large message test which will return an error message. This is a large message test which will return an error message. This is a large message test which will return an error message. This is a large message test which will return an error message. This is a large message test which will return an error message. This is a large message test which will return an error message. This is a large message test which will return an error message. This is a large message test which will return an error message. This is a large message test which will return an error message. This is a large message test which will return an error message. This is a large message test which will return an error message. This is a large message test which will return an error message. This is a large message test which will return an error message. This is a large message test which will return an error message. This is a large message test which will return an error message. This is a large message test which will return an error message. This is a large message test which will return an error message. This is a large message test which will return an error message. This is a large message test which will return an error message. This is a large message test which will return an error message. This is a large message test which will return an error message. This is a large message test which will return an error message. This is a large message test which will return an error message. This is a large message test which will return an error message. This is a large message test which will return an error message. This is a large message test which will return an error message. This is a large message test which will return an error message. This is a large message test which will return an error message. This is a large message test which will return an error message. This is a large message test which will return an error message. "
	returnChannel := make(chan []byte)
	errorChannel := make(chan []byte)
	responseChannel := make(chan string)
	waitChannel := make(chan string)

	go pubnubInstance.Publish(channel, message, returnChannel, errorChannel)
	go ParseLargeResponse("Message Too Large", errorChannel, responseChannel)
	go WaitForCompletion(responseChannel, waitChannel)
	ParseWaitResponse(waitChannel, t, "MessageTooLarge")
}*/

// ParseLargeResponse parses the returnChannel and matches the message m
//
// Parameters:
// m: message to compare
// returnChannel: the channel to read
// responseChannel: the channel to send a response to.
func ParseLargeResponse(m string, returnChannel chan []byte, responseChannel chan string) {
	for {
		value, ok := <-returnChannel
		if !ok {
			break
		}
		returnVal := string(value)
		if returnVal != "[]" {
			var s []interface{}
			errJSON := json.Unmarshal(value, &s)

			if (errJSON == nil) && (len(s) > 0) {
				if message, ok := s[1].(string); ok {
					if message == m {
						responseChannel <- "passed"
					} else {
						responseChannel <- "failed"
					}
				} else {
					responseChannel <- "failed"
				}
			} else {
				responseChannel <- "failed"
			}
			break
		}
	}
}

// TestPublishEnd prints a message on the screen to mark the end of
// publish tests.
// PrintTestMessage is defined in the common.go file.
func TestPublishEnd(t *testing.T) {
	PrintTestMessage("==========Publish tests end==========")
}
