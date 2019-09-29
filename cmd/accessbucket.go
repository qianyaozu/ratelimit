package main

import (
	"encoding/json"
	"errors"
	"github.com/qianyaozu/qconf"
	"io/ioutil"
	"sync"
	"time"
)

var (
	defaultInterval = 1000  //时间间隔ms
	defaultUnit     = 10000 //单位时间内输出
	defaultCap      = 10000 //桶容量
	redisAddress    = ""
	listFilePath    = `E:\go\src\github.com\qianyaozu\qratelimit\bucketList.json`
	configFilePath  = `E:\go\src\github.com\qianyaozu\qratelimit\defaultparam.ini`
)

type AccessBucket struct {
	Name     string    `json:"Name"`
	Alias    string    `json:"Alias"`
	Interval int       `json:"Interval"` //时间间隔
	Unit     int       `json:"Unit"`     //单位时间内输出
	Cap      int       `json:"Cap"`      //桶容量
	reset    time.Time //上次分配时间
	ch       chan struct{}
}

var accessMap sync.Map //桶集合

//初始化运行环境
func Init() error {
	if err := InitConfig(); err != nil {
		return err
	}
	if err := InitList(); err != nil {
		return err
	}
	go loop()
	go SaveList()
	go redisCount()
	return nil
}

//加载配置参数
func InitConfig() error {
	var value int64
	var err error
	conf, err := qconf.LoadConfiguration(configFilePath)
	if err == nil {
		if value, err = conf.GetInteger("defaultInterval"); err == nil {
			defaultInterval = int(value)
		} else {
			return err
		}
		if value, err = conf.GetInteger("defaultUnit"); err == nil {
			defaultUnit = int(value)
		} else {
			return err
		}

		if value, err = conf.GetInteger("defaultCap"); err == nil {
			defaultCap = int(value)
		} else {
			return err
		}

		redisAddress = conf.GetString(redisAddress)
		if redisAddress == "" {
			return errors.New("redis地址为空")
		}
	}
	return err
}

//初始化桶集合配置
func InitList() error {
	var conf []*AccessBucket
	contents, err := ioutil.ReadFile(listFilePath)
	if err != nil {
		return err
	}
	err = json.Unmarshal(contents, &conf)
	if err != nil {
		return err
	}
	for _, v := range conf {
		v.reset = time.Now()
		v.ch = make(chan struct{}, v.Cap)
		accessMap.Store(v.Name, v)
	}
	return nil
}

//定时将桶集合序列化到配置文件中
func SaveList() {
	for {
		time.Sleep(30 * time.Second)
		var buckets = make([]*AccessBucket, 0)
		accessMap.Range(func(key, value interface{}) bool {
			buckets = append(buckets, value.(*AccessBucket))
			return true
		})
		contents, _ := json.Marshal(buckets)
		ioutil.WriteFile(listFilePath, contents, 0777)
	}
}

//从桶集合中获取令牌
func TryTake(name string) bool {
	bucket, ok := accessMap.Load(name)
	if !ok {
		bucket = &AccessBucket{
			Name:     name,
			Interval: defaultInterval,
			Unit:     defaultUnit,
			Cap:      defaultCap,
			reset:    time.Now(),
			ch:       make(chan struct{}, defaultCap),
		}
		bucket, _ = accessMap.LoadOrStore(name, bucket)
	}

	select {
	case <-bucket.(*AccessBucket).ch:
		return true
	case <-time.Tick(10 * time.Millisecond):
		return false
	}
}

//定时添加令牌
func loop() {
	for {
		accessMap.Range(func(key, value interface{}) bool {
			bucket := value.(*AccessBucket)
			if bucket.Unit > 0 && bucket.reset.Add(time.Duration(bucket.Interval)*time.Millisecond).Before(time.Now()) {
				for i := 0; i < bucket.Unit; i++ {
					if len(bucket.ch) < bucket.Cap {
						bucket.ch <- struct{}{}
					} else {
						break
					}
				}
				bucket.reset = time.Now()
			}
			return true
		})
		time.Sleep(1 * time.Second)
	}
}
