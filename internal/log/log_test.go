package log_test

import (
	"bytes"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/superseriousbusiness/gotosocial/internal/log"
)

func TestOutputSplitFunc(t *testing.T) {
	var outbuf, errbuf bytes.Buffer

	out := log.SplitErrOutputs(&outbuf, &errbuf)

	log := logrus.New()
	log.SetOutput(out)
	log.SetLevel(logrus.TraceLevel)

	for _, lvl := range logrus.AllLevels {
		func() {
			defer func() { recover() }()
			log.Log(lvl, "hello world")
		}()

		t.Logf("outbuf=%q errbuf=%q", outbuf.String(), errbuf.String())

		switch lvl {
		case logrus.PanicLevel:
			if outbuf.Len() > 0 || errbuf.Len() == 0 {
				t.Error("expected panic to log to OutputSplitter.Err")
			}
		case logrus.FatalLevel:
			if outbuf.Len() > 0 || errbuf.Len() == 0 {
				t.Error("expected fatal to log to OutputSplitter.Err")
			}
		case logrus.ErrorLevel:
			if outbuf.Len() > 0 || errbuf.Len() == 0 {
				t.Error("expected error to log to OutputSplitter.Err")
			}
		default:
			if outbuf.Len() == 0 || errbuf.Len() > 0 {
				t.Errorf("expected %s to log to OutputSplitter.Out", lvl)
			}
		}

		// Reset buffers
		outbuf.Reset()
		errbuf.Reset()
	}
}
