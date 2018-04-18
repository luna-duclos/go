package pubnub

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"reflect"
	"strconv"

	"github.com/pubnub/go/pnerr"
	"github.com/pubnub/go/utils"

	"net/http"
	"net/url"
)

const PUBLISH_GET_PATH = "/publish/%s/%s/0/%s/%s/%s"
const PUBLISH_POST_PATH = "/publish/%s/%s/0/%s/%s"

var emptyPublishResponse *PublishResponse

type publishOpts struct {
	pubnub *PubNub

	Ttl     int
	Channel string
	Message interface{}
	Meta    interface{}

	UsePost        bool
	ShouldStore    bool
	Serialize      bool
	DoNotReplicate bool

	Transport http.RoundTripper

	ctx Context

	// nil hacks
	setTtl         bool
	setShouldStore bool
}

type PublishResponse struct {
	Timestamp int
}

type publishBuilder struct {
	opts *publishOpts
}

func newPublishResponse(jsonBytes []byte, status StatusResponse) (
	*PublishResponse, StatusResponse, error) {
	var value []interface{}

	err := json.Unmarshal(jsonBytes, &value)
	if err != nil {
		e := pnerr.NewResponseParsingError("Error unmarshalling response",
			ioutil.NopCloser(bytes.NewBufferString(string(jsonBytes))), err)

		return emptyPublishResponse, status, e
	}

	timeString := value[2].(string)
	timestamp, err := strconv.Atoi(timeString)
	if err != nil {
		return emptyPublishResponse, status, err
	}

	return &PublishResponse{
		Timestamp: timestamp,
	}, status, nil
}

func newPublishBuilder(pubnub *PubNub) *publishBuilder {
	builder := publishBuilder{
		opts: &publishOpts{
			pubnub:    pubnub,
			Serialize: true,
		},
	}

	return &builder
}

func newPublishBuilderWithContext(pubnub *PubNub, context Context) *publishBuilder {
	builder := publishBuilder{
		opts: &publishOpts{
			pubnub: pubnub,
			ctx:    context,
		},
	}

	return &builder
}

func (b *publishBuilder) Ttl(ttl int) *publishBuilder {
	b.opts.Ttl = ttl
	b.opts.setTtl = true

	return b
}

func (b *publishBuilder) Channel(ch string) *publishBuilder {
	b.opts.Channel = ch

	return b
}

func (b *publishBuilder) Message(msg interface{}) *publishBuilder {
	b.opts.Message = msg

	return b
}

func (b *publishBuilder) Meta(meta interface{}) *publishBuilder {
	b.opts.Meta = meta

	return b
}

func (b *publishBuilder) UsePost(post bool) *publishBuilder {
	b.opts.UsePost = post

	return b
}

func (b *publishBuilder) ShouldStore(store bool) *publishBuilder {
	b.opts.ShouldStore = store
	if store {
		b.opts.setShouldStore = true
	} else {
		b.opts.setShouldStore = false
	}

	return b
}

func (b *publishBuilder) Serialize(serialize bool) *publishBuilder {
	b.opts.Serialize = serialize

	return b
}

func (b *publishBuilder) DoNotReplicate(repl bool) *publishBuilder {
	b.opts.DoNotReplicate = repl

	return b
}

func (b *publishBuilder) Transport(tr http.RoundTripper) *publishBuilder {
	b.opts.Transport = tr

	return b
}

func (b *publishBuilder) Execute() (*PublishResponse, StatusResponse, error) {
	rawJson, status, err := executeRequest(b.opts)
	if err != nil {
		return emptyPublishResponse, status, err
	}

	return newPublishResponse(rawJson, status)
}

func (o *publishOpts) config() Config {
	return *o.pubnub.Config
}

func (o *publishOpts) client() *http.Client {
	return o.pubnub.GetClient()
}

func (o *publishOpts) context() Context {
	return o.ctx
}

func (o *publishOpts) validate() error {
	if o.config().PublishKey == "" {
		return newValidationError(o, StrMissingPubKey)
	}

	if o.config().SubscribeKey == "" {
		return newValidationError(o, StrMissingSubKey)
	}

	if o.Channel == "" {
		return newValidationError(o, StrMissingChannel)
	}

	if o.Message == nil {
		return newValidationError(o, StrMissingMessage)
	}

	return nil
}

