# novel_retrieval_system
小说检索系统
# 项目说明
- 项目建立目标是为了验证RAG系统在工程上的可行性，作为实际项目的原型，提前探知技术陷阱。

# 运行环境

## 软件环境
- Ubuntu Linux，版本26.04。
- PostgreSQL 数据库，版本18.6。
- Qdrant 向量数据库，版本1.20。
- MinIO 对象存储系统，2026.08.04版本。
- llama.cpp 本地模型基座，版本b10454。
    - Qwen3-Embedding-0.6B-f16，向量模型。
    - Qwen3.6-35B-A3B-Uncensored-HauhauCS-Aggressive-IQ2_M，本地小语言模型。
- Deepseek 外部模型服务。
    - Deepseek V4.1-Flash-1017，外部大语言模型。

## 开发工具
- NodeJS 框架，版本26.13.0。
- Vue3 前端框架，版本3.4.29。
- Element Plus 前端组件库，版本2.14.7。
- Go 语言，版本1.27.1。
- Goland、Trae IDE、Navicat Lite。

## 服务器硬件
- CPU，Intel Core i7-13620H。
- 内存，32GB DDR5（实际只用了16GB内存）。
- 显卡，RTX 4060M 8GB。
- 硬盘，SSD 1TB。

# 项目功能

## 小说内容问答
- 用户输入问题，选择参考书籍，开始问答。系统默认创建会话。
- 用户也可以点击一个已有会话，继续提问。
![31.png](archive/images/31.png)
- 系统回答时，会展示思考过程和答案。
![32.png](archive/images/32.png)
- 回答完毕后，思考过程默认折叠，用户可以点击展开查看。
![33.png](archive/images/33.png)
- 问题答案下方展示参考书籍及章节，可以点击下载原文。
![34.png](archive/images/34.png)
- 用户当前问题无法确定参考资料的，使用之前问题的参考资料。
![35.png](archive/images/35.png)
- 会话内容过多时，系统自动压缩历史对话，形成对话摘要。
![36.png](archive/images/36.png)
- 点击次元统计，可以查看当前对话使用了多少Token。
![37.png](archive/images/37.png)

## 小说文件上传
- 用户选择小说文件上传，系统解析文件内容。目前只支持txt和md格式，后续准备增加pdf和word格式。
![21.png](archive/images/21.png)
- 上传文件后，用户点击启动，系统开始在后台解析文件内容。解析过程是多任务并行的。
![22.png](archive/images/22.png)
- 可以随时查看解析进度，可以随时取消解析任务。
![23.png](archive/images/23.png)
- 已经解析完成的文件，变成就绪状态的小说。解析中的小说不能删除。
![24.png](archive/images/24.png)
- 已经解析完成的小说，可以修改小说的标题，可以删除，可以在问答中使用。
![25.png](archive/images/25.png)

## 登录和用户管理
- 用户使用系统前，需要先登录，默认用户admin，密码password。
![11.png](archive/images/11.png)
- 用户管理界面，可以增加、修改、删除用户。
![12.png](archive/images/12.png)
- 用户管理界面，可以设置用户的权限，更改用户的密码。
![13.png](archive/images/13.png)
- 在用户管理界面，如有权限可以查看每个用户的词元使用统计量。
![14.png](archive/images/14.png)