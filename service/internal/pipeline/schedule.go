package pipeline

import (
	"context"
	"log"
	"service/internal/config"
	"service/internal/storage"
	"sync"
	"time"
)

// schedule 启动定时器。
func schedule(ctx context.Context) {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			for _, cancel := range index {
				cancel()
			}
			return
		case <-ticker.C:
			assign()
			break
		}
	}
}

// assign 查询可用任务。
func assign() {
	missions, err := storage.RDB.QueryMatrix[mission](context.Background(), "select mission.id as mission,novel.id as novel,origin.id as origin,mission.status,origin.type,total_count,embed_count,retry_count from mission inner join novel on novel.id=mission.novel inner join origin on origin.id=novel.origin where mission.status in ('02','03','04','05') order by mission.status desc,mission.launch_time;")
	if err != nil {
		log.Println("查询可用任务失败，详细错误信息：", err)
		return
	}
	for _, mission := range missions {
		if ctx := request(*mission.Mission); ctx != nil {
			go handle(ctx, &mission)
		}
	}
	ids, err := storage.RDB.QueryMatrix[string](context.Background(), "select id from mission where status='08';")
	if err != nil {
		log.Println("查询取消任务失败，详细错误信息：", err)
		return
	}
	for _, id := range ids {
		if cancel, exists := index[id]; exists {
			cancel()
		}
	}
}

// mission 任务信息。
type mission struct {
	Mission *string `db:"mission"`
	Novel   *string `db:"novel"`
	Origin  *string `db:"origin"`
	Status  *string `db:"status"`
	Type    *string `db:"type"`
	Total   *int32  `db:"total_count"`
	Embed   *int32  `db:"embed_count"`
	Retry   *int32  `db:"retry_count"`
}

var (
	// mutex 分配用排他锁。
	mutex sync.Mutex
	// index 活动任务索引。
	index = make(map[string]context.CancelFunc)
)

// request 申请启动任务。
func request(mission string) context.Context {
	mutex.Lock()
	defer mutex.Unlock()
	if len(index) >= config.Config.Station.Concurrency {
		return nil
	}
	if _, exists := index[mission]; exists == true {
		return nil
	}
	ctx, cancel := context.WithCancel(context.Background())
	index[mission] = cancel
	return ctx
}

// release 记录释放任务。
func release(mission string) {
	mutex.Lock()
	defer mutex.Unlock()
	delete(index, mission)
}

// handle 处理单个任务。
func handle(ctx context.Context, mission *mission) {
	defer release(*mission.Mission)
}