func (o *publishOpts) encryptProcessing(cipherKey string) (string, error) {
	var msg string
	var errJsonMarshal error

	o.pubnub.Config.Log.Println("EncryptString: encrypting", fmt.Sprintf("%s", o.Message))
	if o.pubnub.Config.DisablePNOtherProcessing {
		if msg, errJsonMarshal = utils.SerializeEncryptAndSerialize(o.Message, cipherKey, o.Serialize); errJsonMarshal != nil {
			o.pubnub.Config.Log.Println("error in serializing: %s", errJsonMarshal)
			return "", errJsonMarshal
		}
	} else {
		//encrypt pn_other only
		o.pubnub.Config.Log.Println("encrypt pn_other only", "reflect.TypeOf(data).Kind()", reflect.TypeOf(o.Message).Kind(), o.Message)
		switch v := o.Message.(type) {
		case map[string]interface{}:

			msgPart, ok := v["pn_other"].(string)

			if ok {
				o.pubnub.Config.Log.Println(ok, msgPart)
				if encMsg, errJsonMarshal := utils.SerializeAndEncrypt(msgPart, cipherKey, o.Serialize); errJsonMarshal != nil {
					o.pubnub.Config.Log.Println("error in serializing: %s", errJsonMarshal)
					return "", errJsonMarshal
				} else {
					v["pn_other"] = encMsg
				}

				/*jsonSerialized, errJsonMarshal := json.Marshal(msgPart)
				if errJsonMarshal != nil {
					o.pubnub.Config.Log.Println("error in serializing: %s", errJsonMarshal)
					return "", errJsonMarshal
				}

				encMsg := utils.EncryptString(cipherKey, string(jsonSerialized))*/

				/*jsonSerialized, errJsonMarshal = json.Marshal(encMsg)
				if errJsonMarshal != nil {
					o.pubnub.Config.Log.Println("error in serializing: %s", errJsonMarshal)
					return "", errJsonMarshal
				}*/
				//string(jsonSerialized)
				//msg = v

				jsonEncBytes, errEnc := json.Marshal(v)
				if errEnc != nil {
					o.pubnub.Config.Log.Println("ERROR: Publish error: %s", errEnc.Error())
					return "", errEnc
				}
				msg = string(jsonEncBytes)
			} else {
				if msg, errJsonMarshal = utils.SerializeEncryptAndSerialize(o.Message, cipherKey, o.Serialize); errJsonMarshal != nil {
					o.pubnub.Config.Log.Println("error in serializing: %s", errJsonMarshal)
					return "", errJsonMarshal
				}

				/*jsonSerialized, errJsonMarshal := json.Marshal(o.Message)
				if errJsonMarshal != nil {
					o.pubnub.Config.Log.Println("error in serializing: %s", errJsonMarshal)
					return "", errJsonMarshal
				}

				o.pubnub.Config.Log.Println(ok, jsonSerialized)
				msg = utils.EncryptString(cipherKey, string(jsonSerialized))*/
			}
			break
		default:
			if msg, errJsonMarshal = utils.SerializeEncryptAndSerialize(o.Message, cipherKey, o.Serialize); errJsonMarshal != nil {
				o.pubnub.Config.Log.Println("error in serializing: %s", errJsonMarshal)
				return "", errJsonMarshal
			}

			/*jsonSerialized, errJsonMarshal := json.Marshal(o.Message)
			if errJsonMarshal != nil {
				o.pubnub.Config.Log.Println("error in serializing: %s", errJsonMarshal)
				return "", errJsonMarshal
			}

			o.pubnub.Config.Log.Println("default", jsonSerialized)
			//msg = utils.EncryptString(cipherKey, fmt.Sprintf("%s", o.Message))
			msg = utils.EncryptString(cipherKey, string(jsonSerialized))*/
			break
		}
	}
	return msg, nil
}

//TODO Refactor
func (o *publishOpts) buildPath() (string, error) {
	if o.UsePost == true {
		return fmt.Sprintf(PUBLISH_POST_PATH,
			o.pubnub.Config.PublishKey,
			o.pubnub.Config.SubscribeKey,
			utils.UrlEncode(o.Channel),
			"0"), nil
	}

	var msg string
	var errJsonMarshal error

	if cipherKey := o.pubnub.Config.CipherKey; cipherKey != "" {
		if msg, errJsonMarshal = o.encryptProcessing(cipherKey); errJsonMarshal != nil {
			return "", errJsonMarshal
		}

		o.pubnub.Config.Log.Println("EncryptString: encrypted", msg)
	} else {
		if o.Serialize {
			jsonEncBytes, errEnc := json.Marshal(o.Message)
			if errEnc != nil {
				o.pubnub.Config.Log.Println("ERROR: Publish error: %s", errEnc.Error())
				return "", errEnc
			}
			msg = string(jsonEncBytes)
		} else {
			if serializedMsg, ok := o.Message.(string); ok {
				msg = serializedMsg
			} else {
				return "", pnerr.NewBuildRequestError("buildpath: Message is not JSON serialized.")
			}
		}
	}

	/*encodedPath := utils.EncodeJSONAsPathComponent(msg)
	fmt.Println("encodedPath: ", encodedPath, utils.UrlEncode(msg))
	o.pubnub.Config.Log.Println("encodedPath: ", encodedPath)*/
	/*message, err := utils.ValueAsString(msg)
	if err != nil {
		o.pubnub.Config.Log.Println("ERROR: Publish error: %s", err.Error())
		return "", err
	}*/

	return fmt.Sprintf(PUBLISH_GET_PATH,
		o.pubnub.Config.PublishKey,
		o.pubnub.Config.SubscribeKey,
		utils.UrlEncode(o.Channel),
		"0",
		utils.UrlEncode(msg)), nil
}

