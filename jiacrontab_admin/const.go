package admin

const (
	event_DelNodeDesc = "删除节点"
	event_RenameNode  = "更新节点名"

	event_EditCronJob  = "编辑计划任务"
	event_DelCronJob   = "删除计划任务"
	event_StopCronJob  = "停止计划任务"
	event_StartCronJob = "启动计划任务"
	event_ExecCronJob  = "执行了计划任务"
	event_KillCronJob  = "强杀计划任务进程"

	event_EditDaemonJob  = "编辑常驻任务"
	event_DelDaemonJob   = "删除常驻任务"
	event_StartDaemonJob = "启动常驻任务"
	event_StopDaemonJob  = "停止常驻任务"

	event_EditGroup = "编辑分组"
	event_GroupNode = "设置节点分组"

	event_SignUpUser = "创建新用户"

	event_AuditCrontabJob = "审核计划任务"
	event_AuditDaemonJob  = "审核常驻任务"
)
