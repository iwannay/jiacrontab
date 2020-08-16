package admin

const (
	event_DelNodeDesc = "{username}删除了节点{targetName}"
	event_RenameNode  = "{username}将{sourceName}重命名为{targetName}"

	event_EditCronJob  = "{sourceName}{username}编辑了定时任务{targetName}"
	event_DelCronJob   = "{sourceName}{username}删除了定时任务{targetName}"
	event_StopCronJob  = "{sourceName}{username}停止了定时任务{targetName}"
	event_StartCronJob = "{sourceName}{username}启动了定时任务{targetName}"
	event_ExecCronJob  = "{sourceName}{username}执行了定时任务{targetName}"
	event_KillCronJob  = "{sourceName}{username}kill了定时任务进程{targetName}"

	event_EditDaemonJob  = "{sourceName}{username}编辑了常驻任务{targetName}"
	event_DelDaemonJob   = "{sourceName}{username}删除了常驻任务{targetName}"
	event_StartDaemonJob = "{sourceName}{username}启动了常驻任务{targetName}"
	event_StopDaemonJob  = "{sourceName}{username}停止了常驻任务{targetName}"

	event_EditGroup = "{username}编辑了{targetName}组"
	event_GroupNode = "{username}将节点{sourceName}添加到{targetName}组"

	event_SignUpUser      = "{username}创建了用户{targetName}"
	event_EditUser        = "{username}更新了用户信息"
	event_DeleteUser      = "{username}删除了用户{targetName}"
	event_GroupUser       = "{username}将用户{sourceUsername}设置为{targetName}组"
	event_AuditCrontabJob = "{sourceName}{username}审核了定时任务{targetName}"
	event_AuditDaemonJob  = "{sourceName}{username}审核了常驻任务{targetName}"

	event_ClearHistory = "{username}清除了{targetName}前的任务执行日志"
)
