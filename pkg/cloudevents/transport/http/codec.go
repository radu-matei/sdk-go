package http

import (
	"fmt"
	"github.com/cloudevents/sdk-go/pkg/cloudevents"
	"github.com/cloudevents/sdk-go/pkg/cloudevents/datacodec"
	"github.com/cloudevents/sdk-go/pkg/cloudevents/transport"
)

type Codec struct {
	Encoding Encoding

	v01 *CodecV01
	v02 *CodecV02
	v03 *CodecV03
}

var _ transport.Codec = (*Codec)(nil)

func (c *Codec) Encode(e cloudevents.Event) (transport.Message, error) {
	switch c.Encoding {
	case Default:
		fallthrough
	case BinaryV01:
		fallthrough
	case StructuredV01:
		if c.v01 == nil {
			c.v01 = &CodecV01{Encoding: c.Encoding}
		}
		return c.v01.Encode(e)
	case BinaryV02:
		fallthrough
	case StructuredV02:
		if c.v02 == nil {
			c.v02 = &CodecV02{Encoding: c.Encoding}
		}
		return c.v02.Encode(e)
	case BinaryV03:
		fallthrough
	case StructuredV03:
		if c.v03 == nil {
			c.v03 = &CodecV03{Encoding: c.Encoding}
		}
		return c.v03.Encode(e)
	default:
		return nil, fmt.Errorf("unknown encoding: %s", c.Encoding)
	}
}

func (c *Codec) Decode(msg transport.Message) (*cloudevents.Event, error) {
	switch c.inspectEncoding(msg) {
	case BinaryV01:
		fallthrough
	case StructuredV01:
		if c.v01 == nil {
			c.v01 = &CodecV01{Encoding: c.Encoding}
		}
		if event, err := c.v01.Decode(msg); err != nil {
			return nil, err
		} else {
			return c.convertEvent(event), nil
		}
	case BinaryV02:
		fallthrough
	case StructuredV02:
		if c.v02 == nil {
			c.v02 = &CodecV02{Encoding: c.Encoding}
		}
		if event, err := c.v02.Decode(msg); err != nil {
			return nil, err
		} else {
			return c.convertEvent(event), nil
		}
	case BinaryV03:
		fallthrough
	case StructuredV03:
		fallthrough
	case BatchedV03:
		if c.v03 == nil {
			c.v03 = &CodecV03{Encoding: c.Encoding}
		}
		if event, err := c.v03.Decode(msg); err != nil {
			return nil, err
		} else {
			return c.convertEvent(event), nil
		}
	default:
		return nil, fmt.Errorf("unknown encoding for message %v", msg)
	}
}

// Give the context back as the user expects
func (c *Codec) convertEvent(event *cloudevents.Event) *cloudevents.Event {
	if event == nil {
		return nil
	}
	switch c.Encoding {
	case Default:
		return event
	case BinaryV01:
		fallthrough
	case StructuredV01:
		if c.v01 == nil {
			c.v01 = &CodecV01{Encoding: c.Encoding}
		}
		ctx := event.Context.AsV01()
		event.Context = ctx
		return event
	case BinaryV02:
		fallthrough
	case StructuredV02:
		if c.v02 == nil {
			c.v02 = &CodecV02{Encoding: c.Encoding}
		}
		ctx := event.Context.AsV02()
		event.Context = ctx
		return event
	case BinaryV03:
		fallthrough
	case StructuredV03:
		fallthrough
	case BatchedV03:
		if c.v03 == nil {
			c.v03 = &CodecV03{Encoding: c.Encoding}
		}
		ctx := event.Context.AsV03()
		event.Context = ctx
		return event
	default:
		return nil
	}
}

func (c *Codec) inspectEncoding(msg transport.Message) Encoding {
	// TODO: there should be a better way to make the version codecs on demand.
	if c.v01 == nil {
		c.v01 = &CodecV01{Encoding: c.Encoding}
	}
	// Try v0.1 first.
	encoding := c.v01.inspectEncoding(msg)
	if encoding != Unknown {
		return encoding
	}

	if c.v02 == nil {
		c.v02 = &CodecV02{Encoding: c.Encoding}
	}
	// Try v0.2.
	encoding = c.v02.inspectEncoding(msg)
	if encoding != Unknown {
		return encoding
	}

	if c.v03 == nil {
		c.v03 = &CodecV03{Encoding: c.Encoding}
	}
	// Try v0.3.
	encoding = c.v03.inspectEncoding(msg)
	if encoding != Unknown {
		return encoding
	}

	// We do not understand the message encoding.
	return Unknown
}

// ---------
// TODO: Should move these somewhere else. the methods are shared for all versions.

func marshalEventData(encoding string, data interface{}) ([]byte, error) {
	if data == nil {
		return []byte(nil), nil
	}
	// already encoded?
	if b, ok := data.([]byte); ok {
		return b, nil
	}

	return datacodec.Encode(encoding, data)
}
