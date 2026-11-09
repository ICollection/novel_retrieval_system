package pipeline

import (
	"context"
	"log"
	"service/internal/common"
	"service/internal/config"
	"service/internal/storage"
	"sync"
	"time"
)

// schedule 启动定时器。
//
// param: ctx 上下文。
func schedule(ctx context.Context) {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			assign(ctx)
			break
		}
	}
}

// assign 查询可用任务。
//
// param: ctx 上下文。
func assign(ctx context.Context) {
	missions, err := storage.RDB.QueryMatrix[mission](ctx, "select mission.id as mission,novel.id as novel,origin.id as origin,mission.status,origin.type,total_count,embed_count,retry_count from mission inner join novel on novel.id=mission.novel inner join origin on origin.id=novel.origin where mission.status in ($1,$2,$3,$4) order by mission.status desc,mission.launch_time;", common.MissionStatus02, common.MissionStatus03, common.MissionStatus04, common.MissionStatus05)
	if err != nil {
		log.Println("查询可用任务失败，详细错误信息：", err)
		return
	}
	for _, mission := range missions {
		if task := request(ctx, *mission.Mission); task != nil {
			wait.Go(func() { handle(task, &mission) })
		}
	}
	missions, err = storage.RDB.QueryMatrix[mission](ctx, "select mission.id as mission,novel.id as novel,origin.id as origin,mission.status,origin.type,total_count,embed_count,retry_count from mission inner join novel on novel.id=mission.novel inner join origin on origin.id=novel.origin where mission.status=$1 order by mission.status desc,mission.launch_time;", common.MissionStatus08)
	if err != nil {
		log.Println("查询取消任务失败，详细错误信息：", err)
		return
	}
	mutex.Lock()
	defer mutex.Unlock()
	for _, mission := range missions {
		if cancel, exists := index[*mission.Mission]; exists {
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
//
// param: parent 父级上下文。
// param: mission 任务编码。
// return: 任务上下文，并发已满或任务已在运行时返回 nil。
func request(parent context.Context, mission string) context.Context {
	mutex.Lock()
	defer mutex.Unlock()
	if len(index) >= config.Cfg.Station.Concurrency {
		return nil
	}
	if _, exists := index[mission]; exists == true {
		return nil
	}
	ctx, cancel := context.WithCancel(parent)
	index[mission] = cancel
	return ctx
}

// release 记录释放任务。
//
// param: mission 任务编码。
func release(mission string) {
	mutex.Lock()
	defer mutex.Unlock()
	delete(index, mission)
}

// handle 处理单个任务。
//
// param: ctx 上下文。
// param: mission 任务信息。
func handle(ctx context.Context, mission *mission) {
	defer release(*mission.Mission)
	status := *mission.Status
	// 开始执行操作。
	if status == common.MissionStatus02 {
		// 验证取消状态。
		if ctx.Err() != nil {
			return
		}
		// 更新任务状态。
		err := storage.RDB.QueryNaught(ctx, "update mission set status=$1,update_time=$2 where id=$3", common.MissionStatus03, time.Now().UTC(), mission.Mission)
		if err != nil {
			log.Println("更新任务状态失败，详细错误信息：", err.Error())
			return
		}
		// 记录成功日志。
		log.Println("开始解析文件成功，原始文件：", *mission.Origin+"."+*mission.Type)
		status = common.MissionStatus03
	}
	// 执行归一操作。
	if status == common.MissionStatus03 {
		// 验证取消状态。
		if ctx.Err() != nil {
			return
		}
		// 执行归一操作。
		err := normalize(ctx, *mission.Origin, *mission.Type)
		if err != nil {
			log.Println("执行归一操作失败，详细错误信息：", err.Error())
			return
		}
		// 更新任务状态。
		err = storage.RDB.QueryNaught(ctx, "update mission set status=$1,update_time=$2 where id=$3", common.MissionStatus04, time.Now().UTC(), mission.Mission)
		if err != nil {
			log.Println("更新任务状态失败，详细错误信息：", err.Error())
			return
		}
		// 记录成功日志。
		log.Println("执行归一操作成功，归一文件：", *mission.Origin+".md")
		status = common.MissionStatus04
	}
	// 执行分章切片。
	if status == common.MissionStatus04 {
		// 验证取消状态。
		if ctx.Err() != nil {
			return
		}
		// 执行分章切片。
		total_count, character_count, chapter_count, err := organize(ctx, *mission.Origin, *mission.Novel)
		if err != nil {
			log.Println("执行分章切片失败，详细错误信息：", err.Error())
			return
		}
		// 更新任务状态。
		var batch storage.Batch
		batch.Append("update mission set status=$1,total_count=$2,embed_count=0,update_time=$3 where id=$4", common.MissionStatus05, total_count, time.Now().UTC(), mission.Mission)
		batch.Append("update novel set character_count=$1,chapter_count=$2,update_time=$3 where id=$4", character_count, chapter_count, time.Now().UTC(), mission.Novel)
		err = storage.RDB.BatchNaught(ctx, &batch)
		if err != nil {
			log.Println("更新任务状态失败，详细错误信息：", err.Error())
			return
		}
		// 记录成功日志。
		log.Println("执行分章切片成功，切片数量：", total_count)
		status = common.MissionStatus05
	}
	// 执行向量入库。
	if status == common.MissionStatus05 {
		// 验证取消状态。
		if ctx.Err() != nil {
			return
		}
		// 执行向量入库。
		embed_count, err := vectorize(ctx, *mission.Mission, *mission.Novel)
		if err != nil {
			log.Println("执行向量入库失败，详细错误信息：", err.Error())
			return
		}
		// 更新任务状态。
		err = storage.RDB.QueryNaught(ctx, "update mission set status=$1,embed_count=$2,finish_time=$3,update_time=$3 where id=$4", common.MissionStatus06, embed_count, time.Now().UTC(), mission.Mission)
		if err != nil {
			log.Println("更新任务状态失败，详细错误信息：", err.Error())
			return
		}
		// 记录成功日志。
		log.Println("执行向量入库成功，入库数量：", embed_count)
		status = common.MissionStatus06
	}
}
