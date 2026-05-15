package onlyoffice

type Api struct {
	Config
}

func NewApi(config Config) *Api {
	return &Api{Config: config}
}

// 使用示例（注释，仅示范如何使用 Api.UpdateDocx）
// 示例说明：
// 1. 创建 onlyoffice Api 实例并设置 OSS 客户端。
// 2. 在 Gin 路由中将 incoming 请求转换为 vingo.Context 后调用 api.UpdateDocx。
// 3. 下面示例依赖于项目中已有的 oss.NewApi 和 vingo 的上下文构造方法，实际调用时按项目实现调整。
//
// 示例：
//   cfg := onlyoffice.Config{ /* ... 填充配置 ... */ }
//   api := onlyoffice.NewApi(cfg)
//   // 假设有一个构造 OSS 客户端的函数：ossApi := oss.NewApi(ossCfg)
//   api.OSS = ossApi
//
//   // 在 Gin 中注册路由（示例伪代码）：
//   // g.Post("onlyoffice.docx.update", api.UpdateDocx)
