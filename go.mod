module code.superseriousbusiness.org/gotosocial

go 1.24

toolchain go1.24.3

// Replace go-swagger with our version that fixes (ours particularly) use of Go1.23
replace github.com/go-swagger/go-swagger => codeberg.org/superseriousbusiness/go-swagger v0.31.0-gts-go1.23-fix

// Replace modernc/sqlite with our version that fixes the concurrency INTERRUPT issue
replace modernc.org/sqlite => gitlab.com/NyaaaWhatsUpDoc/sqlite v1.37.1-concurrency-workaround

require (
	code.superseriousbusiness.org/activity v1.15.0
	code.superseriousbusiness.org/exif-terminator v0.11.0
	code.superseriousbusiness.org/httpsig v1.4.0
	code.superseriousbusiness.org/oauth2/v4 v4.5.4-0.20250606121655-9d54ef189d42
	codeberg.org/gruf/go-bitutil v1.1.0
	codeberg.org/gruf/go-bytesize v1.0.3
	codeberg.org/gruf/go-byteutil v1.3.0
	codeberg.org/gruf/go-cache/v3 v3.6.1
	codeberg.org/gruf/go-debug v1.3.0
	codeberg.org/gruf/go-errors/v2 v2.3.2
	codeberg.org/gruf/go-fastcopy v1.1.3
	codeberg.org/gruf/go-ffmpreg v0.6.7
	codeberg.org/gruf/go-iotools v0.0.0-20240710125620-934ae9c654cf
	codeberg.org/gruf/go-kv v1.6.5
	codeberg.org/gruf/go-list v0.0.0-20240425093752-494db03d641f
	codeberg.org/gruf/go-mempool v0.0.0-20240507125005-cef10d64a760
	codeberg.org/gruf/go-mutexes v1.5.2
	codeberg.org/gruf/go-runners v1.6.3
	codeberg.org/gruf/go-sched v1.2.4
	codeberg.org/gruf/go-split v1.2.0
	codeberg.org/gruf/go-storage v0.3.1
	codeberg.org/gruf/go-structr v0.9.7
	github.com/DmitriyVTitov/size v1.5.0
	github.com/KimMachineGun/automemlimit v0.7.2
	github.com/SherClockHolmes/webpush-go v1.4.0
	github.com/buckket/go-blurhash v1.1.0
	github.com/coreos/go-oidc/v3 v3.14.1
	github.com/gin-contrib/cors v1.7.5
	github.com/gin-contrib/gzip v1.2.3
	github.com/gin-contrib/sessions v1.0.4
	github.com/gin-gonic/gin v1.10.1
	github.com/go-playground/form/v4 v4.2.1
	github.com/go-swagger/go-swagger v0.31.0
	github.com/google/go-cmp v0.7.0
	github.com/google/uuid v1.6.0
	github.com/gorilla/feeds v1.2.0
	github.com/gorilla/websocket v1.5.3
	github.com/jackc/pgx/v5 v5.7.5
	github.com/k3a/html2text v1.2.1
	github.com/microcosm-cc/bluemonday v1.0.27
	github.com/miekg/dns v1.1.66
	github.com/minio/minio-go/v7 v7.0.92
	github.com/mitchellh/mapstructure v1.5.0
	github.com/ncruces/go-sqlite3 v0.26.0
	github.com/oklog/ulid v1.3.1
	github.com/pquerna/otp v1.5.0
	github.com/rivo/uniseg v0.4.7
	github.com/spf13/cast v1.8.0
	github.com/spf13/cobra v1.9.1
	github.com/spf13/pflag v1.0.6
	github.com/spf13/viper v1.20.1
	github.com/stretchr/testify v1.10.0
	github.com/tdewolff/minify/v2 v2.23.8
	github.com/technologize/otel-go-contrib v1.1.1
	github.com/temoto/robotstxt v1.1.2
	github.com/tetratelabs/wazero v1.9.0
	github.com/tomnomnom/linkheader v0.0.0-20180905144013-02ca5825eb80
	github.com/ulule/limiter/v3 v3.11.2
	github.com/uptrace/bun v1.2.11
	github.com/uptrace/bun/dialect/pgdialect v1.2.11
	github.com/uptrace/bun/dialect/sqlitedialect v1.2.11
	github.com/uptrace/bun/extra/bunotel v1.2.11
	github.com/wagslane/go-password-validator v0.3.0
	github.com/yuin/goldmark v1.7.12
	go.opentelemetry.io/contrib/exporters/autoexport v0.61.0
	go.opentelemetry.io/contrib/instrumentation/runtime v0.61.0
	go.opentelemetry.io/otel v1.36.0
	go.opentelemetry.io/otel/metric v1.36.0
	go.opentelemetry.io/otel/sdk v1.36.0
	go.opentelemetry.io/otel/sdk/metric v1.36.0
	go.opentelemetry.io/otel/trace v1.36.0
	go.uber.org/automaxprocs v1.6.0
	golang.org/x/crypto v0.38.0
	golang.org/x/image v0.27.0
	golang.org/x/net v0.40.0
	golang.org/x/oauth2 v0.30.0
	golang.org/x/sys v0.33.0
	golang.org/x/text v0.25.0
	gopkg.in/mcuadros/go-syslog.v2 v2.3.0
	gopkg.in/yaml.v3 v3.0.1
	modernc.org/sqlite v1.37.1
	mvdan.cc/xurls/v2 v2.6.0
)

