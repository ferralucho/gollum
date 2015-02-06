package producer

import (
	"github.com/trivago/gollum/shared"
	"regexp"
	"strings"
)

// Producer base class
// All producers support a common subset of configuration options:
//
// - "producer.Something":
//   Enable: true
//   Buffer: 1024
//   Forward: false
//   Delimiter: "\r\n"
//   Stream:
//      - "error"
//      - "default"
//
// Enable switches the consumer on or off. By default this value is set to true.
//
// Buffer set the size of the channel used to communicate messages. By default
// this value is set to 1024.
//
// Stream contains either a single string or a list of strings defining the
// message channels this producer will consume. By default this is set to "*"
// which means "all streams".
//
// If Forward is set to true, the message will be passed as-is, so date and
// channel will not be added. The default value is false.
//
// If Delimiter is set another end-of-message delimiter will be appened to the
// end of the message. If Forward is defined and Delimiter is not defined, no
// end-of-line delimiter will be written. Besides this case delimiter is set to
// "\n" by default.
type standardProducer struct {
	messages chan shared.Message
	control  chan shared.ProducerControl
	filter   *regexp.Regexp
	format   shared.MessageFormat
}

func (prod *standardProducer) configureStandardProducer(conf shared.PluginConfig) error {
	prod.messages = make(chan shared.Message, conf.Channel)
	prod.control = make(chan shared.ProducerControl, 1)
	prod.filter = nil

	escapeChars := strings.NewReplacer("\\n", "\n", "\\r", "\r", "\\t", "\t")
	delimiter := escapeChars.Replace(conf.GetString("Delimiter", shared.DefaultDelimiter))

	if conf.GetBool("Forward", false) {
		if conf.HasValue("Delimiter") {
			prod.format = shared.CreateMessageFormatSimple(delimiter)
		} else {
			prod.format = shared.CreateMessageFormatForward()
		}
	} else {
		prod.format = shared.CreateMessageFormatTimestamp(shared.DefaultTimestamp, delimiter)
	}

	filter := conf.GetString("Filter", "")

	if filter != "" {
		var err error
		prod.filter, err = regexp.Compile(filter)
		if err != nil {
			shared.Log.Error("Regex error: ", err)
		}
	}

	return nil
}

func (prod standardProducer) Accepts(message shared.Message) bool {
	if prod.filter == nil {
		return true // ### return, pass everything ###
	}

	return prod.filter.MatchString(message.Data)
}

func (prod standardProducer) Control() chan<- shared.ProducerControl {
	return prod.control
}

func (prod standardProducer) Messages() chan<- shared.Message {
	return prod.messages
}