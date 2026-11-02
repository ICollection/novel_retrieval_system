package deposit

import (
	"context"
	"log"

	"github.com/qdrant/go-client/qdrant"
)

// init 初始化集合。
func init() {
	client, err := qdrant.NewClient(&qdrant.Config{Host: "192.168.1.100", Port: 6334})
	if err != nil {
		log.Fatal("创建连接失败，详细错误信息：", err)
	}
	defer client.Close()
	ctx := context.Background()
	if exists, err := client.CollectionExists(ctx, "novel"); err != nil {
		log.Fatal("创建连接失败，详细错误信息：", err)
	} else if exists == false {
		request := qdrant.CreateCollection{
			CollectionName: "novel",
			VectorsConfig: qdrant.NewVectorsConfig(&qdrant.VectorParams{
				Size:     1024,
				Distance: qdrant.Distance_Cosine,
				Datatype: qdrant.Datatype_Float16.Enum(),
			}),
		}
		if err = client.CreateCollection(ctx, &request); err != nil {
			log.Fatal("创建集合失败，详细错误信息：", err)
		}
	}
}