require (
	code.superseriousbusiness.org/go-jpeg-image-structure/v2 v2.3.0 // indirect
	code.superseriousbusiness.org/go-png-image-structure/v2 v2.3.0 // indirect
	codeberg.org/gruf/go-fastpath/v2 v2.0.0 // indirect
	codeberg.org/gruf/go-mangler v1.4.4 // indirect
	codeberg.org/gruf/go-maps v1.0.4 // indirect
	github.com/Masterminds/goutils v1.1.1 // indirect
	github.com/Masterminds/semver/v3 v3.2.1 // indirect
	github.com/Masterminds/sprig/v3 v3.2.3 // indirect
	github.com/asaskevich/govalidator v0.0.0-20230301143203-a9d515a09cc2 // indirect
	github.com/aymerick/douceur v0.2.0 // indirect
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/boombuler/barcode v1.0.1-0.20190219062509-6c824513bacc // indirect
	github.com/bytedance/sonic v1.13.2 // indirect
	github.com/bytedance/sonic/loader v0.2.4 // indirect
	github.com/cenkalti/backoff/v5 v5.0.2 // indirect
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/cloudwego/base64x v0.1.5 // indirect
	github.com/davecgh/go-spew v1.1.2-0.20180830191138-d8f796af33cc // indirect
	github.com/dsoprea/go-exif/v3 v3.0.0-20210625224831-a6301f85c82b // indirect
	github.com/dsoprea/go-iptc v0.0.0-20200609062250-162ae6b44feb // indirect
	github.com/dsoprea/go-logging v0.0.0-20200710184922-b02d349568dd // indirect
	github.com/dsoprea/go-photoshop-info-format v0.0.0-20200609050348-3db9b63b202c // indirect
	github.com/dsoprea/go-utility/v2 v2.0.0-20200717064901-2fccff4aa15e // indirect
	github.com/dustin/go-humanize v1.0.1 // indirect
	github.com/felixge/httpsnoop v1.0.4 // indirect
	github.com/fsnotify/fsnotify v1.8.0 // indirect
	github.com/gabriel-vasile/mimetype v1.4.8 // indirect
	github.com/gin-contrib/sse v1.0.0 // indirect
	github.com/go-errors/errors v1.1.1 // indirect
	github.com/go-ini/ini v1.67.0 // indirect
	github.com/go-jose/go-jose/v4 v4.0.5 // indirect
	github.com/go-logr/logr v1.4.2 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/go-openapi/analysis v0.23.0 // indirect
	github.com/go-openapi/errors v0.22.0 // indirect
	github.com/go-openapi/inflect v0.21.0 // indirect
	github.com/go-openapi/jsonpointer v0.21.0 // indirect
	github.com/go-openapi/jsonreference v0.21.0 // indirect
	github.com/go-openapi/loads v0.22.0 // indirect
	github.com/go-openapi/runtime v0.28.0 // indirect
	github.com/go-openapi/spec v0.21.0 // indirect
	github.com/go-openapi/strfmt v0.23.0 // indirect
	github.com/go-openapi/swag v0.23.0 // indirect
	github.com/go-openapi/validate v0.24.0 // indirect
	github.com/go-playground/locales v0.14.1 // indirect
	github.com/go-playground/universal-translator v0.18.1 // indirect
	github.com/go-playground/validator/v10 v10.26.0 // indirect
	github.com/go-viper/mapstructure/v2 v2.2.1 // indirect
	github.com/go-xmlfmt/xmlfmt v0.0.0-20191208150333-d5b6f63a941b // indirect
	github.com/goccy/go-json v0.10.5 // indirect
	github.com/golang-jwt/jwt/v5 v5.2.2 // indirect
	github.com/golang/geo v0.0.0-20200319012246-673a6f80352d // indirect
	github.com/gorilla/context v1.1.2 // indirect
	github.com/gorilla/css v1.0.1 // indirect
	github.com/gorilla/handlers v1.5.2 // indirect
	github.com/gorilla/securecookie v1.1.2 // indirect
	github.com/gorilla/sessions v1.4.0 // indirect
	github.com/grpc-ecosystem/grpc-gateway/v2 v2.26.3 // indirect
	github.com/huandu/xstrings v1.4.0 // indirect
	github.com/imdario/mergo v0.3.16 // indirect
	github.com/inconshreveable/mousetrap v1.1.0 // indirect
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgservicefile v0.0.0-20240606120523-5a60cdf6a761 // indirect
	github.com/jackc/puddle/v2 v2.2.2 // indirect
	github.com/jessevdk/go-flags v1.5.0 // indirect
	github.com/jinzhu/inflection v1.0.0 // indirect
	github.com/josharian/intern v1.0.0 // indirect
	github.com/json-iterator/go v1.1.12 // indirect
	github.com/klauspost/compress v1.18.0 // indirect
	github.com/klauspost/cpuid/v2 v2.2.10 // indirect
	github.com/kr/pretty v0.3.1 // indirect
	github.com/kr/text v0.2.0 // indirect
	github.com/leodido/go-urn v1.4.0 // indirect
	github.com/mailru/easyjson v0.7.7 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	github.com/minio/crc64nvme v1.0.1 // indirect
	github.com/minio/md5-simd v1.1.2 // indirect
	github.com/mitchellh/copystructure v1.2.0 // indirect
	github.com/mitchellh/reflectwalk v1.0.2 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.2 // indirect
	github.com/munnerz/goautoneg v0.0.0-20191010083416-a7dc8b61c822 // indirect
	github.com/ncruces/go-strftime v0.1.9 // indirect
	github.com/ncruces/julianday v1.0.0 // indirect
	github.com/pbnjay/memory v0.0.0-20210728143218-7b4eea64cf58 // indirect
	github.com/pelletier/go-toml/v2 v2.2.3 // indirect
	github.com/philhofer/fwd v1.1.3-0.20240916144458-20a13a1f6b7c // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/pmezard/go-difflib v1.0.1-0.20181226105442-5d4384ee4fb2 // indirect
	github.com/prometheus/client_golang v1.22.0 // indirect
	github.com/prometheus/client_model v0.6.2 // indirect
	github.com/prometheus/common v0.64.0 // indirect
	github.com/prometheus/procfs v0.16.1 // indirect
	github.com/puzpuzpuz/xsync/v3 v3.5.1 // indirect
	github.com/quasoft/memstore v0.0.0-20191010062613-2bce066d2b0b // indirect
	github.com/remyoudompheng/bigfft v0.0.0-20230129092748-24d4a6f8daec // indirect
	github.com/rogpeppe/go-internal v1.13.2-0.20241226121412-a5dc8ff20d0a // indirect
	github.com/rs/xid v1.6.0 // indirect
	github.com/sagikazarmark/locafero v0.7.0 // indirect
	github.com/shopspring/decimal v1.3.1 // indirect
	github.com/sourcegraph/conc v0.3.0 // indirect
	github.com/spf13/afero v1.12.0 // indirect
	github.com/subosito/gotenv v1.6.0 // indirect
	github.com/tdewolff/parse/v2 v2.8.1 // indirect
	github.com/tinylib/msgp v1.3.0 // indirect
	github.com/tmthrgd/go-hex v0.0.0-20190904060850-447a3041c3bc // indirect
	github.com/toqueteos/webbrowser v1.2.0 // indirect
	github.com/twitchyliquid64/golang-asm v0.15.1 // indirect
	github.com/ugorji/go/codec v1.2.12 // indirect
	github.com/uptrace/opentelemetry-go-extra/otelsql v0.3.2 // indirect
	github.com/vmihailenco/msgpack/v5 v5.4.1 // indirect
	github.com/vmihailenco/tagparser/v2 v2.0.0 // indirect
	go.mongodb.org/mongo-driver v1.17.3 // indirect
	go.opentelemetry.io/auto/sdk v1.1.0 // indirect
	go.opentelemetry.io/contrib/bridges/prometheus v0.61.0 // indirect
	go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploggrpc v0.12.2 // indirect
	go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploghttp v0.12.2 // indirect
	go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc v1.36.0 // indirect
	go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp v1.36.0 // indirect
	go.opentelemetry.io/otel/exporters/otlp/otlptrace v1.36.0 // indirect
	go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc v1.36.0 // indirect
	go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp v1.36.0 // indirect
	go.opentelemetry.io/otel/exporters/prometheus v0.58.0 // indirect
	go.opentelemetry.io/otel/exporters/stdout/stdoutlog v0.12.2 // indirect
	go.opentelemetry.io/otel/exporters/stdout/stdoutmetric v1.36.0 // indirect
	go.opentelemetry.io/otel/exporters/stdout/stdouttrace v1.36.0 // indirect
	go.opentelemetry.io/otel/log v0.12.2 // indirect
	go.opentelemetry.io/otel/sdk/log v0.12.2 // indirect
	go.opentelemetry.io/proto/otlp v1.6.0 // indirect
	go.uber.org/multierr v1.11.0 // indirect
	golang.org/x/arch v0.16.0 // indirect
	golang.org/x/exp v0.0.0-20250408133849-7e4ce0ab07d0 // indirect
	golang.org/x/mod v0.24.0 // indirect
	golang.org/x/sync v0.14.0 // indirect
	golang.org/x/tools v0.33.0 // indirect
	google.golang.org/genproto/googleapis/api v0.0.0-20250519155744-55703ea1f237 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20250519155744-55703ea1f237 // indirect
	google.golang.org/grpc v1.72.1 // indirect
	google.golang.org/protobuf v1.36.6 // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
	modernc.org/libc v1.65.7 // indirect
	modernc.org/mathutil v1.7.1 // indirect
	modernc.org/memory v1.11.0 // indirect
)
