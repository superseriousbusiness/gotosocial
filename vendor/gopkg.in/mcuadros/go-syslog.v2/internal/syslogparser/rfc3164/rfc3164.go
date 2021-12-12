package rfc3164

import (
	"bytes"
	"os"
	"time"

	"gopkg.in/mcuadros/go-syslog.v2/internal/syslogparser"
)

type Parser struct {
	buff     []byte
	cursor   int
	l        int
	priority syslogparser.Priority
	version  int
	header   header
	message  rfc3164message
	location *time.Location
	skipTag  bool
}

type header struct {
	timestamp time.Time
	hostname  string
}

type rfc3164message struct {
	tag     string
	content string
}

func NewParser(buff []byte) *Parser {
	return &Parser{
		buff:     buff,
		cursor:   0,
		l:        len(buff),
		location: time.UTC,
	}
}

func (p *Parser) Location(location *time.Location) {
	p.location = location
}

func (p *Parser) Parse() error {
	tcursor := p.cursor
	pri, err := p.parsePriority()
	if err != nil {
		// RFC3164 sec 4.3.3
		p.priority = syslogparser.Priority{13, syslogparser.Facility{Value: 1}, syslogparser.Severity{Value: 5}}
		p.cursor = tcursor
		content, err := p.parseContent()
		p.header.timestamp = time.Now().Round(time.Second)
		if err != syslogparser.ErrEOL {
			return err
		}
		p.message = rfc3164message{content: content}
		return nil
	}

	tcursor = p.cursor
	hdr, err := p.parseHeader()
	if err == syslogparser.ErrTimestampUnknownFormat {
		// RFC3164 sec 4.3.2.
		hdr.timestamp = time.Now().Round(time.Second)
		// No tag processing should be done
		p.skipTag = true
		// Reset cursor for content read
		p.cursor = tcursor
	} else if err != nil {
		return err
	} else {
		p.cursor++
	}

	msg, err := p.parsemessage()
	if err != syslogparser.ErrEOL {
		return err
	}

	p.priority = pri
	p.version = syslogparser.NO_VERSION
	p.header = hdr
	p.message = msg

	return nil
}

func (p *Parser) Dump() syslogparser.LogParts {
	return syslogparser.LogParts{
		"timestamp": p.header.timestamp,
		"hostname":  p.header.hostname,
		"tag":       p.message.tag,
		"content":   p.message.content,
		"priority":  p.priority.P,
		"facility":  p.priority.F.Value,
		"severity":  p.priority.S.Value,
	}
}

func (p *Parser) parsePriority() (syslogparser.Priority, error) {
	return syslogparser.ParsePriority(p.buff, &p.cursor, p.l)
}

func (p *Parser) parseHeader() (header, error) {
	hdr := header{}
	var err error

	ts, err := p.parseTimestamp()
	if err != nil {
		return hdr, err
	}

	hostname, err := p.parseHostname()
	if err != nil {
		return hdr, err
	}

	hdr.timestamp = ts
	hdr.hostname = hostname

	return hdr, nil
}

func (p *Parser) parsemessage() (rfc3164message, error) {
	msg := rfc3164message{}
	var err error

	if !p.skipTag {
		tag, err := p.parseTag()
		if err != nil {
			return msg, err
		}
		msg.tag = tag
	}

	content, err := p.parseContent()
	if err != syslogparser.ErrEOL {
		return msg, err
	}

	msg.content = content

	return msg, err
}

// https://tools.ietf.org/html/rfc3164#section-4.1.2
func (p *Parser) parseTimestamp() (time.Time, error) {
	var ts time.Time
	var err error
	var tsFmtLen int
	var sub []byte

	tsFmts := []string{
		time.Stamp,
		time.RFC3339,
	}
	// if timestamps starts with numeric try formats with different order
	// it is more likely that timestamp is in RFC3339 format then
	if c := p.buff[p.cursor]; c > '0' && c < '9' {
		tsFmts = []string{
			time.RFC3339,
			time.Stamp,
		}
	}

	found := false
	for _, tsFmt := range tsFmts {
		tsFmtLen = len(tsFmt)

		if p.cursor+tsFmtLen > p.l {
			continue
		}

		sub = p.buff[p.cursor : tsFmtLen+p.cursor]
		ts, err = time.ParseInLocation(tsFmt, string(sub), p.location)
		if err == nil {
			found = true
			break
		}
	}

	if !found {
		p.cursor = len(time.Stamp)

		// XXX : If the timestamp is invalid we try to push the cursor one byte
		// XXX : further, in case it is a space
		if (p.cursor < p.l) && (p.buff[p.cursor] == ' ') {
			p.cursor++
		}

		return ts, syslogparser.ErrTimestampUnknownFormat
	}

	fixTimestampIfNeeded(&ts)

	p.cursor += tsFmtLen

	if (p.cursor < p.l) && (p.buff[p.cursor] == ' ') {
		p.cursor++
	}

	return ts, nil
}

func (p *Parser) parseHostname() (string, error) {
	oldcursor := p.cursor
	hostname, err := syslogparser.ParseHostname(p.buff, &p.cursor, p.l)
	if err == nil && len(hostname) > 0 && string(hostname[len(hostname)-1]) == ":" { // not an hostname! we found a GNU implementation of syslog()
		p.cursor = oldcursor - 1
		myhostname, err := os.Hostname()
		if err == nil {
			return myhostname, nil
		}
		return "", nil
	}
	return hostname, err
}

// http://tools.ietf.org/html/rfc3164#section-4.1.3
func (p *Parser) parseTag() (string, error) {
	var b byte
	var endOfTag bool
	var bracketOpen bool
	var tag []byte
	var err error
	var found bool

	from := p.cursor

	for {
		if p.cursor == p.l {
			// no tag found, reset cursor for content
			p.cursor = from
			return "", nil
		}

		b = p.buff[p.cursor]
		bracketOpen = (b == '[')
		endOfTag = (b == ':' || b == ' ')

		// XXX : parse PID ?
		if bracketOpen {
			tag = p.buff[from:p.cursor]
			found = true
		}

		if endOfTag {
			if !found {
				tag = p.buff[from:p.cursor]
				found = true
			}

			p.cursor++
			break
		}

		p.cursor++
	}

	if (p.cursor < p.l) && (p.buff[p.cursor] == ' ') {
		p.cursor++
	}

	return string(tag), err
}

func (p *Parser) parseContent() (string, error) {
	if p.cursor > p.l {
		return "", syslogparser.ErrEOL
	}

	content := bytes.Trim(p.buff[p.cursor:p.l], " ")
	p.cursor += len(content)

	return string(content), syslogparser.ErrEOL
}

func fixTimestampIfNeeded(ts *time.Time) {
	now := time.Now()
	y := ts.Year()

	if ts.Year() == 0 {
		y = now.Year()
	}

	newTs := time.Date(y, ts.Month(), ts.Day(), ts.Hour(), ts.Minute(),
		ts.Second(), ts.Nanosecond(), ts.Location())

	*ts = newTs
}
