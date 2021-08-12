module github.com/dsoprea/go-jpeg-image-structure

go 1.13

// Development only
// replace github.com/dsoprea/go-utility => ../go-utility
// replace github.com/dsoprea/go-logging => ../go-logging
// replace github.com/dsoprea/go-exif/v2 => ../go-exif/v2
// replace github.com/dsoprea/go-photoshop-info-format => ../go-photoshop-info-format
// replace github.com/dsoprea/go-iptc => ../go-iptc

require (
	github.com/dsoprea/go-exif/v2 v2.0.0-20200604193436-ca8584a0e1c4
	github.com/dsoprea/go-exif/v3 v3.0.0-20210512043655-120bcdb2a55e // indirect
	github.com/dsoprea/go-iptc v0.0.0-20200609062250-162ae6b44feb
	github.com/dsoprea/go-logging v0.0.0-20200517223158-a10564966e9d
	github.com/dsoprea/go-photoshop-info-format v0.0.0-20200609050348-3db9b63b202c
	github.com/dsoprea/go-utility v0.0.0-20200711062821-fab8125e9bdf
	github.com/go-xmlfmt/xmlfmt v0.0.0-20191208150333-d5b6f63a941b
	github.com/jessevdk/go-flags v1.4.0
	golang.org/x/net v0.0.0-20200520182314-0ba52f642ac2 // indirect
)
