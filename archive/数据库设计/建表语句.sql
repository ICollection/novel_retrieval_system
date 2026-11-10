-- account[系统用户]
create table account (
  id varchar(40) not null,
  name varchar(10) not null,
  login varchar(20) not null,
  password varchar(250) not null,
  permission varchar(250),
  primary key (id)
);
comment on table account is '系统用户';
comment on column account.id is '用户编码（UUID）';
comment on column account.name is '用户名称';
comment on column account.login is '登录名称';
comment on column account.password is '用户密码';
comment on column account.permission is '用户权限';
alter table account add unique (login);

-- chapter[章节信息]
create table chapter (
  id varchar(40) not null,
  novel varchar(40) not null,
  serial int4 not null,
  title varchar(200) not null,
  start int4 not null,
  final int4 not null,
  character_count int4 not null,
  primary key (id)
);
comment on table chapter is '章节信息';
comment on column chapter.id is '章节编码（UUID）';
comment on column chapter.novel is '小说编码';
comment on column chapter.serial is '章节序号';
comment on column chapter.title is '章节标题';
comment on column chapter.start is '起始位置（字符下标）';
comment on column chapter.final is '结束位置（字符下标）';
comment on column chapter.character_count is '字符总数';

-- consume[词元消费]
create table consume (
  id varchar(40) not null,
  account varchar(40) not null,
  session varchar(40) not null,
  message varchar(40),
  model varchar(5) not null,
  prompt int4 not null,
  answer int4 not null,
  primary key (id)
);
comment on table consume is '词元消费';
comment on column consume.id is '消费编码';
comment on column consume.account is '所属用户';
comment on column consume.session is '所属会话';
comment on column consume.message is '所属消息';
comment on column consume.model is '使用模型';
comment on column consume.prompt is '输入词元';
comment on column consume.answer is '输出词元';

-- message[会话消息]
create table message (
  id varchar(40) not null,
  session varchar(40) not null,
  serial int4 not null,
  human bool not null,
  content text,
  novels text not null,
  segments text,
  complete bool not null,
  create_time timestamptz(6) not null,
  primary key (id)
);
comment on table message is '会话消息';
comment on column message.id is '消息编码（UUID）';
comment on column message.session is '所属会话';
comment on column message.serial is '消息序号（从1开始）';
comment on column message.human is '是否用户';
comment on column message.content is '消息内容';
comment on column message.novels is '引用小说（逗号分隔小说编码）';
comment on column message.segments is '召回切片（逗号分隔切片编码）';
comment on column message.complete is '是否完整';
comment on column message.create_time is '创建时间';
alter table message add unique (session, serial);

-- mission[处理任务]
create table mission (
  id varchar(40) not null,
  novel varchar(40) not null,
  status varchar(2) not null,
  total_count int4 not null,
  embed_count int4 not null,
  launch_time timestamptz(6),
  finish_time timestamptz(6),
  update_time timestamptz(6) not null,
  retry_count int4 not null,
  message varchar(2000),
  primary key (id)
);
comment on table mission is '处理任务';
comment on column mission.id is '任务编码（UUID）';
comment on column mission.novel is '小说编码';
comment on column mission.status is '任务状态（01：待启动，02：已启动，03：解析中，04：切分中，05：向量中，06：已完成，07：已失败，08：已取消）';
comment on column mission.total_count is '分片总数';
comment on column mission.embed_count is '入库数量';
comment on column mission.launch_time is '开始时间';
comment on column mission.finish_time is '完成时间';
comment on column mission.update_time is '更新时间';
comment on column mission.retry_count is '重试次数';
comment on column mission.message is '失败消息';
alter table mission add unique (novel);

-- novel[参考小说]
create table novel (
  id varchar(40) not null,
  origin varchar(40) not null,
  title varchar(200) not null,
  character_count int4 not null,
  chapter_count int4 not null,
  create_user varchar(40) not null,
  create_time timestamptz(6) not null,
  update_time timestamptz(6) not null,
  primary key (id)
);
comment on table novel is '参考小说';
comment on column novel.id is '小说编码（UUID）';
comment on column novel.origin is '原始编码';
comment on column novel.title is '小说名称';
comment on column novel.character_count is '字符总数';
comment on column novel.chapter_count is '章节总数';
comment on column novel.create_user is '创建用户编码';
comment on column novel.create_time is '创建时间';
comment on column novel.update_time is '更新时间';
alter table novel add unique (origin);

-- origin[原始文件]
create table origin (
  id varchar(40) not null,
  name varchar(200) not null,
  type varchar(10) not null,
  size int4 not null,
  hash varchar(80) not null,
  create_user varchar(40) not null,
  create_time timestamptz(6) not null,
  primary key (id)
);
comment on table origin is '原始文件';
comment on column origin.id is '文件编码（UUID）';
comment on column origin.name is '文件名称';
comment on column origin.type is '文件类型';
comment on column origin.size is '文件大小';
comment on column origin.hash is '文件摘要（SHA-1）';
comment on column origin.create_user is '创建用户编码';
comment on column origin.create_time is '创建时间';
alter table origin add unique (hash);

-- segment[分片信息]
create table segment (
  id varchar(40) not null,
  novel varchar(40) not null,
  chapter varchar(40) not null,
  serial int4 not null,
  point1 int4 not null,
  prefix text,
  point2 int4 not null,
  middle text not null,
  point3 int4 not null,
  suffix text,
  point4 int4 not null,
  primary key (id)
);
comment on table segment is '分片信息';
comment on column segment.id is '分片编码（UUID）';
comment on column segment.novel is '小说编码';
comment on column segment.chapter is '章节编码';
comment on column segment.serial is '分片序号';
comment on column segment.point1 is '一分位点';
comment on column segment.prefix is '分片前缀';
comment on column segment.point2 is '二分位点';
comment on column segment.middle is '主要内容';
comment on column segment.point3 is '三分位点';
comment on column segment.suffix is '分片后缀';
comment on column segment.point4 is '四分位点';
alter table segment add unique (novel, serial);
create index on segment (novel, chapter, serial);

-- session[用户会话]
create table session (
  id varchar(40) not null,
  title varchar(100) not null,
  summary text,
  contain int4,
  current int4 not null,
  create_user varchar(40) not null,
  create_time timestamptz(6) not null,
  update_time timestamptz(6) not null,
  primary key (id)
);
comment on table session is '用户会话';
comment on column session.id is '会话编码（UUID）';
comment on column session.title is '会话标题';
comment on column session.summary is '前情摘要';
comment on column session.contain is '截至序号';
comment on column session.current is '当前序号';
comment on column session.create_user is '创建用户';
comment on column session.create_time is '创建时间';
comment on column session.update_time is '更新时间';

-- 创建外键
alter table chapter add foreign key (novel) references novel (id);
alter table consume add foreign key (account) references account (id);
alter table message add foreign key (session) references session (id);
alter table mission add foreign key (novel) references novel (id);
alter table novel add foreign key (origin) references origin (id);
alter table novel add foreign key (create_user) references account (id);
alter table origin add foreign key (create_user) references account (id);
alter table segment add foreign key (novel) references novel (id);
alter table segment add foreign key (chapter) references chapter (id);
alter table session add foreign key (create_user) references account (id);

-- 插入管理员用户（密码password）
insert into account (id, name, login, password, permission) values ('00000000-0000-0000-0000-000000000000', '管理员', 'admin', '$2a$10$vq4yO1HR/Ufey4.HrmN3R.TfE3CzCHogQx1E918/2R0cUUk3oMp42', 'ManageUser,ManageOrigin,ManageMission');
