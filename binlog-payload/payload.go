package blp

type CanalPayload struct {
	Id        int64    `json:"id"`
	Db        string   `json:"database"`
	Table     string   `json:"table"`
	PKColumn  []string `json:"pkNames"`
	IsDdl     bool     `json:"isDdl"`
	EventType string   `json:"type"`
	Es        int64    `json:"es"`
	Ts        int64    `json:"ts"`
	Sql       string   `json:"sql"`
	// MysqlType map[string]string        `json:"mysqlType"`
	// SqlType   map[string]int16         `json:"sqlType"`
	Rows []map[string]interface{} `json:"data"`
	Olds []map[string]interface{} `json:"old"`
}
