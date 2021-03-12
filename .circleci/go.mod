module github.com/signalfx/signalfx-go-tracing

go 1.12

require (
	github.com/DataDog/zstd v1.4.5 // indirect
	github.com/aws/aws-sdk-go v1.30.9
	github.com/bradfitz/gomemcache v0.0.0-20190913173617-a41fca850d0b
	github.com/confluentinc/confluent-kafka-go v1.4.0
	github.com/davecgh/go-spew v1.1.1
	github.com/emicklei/go-restful v2.12.0+incompatible
	github.com/garyburd/redigo v1.6.0
	github.com/gin-gonic/gin v1.6.2
	github.com/globalsign/mgo v0.0.0-20181015135952-eeefdecb41b8
	github.com/go-chi/chi v4.1.1+incompatible
	github.com/go-logfmt/logfmt v0.3.0
	github.com/go-redis/redis v6.15.7+incompatible
	github.com/go-sql-driver/mysql v1.5.0
	github.com/go-stack/stack v1.8.0
	github.com/gocql/gocql v0.0.0-20200410100145-b454769479c6
	github.com/golang/protobuf v1.4.0
	github.com/gomodule/redigo v2.0.0+incompatible
	github.com/gorilla/mux v1.7.4
	github.com/graph-gophers/graphql-go v0.0.0-20200309224638-dae41bde9ef9
	github.com/imdario/mergo v0.3.9 // indirect
	github.com/jinzhu/gorm v1.9.12
	github.com/jmoiron/sqlx v1.2.0
	github.com/julienschmidt/httprouter v1.3.0
	github.com/kr/logfmt v0.0.0-20140226030751-b84e30acd515
	github.com/labstack/echo v3.3.10+incompatible
	github.com/labstack/gommon v0.3.0 // indirect
	github.com/lib/pq v1.2.0
	github.com/mailru/easyjson v0.7.1
	github.com/mattn/go-sqlite3 v2.0.3+incompatible
	github.com/miekg/dns v1.1.29
	github.com/opentracing/opentracing-go v1.1.0
	github.com/philhofer/fwd v1.0.0
	github.com/pierrec/lz4 v2.5.1+incompatible // indirect
	github.com/pmezard/go-difflib v1.0.0
	github.com/rcrowley/go-metrics v0.0.0-20200313005456-10cdbea86bc0 // indirect
	github.com/signalfx/golib v2.4.0+incompatible
	github.com/stretchr/testify v1.5.1
	github.com/syndtr/goleveldb v1.0.0
	github.com/tebeka/go2xunit v1.4.10
	github.com/tidwall/buntdb v1.1.2
	github.com/tinylib/msgp v1.1.0
	go.mongodb.org/mongo-driver v1.3.2
	golang.org/x/net v0.0.0-20200202094626-16171245cfb2
	golang.org/x/oauth2 v0.0.0-20200107190931-bf48bf16ab8d
	golang.org/x/sys v0.0.0-20200116001909-b77594299b42
	google.golang.org/api v0.21.0
	google.golang.org/grpc v1.28.1
	gopkg.in/Shopify/sarama.v1 v1.20.1
	gopkg.in/olivere/elastic.v3 v3.0.75
	gopkg.in/olivere/elastic.v5 v5.0.85
	gopkg.in/yaml.v2 v2.2.8
	gotest.tools v2.2.0+incompatible
	k8s.io/api v0.0.0-20190819141258-3544db3b9e44
	k8s.io/apimachinery v0.0.0-20190817020851-f2f3a405f61d
	k8s.io/client-go v0.0.0-20190819141724-e14f31a72a77
	k8s.io/utils v0.0.0-20200414100711-2df71ebbae66 // indirect
)

replace (
    gopkg.in/Shopify/sarama.v1 => github.com/Shopify/sarama v1.26.1
    github.com/googleapis/gnostic => github.com/googleapis/gnostic v0.4.0
)