func (o *publishOpts) buildQuery() (*url.Values, error) {
	q := defaultQuery(o.pubnub.Config.Uuid, o.pubnub.telemetryManager)

	if o.Meta != nil {
		meta, err := utils.ValueAsString(o.Meta)
		if err != nil {
			return &url.Values{}, err
		}

		q.Set("meta", string(meta))
	}

	if o.setShouldStore {
		if o.ShouldStore {
			q.Set("store", "1")
		} else {
			q.Set("store", "0")
		}
	}

	if o.setTtl {
		if o.Ttl > 0 {
			q.Set("ttl", strconv.Itoa(o.Ttl))
		}
	}

	seqn := strconv.Itoa(o.pubnub.getPublishSequence())
	o.pubnub.Config.Log.Println("seqn:", seqn)
	q.Set("seqn", seqn)

	if o.DoNotReplicate == true {
		q.Set("norep", "true")
	}

	return q, nil
}

func (o *publishOpts) buildBody() ([]byte, error) {
	if o.UsePost {
		/*var msg []byte

		if o.Serialize {
			m, err := utils.ValueAsString(o.Message)
			if err != nil {
				return []byte{}, err
			}
			msg = []byte(m)
		} else {
			if s, ok := o.Message.(string); ok {
				msg = []byte(s)
			} else {
				err := pnerr.NewBuildRequestError("Type error, only string is expected")
				return []byte{}, err
			}
		}*/

		if cipherKey := o.pubnub.Config.CipherKey; cipherKey != "" {
			if msg, errJsonMarshal := o.encryptProcessing(cipherKey); errJsonMarshal != nil {
				return []byte{}, errJsonMarshal
			} else {
				return []byte(msg), nil
			}

			/*if msg, errJsonMarshal := utils.SerializeEncryptAndSerialize(o.Message, cipherKey, o.Serialize); errJsonMarshal != nil {
				o.pubnub.Config.Log.Println("error in serializing: %s", errJsonMarshal)
				return []byte{}, errJsonMarshal
			} else {
				return []byte(msg), nil
			}*/

			//enc := utils.EncryptString(cipherKey, string(msg))
			/*msg, err := utils.ValueAsString(enc)
			if err != nil {
				return []byte{}, err
			}*/

		} else {
			if o.Serialize {
				jsonEncBytes, errEnc := json.Marshal(o.Message)
				if errEnc != nil {
					o.pubnub.Config.Log.Println("ERROR: Publish error: %s", errEnc.Error())
					return []byte{}, errEnc
				}
				return jsonEncBytes, nil
			} else {
				if serializedMsg, ok := o.Message.(string); ok {
					return []byte(serializedMsg), nil
				} else {
					return []byte{}, pnerr.NewBuildRequestError("buildBody: Message is not JSON serialized.")
				}

			}

			/*jsonEncBytes, errEnc := json.Marshal(o.Message)
			if errEnc != nil {
				o.pubnub.Config.Log.Println("ERROR: Publish error: %s", errEnc.Error())
				return , errEnc
			}

			return jsonEncBytes, nil*/
		}
	} else {
		return []byte{}, nil
	}
}

func (o *publishOpts) httpMethod() string {
	if o.UsePost {
		return "POST"
	} else {
		return "GET"
	}
}

func (o *publishOpts) isAuthRequired() bool {
	return true
}

func (o *publishOpts) requestTimeout() int {
	return o.pubnub.Config.NonSubscribeRequestTimeout
}

func (o *publishOpts) connectTimeout() int {
	return o.pubnub.Config.ConnectTimeout
}

func (o *publishOpts) operationType() OperationType {
	return PNPublishOperation
}

func (o *publishOpts) telemetryManager() *TelemetryManager {
	return o.pubnub.telemetryManager
}
