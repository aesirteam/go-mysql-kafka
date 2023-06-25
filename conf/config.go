package conf

import (
	"io/ioutil"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/go-mysql-org/go-mysql/schema"
	log "github.com/sirupsen/logrus"
)

type ConfigSet struct {
	Debug    bool     `toml:"debug"`    // 是否开启debug模式
	Env      string   `toml:"env"`      // 运行环境
	SourceDB MysqlSet `toml:"sourceDB"` // 源数据库的配置
	Mapper   Mapper   `toml:"mapper"`   // 分表分库匹配规则
	Kafka    KafkaSet `toml:"kafka"`
}

// 分表分库
type Mapper struct {
	Schemas []string `toml:"schemas"`
}

type KafkaSet struct {
	Brokers  []string         `toml:"brokers"`
	Version  string           `toml:"version"` // kafka的版本
	Producer KafkaProducerSet `toml:"producer"`

	InsecureSkipVerify bool   `toml:"insecureSkipVerify"`
	SaslEnable         bool   `toml:"saslEnable"`
	Username           string `toml:"username"`
	Password           string `toml:"password"`
	CertFile           string `toml:"certFile"`
	Mechanism          string `toml:"mechanism"`
}

type KafkaProducerSet struct {
	RequiredAcks     int                `toml:"requiredAcks"`
	ReturnSuccesses  bool               `toml:"returnSuccesses"`
	ReturnErrors     bool               `toml:"returnErrors"`
	Async            bool               `toml:"async"`
	RetryMax         int                `toml:"retryMax"`
	Headers          []KafkaHeader      `toml:"headers"`
	TableMapperTopic []KafkaMapperTopic `toml:"mapper"`
	PartitionerType  string             `toml:"partitionerType"`
}

type KafkaMapperTopic struct {
	Topic       string `toml:"topic"`
	SourceTable string `toml:"sourceTable"`
}

type KafkaHeader struct {
	Key   string `toml:"key"`
	Value string `toml:"value"`
}

type MysqlSet struct {
	Host           string        `toml:"host"`
	Port           int           `toml:"port"`
	UserName       string        `toml:"username"`
	Password       string        `toml:"password"`
	Charset        string        `toml:"charset"`
	ServerID       uint32        `toml:"serverID"`
	Flavor         string        `toml:"flavor"`    // mysql or mariadb
	DumpExec       string        `toml:"mysqldump"` // if not set or empty, ignore mysqldump.
	BulkSize       int           `toml:"bulkSize"`  // minimal items to be inserted in one bulk
	FlushBulkTime  time.Duration `toml:"flushBulkTime"`
	SkipNoPkTable  bool          `toml:"skipNoPkTable"`
	SkipMasterData bool          `toml:"skipMasterData"`
	DataDir        string        `toml:"DataDir"` // 保存binlog到本地文件这个方法没有测试过

	Sources []SourceConfig `toml:"sources"`
	//Rules  []*RuleConfig  `toml:"rule"`
}

type SourceConfig struct {
	Schema string   `toml:"schema"`
	Tables []string `toml:"tables"`
}

type RuleConfig struct {
	Schema string   `toml:"schema"`
	Table  string   `toml:"table"`
	Index  string   `toml:"index"`
	Type   string   `toml:"type"`
	Parent string   `toml:"parent"`
	ID     []string `toml:"id"`

	// Default, a MySQL table field name is mapped to Elasticsearch field name.
	// Sometimes, you want to use different name, e.g, the MySQL file name is title,
	// but in Elasticsearch, you want to name it my_title.
	FieldMapping map[string]string `toml:"field"`

	// MySQL table information
	TableInfo *schema.Table

	//only MySQL fields in filter will be synced , default sync all fields
	Filter []string `toml:"filter"`

	// Elasticsearch pipeline
	// To pre-process documents before indexing
	Pipeline string `toml:"pipeline"`
}

var Config = &ConfigSet{}

func init() {
	log.SetFormatter(&log.TextFormatter{
		DisableColors: true,
		FullTimestamp: true,
	})
}

func Setup(cfg string) {
	configPath := cfg
	data, err := ioutil.ReadFile(configPath)
	if err != nil {
		log.Fatalf("read toml config err: %+v", err)
	}

	if _, err := toml.Decode(string(data), &Config); err != nil {
		log.Fatalf("decode toml config err: %+v", err)
	}

	if Config.Debug == true {
		log.SetLevel(log.DebugLevel)
	}

	// kafka 异步投递会卡死,目前先不开放
	Config.Kafka.Producer.Async = false

	//fmt.Printf("%+v", Config)
}

// 检查配置文件
func (c *ConfigSet) checkConfig() {
	// kafka配置检查
	if len(c.Kafka.Brokers) < 1 {
		log.Fatalf("kafka brokers can not be empty, err")
	}

}
